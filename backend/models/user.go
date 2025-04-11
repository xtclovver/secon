package models

    // User represents an employee in the system
    type User struct {
    	ID        uint   `json:"id" gorm:"primaryKey;type:int unsigned"` // Explicitly set type for GORM
    	FirstName string `json:"first_name"`
    	LastName  string `json:"last_name"`
	Email     string `json:"email" gorm:"unique"`
	Password  string `json:"-"` // Password hash - not exposed in JSON
	IsAdmin   bool   `json:"is_admin"` // Role: true for admin, false for regular user
	// Add other relevant fields like DepartmentID, Position, etc.
	VacationLimit int `json:"vacation_limit"` // Total allowed vacation days per year
}
