package handlers

import (
	"net/http"
	"time"
	"vacation-scheduler/backend/database"
	"vacation-scheduler/backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// TODO: Move this to a secure configuration/environment variable
var jwtKey = []byte("your_very_secret_key_change_this") // CHANGE THIS!

// RegisterInput defines the expected input for user registration
type RegisterInput struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"` // Require minimum password length
}

// LoginInput defines the expected input for user login
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Claims defines the structure of the JWT claims
type Claims struct {
	UserID  uint   `json:"user_id"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// RegisterUser handles new user registration
func RegisterUser(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := database.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create the user (default IsAdmin is false)
	user := models.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Password:  string(hashedPassword),
		IsAdmin:   false, // Default to non-admin
		// Set default VacationLimit or get from config/input if needed
		VacationLimit: 28,
	}

	// Save user to database
	if result := database.DB.Create(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + result.Error.Error()})
		return
	}

	// Do not return password hash
	user.Password = ""

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user": user})
}

// LoginUser handles user login and JWT generation
func LoginUser(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	var user models.User
	// Find user by email
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"}) // Generic error
		return
	}

	// Compare submitted password with stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"}) // Generic error
		return
	}

	// Generate JWT token
	expirationTime := time.Now().Add(24 * time.Hour) // Token valid for 24 hours
	claims := &Claims{
		UserID:  user.ID,
		IsAdmin: user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     tokenString,
		"user_id":   user.ID,
		"is_admin":  user.IsAdmin,
		"expiresAt": expirationTime.Format(time.RFC3339),
	})
}
