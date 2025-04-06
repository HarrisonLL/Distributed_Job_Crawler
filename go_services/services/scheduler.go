package services

import (
	"fmt"
	"go_services/database"
	"go_services/models"
	"go_services/utils"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

func ComposeEmailTask(taskIDs []string) {
	var users []models.User
	if err := database.DB.Where("email_subscription = ?", true).Find(&users).Error; err != nil {
		log.Printf("Failed to fetch users: %v", err)
	}
	for _, user := range users {
		if !user.EmailSubscription {
			continue
		}
		userCompanies := strings.ToLower(user.Company)
		userJobType := strings.ToLower(user.JobType)
		emailData := map[string]interface{}{
			"username": user.Username,
			"email":    user.Email,
			"jobType":  userJobType,
		}
		for _, taskID := range taskIDs {
			var task models.Task
			result := database.DB.Where("task_id = ?", taskID).First(&task)
			if result.Error != nil {
				log.Printf("Could not fetch task %s: %v", taskID, result.Error)
				continue
			}
			if len(task.SuccessJobIDs) == 0 {
				continue
			}
			if strings.Contains(userCompanies, task.Company) && strings.Contains(userJobType, task.JobType) {
				emailData[task.Company] = task.SuccessJobIDs
			}
		}
		if len(emailData) == 3 {
			continue
		}
		StartEmailProducer(emailData)
	}
}

func CrawlerTaskBase() {
	mode := os.Getenv("MODE")
	var jobTypes []models.JobType
	if err := database.DB.Find(&jobTypes).Error; err != nil {
		log.Fatalf("Failed to fetch job types: %v", err)
	}
	var wg sync.WaitGroup
	var taskIDs []string
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
		taskIDs = append(taskIDs, taskID)
		envVars := []string{
			fmt.Sprintf("MONGOURL=%s", os.Getenv("MONGOURL")),
			fmt.Sprintf("GS_URL=%s", os.Getenv("GS_URL")),
		}

		// Start the crawler work
		if mode == "docker" {
			wg.Add(1)
			go func(jobType models.JobType) {
				dockerCmd := []string{
					"--job_type", jobType.JobTypeName,
					"--location", "USA",
					"--company", jobType.CompanyName,
					"--task_id", taskID,
				}
				containerID, err := utils.RunDockerContainer(jobType.DockerImageName, envVars, []string{}, dockerCmd, false, &wg)
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
			wg.Add(1)
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
				utils.RunProcessOnHost(pythonCmd, jobType, taskID, &wg)
			}(jobType)
		}
	}

	wg.Wait()
	log.Printf("All Crawler Finished!")
	// send to queue
	ComposeEmailTask(taskIDs)

}
