package models

import "time"

// VacationStatus represents the status of a vacation request
type VacationStatus string

const (
	StatusPending   VacationStatus = "pending"
	StatusApproved  VacationStatus = "approved"
	StatusRejected  VacationStatus = "rejected"
	StatusSubmitted VacationStatus = "submitted" // Submitted by employee, pending manager approval
)

    // VacationRequest represents a single vacation request from a user
    type VacationRequest struct {
    	ID        uint           `json:"id" gorm:"primaryKey;type:int unsigned"` // Explicitly set type
    	UserID    uint           `json:"user_id" gorm:"type:int unsigned"` // Explicitly set type for FK
    	User      User           `json:"user" gorm:"foreignKey:UserID"` // Belongs to User
    	StartDate time.Time      `json:"start_date"`
	EndDate   time.Time      `json:"end_date"`
	Status    VacationStatus `json:"status" gorm:"default:'pending'"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	// Optional: Add comments field if needed
	// Comment string `json:"comment,omitempty"`
}

// VacationPeriod represents a single continuous part of a vacation
type VacationPeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}
