package main

import (
	"bytes"
	"fmt"
	"go_services/database"
	"go_services/handlers"
	"go_services/models"
	"go_services/utils"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func startWeb() {
	router := gin.Default()
	router.GET("/api/v1/tasks", handlers.GetTasks)
	router.PATCH("/api/v1/tasks/:task_id", handlers.UpdateTask)
	router.GET("/api/v1/tasks/:task_id", handlers.GetTaskByID)
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}

func startScheduler() {
	mode := os.Getenv("MODE")
	for {
		var jobTypes []models.JobType
		if err := database.DB.Find(&jobTypes).Error; err != nil {
			log.Fatalf("Failed to fetch job types: %v", err)
		}

		for _, jobType := range jobTypes {
			if mode == "docker" {
				imageID, err := utils.PullDockerImage(jobType.DockerImageName)
				if err != nil {
					log.Printf("Failed to pull Docker image for company %s: %v", jobType.CompanyName, err)
					continue
				}
				if jobType.DockerImageID != imageID {
					jobType.DockerImageID = imageID
					jobType.PullDate = time.Now().Format(time.RFC3339)
					if err := database.DB.Save(&jobType).Error; err != nil {
						log.Printf("Failed to update Docker image for company %s: %v", jobType.CompanyName, err)
						continue
					}
				}
			}

			taskID := uuid.New().String()
			envVars := []string{
				fmt.Sprintf("MONGOURL=%s", os.Getenv("MONGOURL")),
			}
			htmlPath := os.Getenv("HTML_PATH")

			// Start the crawler work
			if mode == "docker" {
				go func(jobType models.JobType) {
					volumeMappings := []string{
						htmlPath + ":/app/html_data",
					}
					cmd := []string{
						"--job_type", jobType.JobTypeName,
						"--location", "USA",
						"--company", jobType.CompanyName,
						"--task_id", taskID,
					}
					containerID, err := utils.RunDockerContainer(jobType.DockerImageName, envVars, volumeMappings, cmd, false)
					if err != nil {
						log.Printf("Failed to start crawler for company %s: %v", jobType.CompanyName, err)
					} else {
						log.Printf("Started container %s for company %s", containerID, jobType.CompanyName)
						args := models.JSONMap{
							"job_type": jobType.JobTypeName,
							"location": "USA",
							"company":  jobType.CompanyName,
						}
						err = database.CreateTask(taskID, containerID, args, false, "")
						if err != nil {
							log.Printf("Failed to create task for company %s: %v", jobType.CompanyName, err)
						}
					}
				}(jobType)
			} else {
				go func(jobType models.JobType) {
					pythonCmdDir := os.Getenv("PYTHONFILEPATH")
					pythonCmd := exec.Command("python3", "main.py",
						"--job_type", jobType.JobTypeName,
						"--location", "USA",
						"--company", jobType.CompanyName,
						"--task_id", taskID,
					)
					pythonCmd.Env = append(os.Environ(), envVars...)
					pythonCmd.Dir = pythonCmdDir
					var stderr bytes.Buffer
					pythonCmd.Stderr = &stderr
					if err := pythonCmd.Start(); err != nil {
						log.Printf("Failed to start crawler for company %s: %v", jobType.CompanyName, err, stderr.String())
					} else {
						log.Printf("Started Python crawler for company %s", jobType.CompanyName)
						args := models.JSONMap{
							"job_type": jobType.JobTypeName,
							"location": "USA",
							"company":  jobType.CompanyName,
						}
						err = database.CreateTask(taskID, "", args, false, "")
						if err != nil {
							log.Printf("Failed to create task for company %s: %v", jobType.CompanyName, err)
						}
						// Start a thread to wait till process finishes and release its resouce
						go func() {
							if err := pythonCmd.Wait(); err != nil {
								log.Printf("Python crawler for company %s finished with error: %v", jobType.CompanyName, err)
							} else {
								log.Printf("Python crawler for company %s finished successfully", jobType.CompanyName)
							}
						}()
					}
				}(jobType)
			}

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
