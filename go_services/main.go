package main

import (
	"fmt"
	"go_services/database"
	"go_services/handlers"
	"go_services/models"
	"go_services/utils"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func startWeb() {
	router := gin.Default()
	router.GET("/api/v1/tasks", handlers.GetTasks)
	router.POST("/api/v1/tasks", handlers.CreateTask)
	router.PATCH("/api/v1/tasks/:task_id", handlers.UpdateTask)
	router.GET("/api/v1/tasks/:task_id", handlers.GetTaskByID)
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start web server: %v", err)

	}
}

func startScheduler() {
	for {
		var companies []models.Company
		if err := database.DB.Find(&companies).Error; err != nil {
			log.Fatalf("Failed to fetch companies: %v", err)
		}

		for _, company := range companies {
			imageID, err := utils.PullDockerImage(company.DockerImageName)
			if err != nil {
				log.Printf("Failed to pull Docker image for company %s: %v", company.CompanyName, err)
				continue
			}

			// Update the company record with the new image ID and pull date
			if company.DockerImageID != imageID {
				company.DockerImageID = imageID
				company.PullDate = time.Now().Format(time.RFC3339)
				if err := database.DB.Save(&company).Error; err != nil {
					log.Printf("Failed to update Docker image for company %s: %v", company.CompanyName, err)
					continue
				}
			}

			taskID := uuid.New().String()
			envVars := []string{
				fmt.Sprintf("TASKID=%s", taskID),
				fmt.Sprintf("MONGOURL=%s", os.Getenv("MONGOURL")),
			}
			htmlPath := os.Getenv("HTML_PATH")
			volumeMappings := []string{
				htmlPath + ":/app/html_data",
			}
			cmd := []string{
				"--job_type", "software engineer",
				"--location", "USA",
				"--company", company.CompanyName,
			}

			// Start the crawler work
			go func(company models.Company) {
				containerID, err := utils.RunDockerContainer(company.DockerImageName, envVars, volumeMappings, cmd)
				if err != nil {
					log.Printf("Failed to start crawler for company %s: %v", company.CompanyName, err)
				} else {
					log.Printf("Started container %s for company %s", containerID, company.CompanyName)
					// Update task DB
					newTask := models.Task{
						TaskID:      taskID,
						ContainerID: containerID,
						DateTime:    time.Now().Format(time.RFC3339),
						Args: models.JSONMap{
							"job_type": "software engineer",
							"location": "USA",
							"company":  company.CompanyName,
						},
						Status:         models.Started,
						SuccessJobIDs:  []string{},
						FailedJobIDs:   []string{},
						CompletionRate: 0.0,
						IsRetryTask:    false,
						ParentTaskID:   "",
					}
					if err := database.DB.Create(&task).Error; err != nil {
						log.Printf("Failed to create task for company %s: %v", company.CompanyName, err)
					}
				}

			}(company)
		}

		// Schedule every 6 hours
		time.Sleep(6 * time.Hour)
	}

}

func main() {
	database.Init()
	go startWeb()
	go startScheduler()
	// Keep the main function running to allow the goroutines to execute
	select {}
}
