package database

import (
	"job-scheduler/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is the global database connection
var DB *gorm.DB

// Init initializes the database connection and migrates the schema
func Init() {
	// Connection string for PostgreSQL
	dsn := "host=localhost user=youruser password=yourpassword dbname=yourdbname port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	var err error
	// Open the database connection
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// Migrate the schema
	if err := DB.AutoMigrate(&models.Task{}); err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}
}
