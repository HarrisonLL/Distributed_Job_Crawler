package database

import (
	"go_services/config"
	"go_services/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	dsn, getURLErr := config.GetURL("POSTGRES_URL")
	if getURLErr != "" {
		log.Fatal(getURLErr)
	}
	var err error
	// Open the database connection
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// Migrate the schema
	if err := DB.AutoMigrate(&models.Task{}, &models.JobType{}, &models.User{}); err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}
}

func InitUserSvc() {
	// Connect with the user_manager role that only has access to the user table
	userDsn, getUserURLErr := config.GetURL("POSTGRES_USER_MANAGER_URL")
	if getUserURLErr != "" {
		log.Fatal(getUserURLErr)
	}

	var err error
	// Open the database connection with limited privileges
	DB, err = gorm.Open(postgres.Open(userDsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database as user_manager: ", err)
	}

}
