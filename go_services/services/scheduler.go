package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go_services/config"
	"go_services/database"
	"go_services/models"
	"go_services/utils"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
)

func emailRequest(taskIDs []string, gsURL string) {
	payload := map[string][]string{"task_ids": taskIDs}
	jsonBody, _ := json.Marshal(payload)
	asyncEmailURL := fmt.Sprintf("%s/api/v1/compose_email_task", gsURL)
	http.Post(asyncEmailURL, "application/json", bytes.NewBuffer(jsonBody))
}

func crawlerTaskBase() {
	concurrency := config.GetConcurrency()
	mode := config.GetMode()
	mongoURL, err := config.GetURL("MONGOURL")
	location := config.GetLocation()
	if err != "" {
		log.Fatalf("Failed to get MONGOURL: %v", err)
	}
	gsURL, err := config.GetURL("GS_URL")
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
		emailRequest(taskIDs, gsURL)

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
		emailRequest(taskIDs, gsURL)
	}
}

func StartScheduler() {
	s := gocron.NewScheduler(time.UTC)
	s.Every(config.GetCrawlingTimeInterval()).Hours().Do(crawlerTaskBase)
	s.StartBlocking()
}
