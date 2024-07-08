package database

import (
	"go_services/models"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		log.Fatal("POSTGRES_URL is not set!")
	}
	var err error
	// Open the database connection
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// Migrate the schema
	if err := DB.AutoMigrate(&models.Task{}, &models.Company{}); err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}
}
