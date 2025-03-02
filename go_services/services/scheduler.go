package services

import (
	"bytes"
	"fmt"
	"go_services/database"
	"go_services/models"
	"go_services/utils"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func CrawlerTaskBase() {
	mode := os.Getenv("MODE")
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

		// Start the crawler work
		if mode == "docker" {
			go func(jobType models.JobType) {
				cmd := []string{
					"--job_type", jobType.JobTypeName,
					"--location", "USA",
					"--company", jobType.CompanyName,
					"--task_id", taskID,
				}
				containerID, err := utils.RunDockerContainer(jobType.DockerImageName, envVars, []string{}, cmd, false)
				if err != nil {
					log.Printf("Failed to start crawler for company %s: %v", jobType.CompanyName, err)
				} else {
					log.Printf("Started container %s for company %s", containerID, jobType.CompanyName)
					err = database.CreateTask(taskID, containerID, jobType.CompanyName, jobType.JobTypeName, "USA")
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
					PID := strconv.Itoa(pythonCmd.Process.Pid)
					log.Printf("Started Python crawler for company %s", jobType.CompanyName)
					err = database.CreateTask(taskID, PID, jobType.CompanyName, jobType.JobTypeName, "USA")
					if err != nil {
						log.Printf("Failed to create task for company %s: %v", jobType.CompanyName, err)
					}
					// Start a thread to wait till process finishes and release its resource
					go func() {
						if err := pythonCmd.Wait(); err != nil {
							log.Printf("Python crawler for company %s finished with error: %v", jobType.CompanyName, err)
							DBerr := database.UpdateTaskStatus(taskID, "", models.Error)
							if err != nil {
								log.Printf("Failed to update task %s: %v", taskID, DBerr)
							}
						} else {
							log.Printf("Python crawler for company %s finished successfully", jobType.CompanyName)
						}
					}()
				}
			}(jobType)
		}
	}
}
