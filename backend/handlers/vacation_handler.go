package handlers

import (
	"net/http"
	"strconv"
	"time"
	"vacation-scheduler/backend/database"
	"vacation-scheduler/backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm" // Import gorm package
)

// calculateDuration calculates the duration between two dates in days (inclusive)
func calculateDuration(start, end time.Time) int {
	// Ensure dates are treated as whole days in UTC for consistent calculation
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	end = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)
	// Add one day because the end date is inclusive
	return int(end.Sub(start).Hours()/24) + 1
}

// Helper function to get user info from context
func getUserInfoFromContext(c *gin.Context) (userID uint, isAdmin bool, ok bool) {
	idVal, idExists := c.Get("userID")
	adminVal, adminExists := c.Get("isAdmin")

	if !idExists || !adminExists {
		return 0, false, false
	}

	userID, idOk := idVal.(uint)
	isAdmin, adminOk := adminVal.(bool)

	if !idOk || !adminOk {
		// Log error or handle type assertion failure
		return 0, false, false
	}
	return userID, isAdmin, true
}


// CreateVacationRequest handles the creation of a new vacation request for the logged-in user
func CreateVacationRequest(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	requestingUserID, _, ok := getUserInfoFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information from context"})
		return
	}

	var input struct {
		// UserID is no longer needed in input, taken from token
		StartDate string `json:"start_date" binding:"required"` // Expecting "YYYY-MM-DD"
		EndDate   string `json:"end_date" binding:"required"`   // Expecting "YYYY-MM-DD"
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Parse dates
	layout := "2006-01-02" // Go's reference date format for YYYY-MM-DD
	startDate, errStart := time.Parse(layout, input.StartDate)
	endDate, errEnd := time.Parse(layout, input.EndDate)

	if errStart != nil || errEnd != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD."})
		return
	}

	// Basic validation: end date must be after start date
	if !endDate.After(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "End date must be after start date"})
		return
	}

	// --- Validation ---
	// 1. Fetch User (the one making the request) to get vacation limit
	var user models.User
	if err := database.DB.First(&user, requestingUserID).Error; err != nil {
		// User should exist if token is valid, but check anyway
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authenticated user not found in database"}) // Should not happen
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user data: " + err.Error()})
		}
		return
	}

	// 2. Calculate duration of the new request
	newRequestDuration := calculateDuration(startDate, endDate)

	// 3. Fetch existing approved/pending requests for the year to calculate used days
	// Assuming vacations are planned for the year of the start date
	currentYear := startDate.Year()
	var existingRequests []models.VacationRequest
	// Fetch for the requesting user
	if err := database.DB.Where("user_id = ? AND status IN (?, ?) AND strftime('%Y', start_date) = ?",
		requestingUserID, models.StatusApproved, models.StatusPending, strconv.Itoa(currentYear)).
		Find(&existingRequests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing vacations: " + err.Error()})
		return
	}

	// 4. Calculate total days already used/planned + new request duration
	totalPlannedDays := newRequestDuration
	has14DayPart := newRequestDuration >= 14 // Check if the new request satisfies the 14-day rule

	for _, req := range existingRequests {
		duration := calculateDuration(req.StartDate, req.EndDate)
		totalPlannedDays += duration
		if duration >= 14 {
			has14DayPart = true // Found an existing part satisfying the rule
		}
	}

	// 5. Check against user's limit
	if totalPlannedDays > user.VacationLimit {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":            "Vacation request exceeds the annual limit.",
			"limit":            user.VacationLimit,
			"requested_days":   newRequestDuration,
			"total_planned":    totalPlannedDays,
			"remaining_budget": user.VacationLimit - (totalPlannedDays - newRequestDuration),
		})
		return
	}

	// 6. Check 14-day rule (only if this is the *last* part being added within the limit)
	// This logic assumes the user adds parts sequentially. A more robust check might be needed
	// if users can edit/delete requests freely. We check if *any* part (new or existing) is >= 14 days.
	if !has14DayPart {
		// Check if adding this request *completes* the vacation planning up to the limit
		// If the total planned days *after* adding this request equals the limit,
		// and no part is >= 14 days, then it's an error.
		// A simpler approach: always require at least one part >= 14 days among all requests for the year.
		// Let's re-evaluate the 14-day rule check based on all requests for the year.
		allRequestsForYear := append(existingRequests, models.VacationRequest{StartDate: startDate, EndDate: endDate}) // Include the new one
		found14DayPart := false
		for _, req := range allRequestsForYear {
			if calculateDuration(req.StartDate, req.EndDate) >= 14 {
				found14DayPart = true
				break
			}
		}
		if !found14DayPart {
			// This check might be too strict if applied *during* planning.
			// Maybe only enforce when submitting the final schedule?
			// For now, let's enforce it on creation:
			c.JSON(http.StatusBadRequest, gin.H{"error": "At least one part of the vacation must be 14 days or longer."})
			return
		}

	}

	// TODO: Check for overlapping vacations for the same user (optional, depends on requirements)

	// --- Create Request ---
	request := models.VacationRequest{
		UserID:    requestingUserID, // Use ID from token
		StartDate: startDate,
		EndDate:   endDate,
		Status:    models.StatusPending, // Initial status
	}

	if err := database.DB.Create(&request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vacation request: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, request)
}

