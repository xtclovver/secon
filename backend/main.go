package main

import (
	"fmt"
	"log"
	"vacation-scheduler/backend/database"   // Import database package
	"vacation-scheduler/backend/handlers"   // Import handlers package
	"vacation-scheduler/backend/middleware" // Import middleware package

	"github.com/gin-contrib/cors" // Import CORS middleware
	"github.com/gin-gonic/gin"    // Import Gin
)

func main() {
	// Connect to the database
	database.ConnectDatabase()

	// Initialize Gin router
	router := gin.Default()

	// --- Middleware ---
	// Add CORS middleware - configure origins as needed for production
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"} // Allow frontend origin
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(config))

	// --- API Routes ---
	api := router.Group("/api")
	{
		// --- Public User Routes (Auth) ---
		userRoutes := api.Group("/users")
		{
			userRoutes.POST("/register", handlers.RegisterUser)
			userRoutes.POST("/login", handlers.LoginUser)
			// TODO: Add routes for password reset, email verification if needed
		}

		// --- Authenticated Routes ---
		// All routes in this group require a valid JWT token
		authenticated := api.Group("/")
		authenticated.Use(middleware.AuthMiddleware()) // Apply auth middleware to this group
		{
			// --- Vacation Routes (require authentication) ---
			vacationRoutes := authenticated.Group("/vacations")
			{
				// Routes accessible to all authenticated users
				vacationRoutes.POST("/", handlers.CreateVacationRequest)    // Create a new request for the logged-in user
				vacationRoutes.GET("/my", handlers.GetUserVacationRequests) // Get requests for the logged-in user (adjust handler needed)

				// Routes requiring Admin privileges
				adminVacationRoutes := vacationRoutes.Group("/")
				adminVacationRoutes.Use(middleware.AdminMiddleware()) // Apply admin middleware
				{
					adminVacationRoutes.GET("/", handlers.GetVacationRequests)                 // Get all requests (admin view)
					adminVacationRoutes.PUT("/:id", handlers.UpdateVacationRequest)            // Update any request (status, dates)
					adminVacationRoutes.DELETE("/:id", handlers.DeleteVacationRequest)         // Delete any request
					adminVacationRoutes.GET("/overlaps", handlers.CheckOverlappingVacations)   // Check overlaps (admin view)
					adminVacationRoutes.GET("/user/:userId", handlers.GetUserVacationRequests) // Get requests for a specific user (admin view)
					// TODO: Add route for submitting schedule (admin?)
				}
			}

			// --- Other Authenticated Routes (e.g., User Profile) ---
			// profileRoutes := authenticated.Group("/profile")
			// {
			//  profileRoutes.GET("/", handlers.GetUserProfile)
			//  profileRoutes.PUT("/", handlers.UpdateUserProfile)
			// }
		}
	}

	// Simple health check route (public)
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Vacation Scheduler Backend is running"})
	})

	// Start the server
	port := "8080" // Or get from environment variable
	fmt.Printf("Starting server on port %s\n", port)
	log.Fatal(router.Run(":" + port))
}
