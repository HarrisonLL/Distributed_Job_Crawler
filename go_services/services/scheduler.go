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

		// Create a set of jobIDs and aggregate them
		companyJobSet := make(map[string]map[string]struct{})

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
				if _, exists := companyJobSet[task.Company]; !exists {
					companyJobSet[task.Company] = make(map[string]struct{})
				}
				for _, jobID := range task.SuccessJobIDs {
					companyJobSet[task.Company][jobID] = struct{}{}
				}
			}
		}

		// Convert set map to slice and store in emailData
		for company, jobSet := range companyJobSet {
			var jobList []string
			for jobID := range jobSet {
				jobList = append(jobList, jobID)
			}
			emailData[company] = jobList
		}

		// No job found if username, email, jobType only
		if len(emailData) == 3 {
			continue
		}

		StartEmailProducer(emailData)
	}
}

func CrawlerTaskBase() {
	concurrency := utils.GetConcurrency()
	mode := utils.GetMode()
	mongoURL, err := utils.GetURL("MONGOURL")
	location := utils.GetLocation()
	if err != "" {
		log.Fatalf("Failed to get MONGOURL: %v", err)
	}
	gsURL, err := utils.GetURL("GS_URL")
	if err != "" {
		log.Fatalf("Failed to get GS_URL: %v", err)
	}

	var taskIDs []string
	var jobTypes []models.JobType
	if err := database.DB.Find(&jobTypes).Error; err != nil {
		log.Fatalf("Failed to fetch job types: %v", err)
	}
	envVars := []string{
		fmt.Sprintf("MONGOURL=%s", mongoURL),
		fmt.Sprintf("GS_URL=%s", gsURL),
	}
	if concurrency == 1 {
		// Blocking mode
		for _, jobType := range jobTypes {
			taskID := uuid.New().String()
			if mode == "host" {
				taskIDs = append(taskIDs, taskID)
				pythonCmdDir := os.Getenv("PYTHONFILEPATH")
				pythonCmd := exec.Command("python3", "main.py",
					"--job_type", jobType.JobTypeName,
					"--location", location,
					"--company", jobType.CompanyName,
					"--task_id", taskID,
				)
				pythonCmd.Env = append(os.Environ(), envVars...)
				pythonCmd.Dir = pythonCmdDir
				utils.RunProcessOnHost(pythonCmd, jobType, taskID, nil, nil, true)
			} else if mode == "docker" {
				taskIDs = append(taskIDs, taskID)
				dockerCmd := []string{
					"--job_type", jobType.JobTypeName,
					"--location", location,
					"--company", jobType.CompanyName,
					"--task_id", taskID,
				}
				utils.RunDockerContainer(envVars, []string{}, dockerCmd, jobType, taskID, false, nil, nil, true)
			} else {
				log.Fatalf("Invalid mode: %s", mode)
				return
			}
		}
		log.Printf("All Crawler Finished!")
		// Send msg to email queue
		ComposeEmailTask(taskIDs)

	} else {
		// Nonblocking mode
		sem := make(chan struct{}, concurrency)
		var wg sync.WaitGroup
		for _, jobType := range jobTypes {
			wg.Add(1)
			taskID := uuid.New().String()
			if mode == "host" {
				taskIDs = append(taskIDs, taskID)
				go func(jobType models.JobType) {
					pythonCmdDir := os.Getenv("PYTHONFILEPATH")
					pythonCmd := exec.Command("python3", "main.py",
						"--job_type", jobType.JobTypeName,
						"--location", location,
						"--company", jobType.CompanyName,
						"--task_id", taskID,
					)
					pythonCmd.Env = append(os.Environ(), envVars...)
					pythonCmd.Dir = pythonCmdDir
					sem <- struct{}{}                // Acquire a token
					releaseToken := func() { <-sem } // Release token when done
					utils.RunProcessOnHost(pythonCmd, jobType, taskID, releaseToken, &wg, false)
				}(jobType)
			} else if mode == "docker" {
				taskIDs = append(taskIDs, taskID)
				go func(jobType models.JobType) {
					dockerCmd := []string{
						"--job_type", jobType.JobTypeName,
						"--location", location,
						"--company", jobType.CompanyName,
						"--task_id", taskID,
					}
					sem <- struct{}{}                // Acquire a token
					releaseToken := func() { <-sem } // Release the token when done
					utils.RunDockerContainer(envVars, []string{}, dockerCmd, jobType, taskID, false, releaseToken, &wg, false)
				}(jobType)
			} else {
				log.Fatalf("Invalid mode: %s", mode)
				return
			}
		}
		wg.Wait()
		log.Printf("All Crawler Finished!")
		// Send msg to email queue
		ComposeEmailTask(taskIDs)
	}
}