// GetVacationRequests handles fetching all vacation requests (potentially filtered)
func GetVacationRequests(c *gin.Context) {
	var requests []models.VacationRequest
	// TODO: Add filtering options (by user, status, date range) via query parameters
	// Example: database.DB.Preload("User").Find(&requests) to include user details
	if err := database.DB.Preload("User").Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vacation requests: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, requests)
}

// GetUserVacationRequests handles fetching vacation requests.
// If called via /my, it fetches for the logged-in user.
// If called via /user/:userId (admin only), it fetches for the specified user.
func GetUserVacationRequests(c *gin.Context) {
	var targetUserID uint

	// Check if userId parameter exists (for admin route)
	userIDStr := c.Param("userId")
	if userIDStr != "" {
		// Admin route: Get target user ID from parameter
		userIDParsed, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID parameter"})
			return
		}
		targetUserID = uint(userIDParsed)
		// Authorization is handled by AdminMiddleware in main.go for this route
	} else {
		// /my route: Get user ID from context
		loggedInUserID, _, ok := getUserInfoFromContext(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information from context"})
			return
		}
		targetUserID = loggedInUserID
	}


	var requests []models.VacationRequest
	// Fetch requests for the target user ID
	if err := database.DB.Where("user_id = ?", targetUserID).Preload("User").Order("start_date asc").Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user vacation requests: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, requests)
}

// UpdateVacationRequest handles updating a vacation request (e.g., status change by manager)
func UpdateVacationRequest(c *gin.Context) {
	requestIDStr := c.Param("id")
	requestID, err := strconv.ParseUint(requestIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	var input struct {
		Status    *models.VacationStatus `json:"status"` // Use pointer to detect if status was provided
		StartDate *string                `json:"start_date"`
		EndDate   *string                `json:"end_date"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	var request models.VacationRequest
	if err := database.DB.First(&request, uint(requestID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vacation request not found"})
		return
	}

	// Authorization is handled by AdminMiddleware in main.go

	updated := false
	if input.Status != nil {
		// Validate status transition if needed
		request.Status = *input.Status
		updated = true
	}

	// Handle date updates if provided
	layout := "2006-01-02"
	if input.StartDate != nil {
		startDate, err := time.Parse(layout, *input.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
		request.StartDate = startDate
		updated = true
	}
	if input.EndDate != nil {
		endDate, err := time.Parse(layout, *input.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
		request.EndDate = endDate
		updated = true
	}

	// Re-validate dates if they were updated
	if updated && !request.EndDate.After(request.StartDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "End date must be after start date"})
		return
	}

	// TODO: Re-run validation rules (14-day, limit, overlap) if dates changed

	if updated {
		if err := database.DB.Save(&request).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vacation request: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, request)
}

// DeleteVacationRequest handles deleting a vacation request
func DeleteVacationRequest(c *gin.Context) {
	requestIDStr := c.Param("id")
	requestID, err := strconv.ParseUint(requestIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	// Authorization is handled by AdminMiddleware in main.go
	// TODO: Consider allowing users to delete their own PENDING requests via a different route/logic

	if err := database.DB.Delete(&models.VacationRequest{}, uint(requestID)).Error; err != nil {
		// Check if the error is because the record was not found or another DB error
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Vacation request not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vacation request: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vacation request deleted successfully"})
}

// OverlapInfo details an overlap between two vacation requests
type OverlapInfo struct {
	User1    models.User           `json:"user1"`
	Request1 models.VacationRequest `json:"request1"`
	User2    models.User           `json:"user2"`
	Request2 models.VacationRequest `json:"request2"`
}

// CheckOverlappingVacations finds and returns overlapping vacation periods among approved requests
func CheckOverlappingVacations(c *gin.Context) {
	// Fetch all approved vacation requests (consider adding date range filters)
	var approvedRequests []models.VacationRequest
	if err := database.DB.Where("status = ?", models.StatusApproved).Preload("User").Order("start_date asc").Find(&approvedRequests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch approved vacations: " + err.Error()})
		return
	}

	var overlaps []OverlapInfo

	// Compare each request with every subsequent request
	for i := 0; i < len(approvedRequests); i++ {
		for j := i + 1; j < len(approvedRequests); j++ {
			req1 := approvedRequests[i]
			req2 := approvedRequests[j]

			// Skip if requests belong to the same user
			if req1.UserID == req2.UserID {
				continue
			}

			// Check for overlap: (StartA <= EndB) and (EndA >= StartB)
			if (req1.StartDate.Before(req2.EndDate) || req1.StartDate.Equal(req2.EndDate)) &&
				(req1.EndDate.After(req2.StartDate) || req1.EndDate.Equal(req2.StartDate)) {
				overlaps = append(overlaps, OverlapInfo{
					User1:    req1.User,
					Request1: req1,
					User2:    req2.User,
					Request2: req2,
				})
			}
		}
	}

	// TODO: Implement notification logic for the manager here or in a separate service

	if len(overlaps) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":  "Overlapping vacations found.",
			"overlaps": overlaps,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "No overlapping vacations found."})
	}
}

// TODO: Add handler for submitting the schedule to the manager
