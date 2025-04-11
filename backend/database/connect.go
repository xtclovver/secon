package database

import (
	"fmt"
	"log"
	"os" // Import os package to read environment variables

	"vacation-scheduler/backend/models" // Import models package

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDatabase initializes the database connection and runs migrations
func ConnectDatabase() {
	// Read database connection details from environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	// Provide default values if environment variables are not set (optional, good for local dev without Docker)
	if dbUser == "" {
		dbUser = "root" // Default user from docker-compose
	}
	if dbPassword == "" {
		dbPassword = "v?jKm}J7R8(X/+xZ" // Default password from docker-compose
	}
	if dbName == "" {
		dbName = "vacation_db" // Default db name from docker-compose
	}
	if dbHost == "" {
		dbHost = "34.88.50.168" // Default to localhost if not running in Docker Compose network
	}
	if dbPort == "" {
		dbPort = "3306" // Default MySQL port
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
		os.Exit(1) // Exit if connection fails
	}

	fmt.Println("Database connection successfully opened.")

	// AutoMigrate the schema
	fmt.Println("Running database migrations...")
	err = DB.AutoMigrate(&models.User{}, &models.VacationRequest{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
		os.Exit(1) // Exit if migration fails
	}
	fmt.Println("Database migrated successfully.")
}
