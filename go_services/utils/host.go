package utils

import (
	"bytes"
	"go_services/database"
	"go_services/models"
	"log"
	"os/exec"
	"strconv"
	"sync"
)

func RunProcessOnHost(pythonCmd *exec.Cmd, jobType models.JobType, taskID string, releaseToken func(), wg *sync.WaitGroup, blocking bool) {
	var stderr bytes.Buffer
	pythonCmd.Stderr = &stderr
	err := pythonCmd.Start()
	if err != nil {
		log.Printf("Failed to start crawler for company %s: %v", jobType.CompanyName, err)
		if wg != nil {
			wg.Done()
		}
		return
	}
	PID := strconv.Itoa(pythonCmd.Process.Pid)
	log.Printf("Started Python crawler for company %s", jobType.CompanyName)
	err = database.CreateTask(taskID, PID, jobType.CompanyName, jobType.JobTypeName, "USA")
	if err != nil {
		log.Printf("Failed to create task for company %s: %v", jobType.CompanyName, err)
	}
	if !blocking {
		// Start a thread to wait till process finishes and release its resource
		go func() {
			defer wg.Done()
			defer releaseToken()
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
	} else {
		if err := pythonCmd.Wait(); err != nil {
			log.Printf("Python crawler for company %s finished with error: %v", jobType.CompanyName, err)
			DBerr := database.UpdateTaskStatus(taskID, "", models.Error)
			if err != nil {
				log.Printf("Failed to update task %s: %v", taskID, DBerr)
			}
		} else {
			log.Printf("Python crawler for company %s finished successfully", jobType.CompanyName)
		}
	}
}
