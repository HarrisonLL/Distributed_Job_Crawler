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

func RunProcessOnHost(pythonCmd *exec.Cmd, jobType models.JobType, taskID string, wg *sync.WaitGroup) {
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
			defer wg.Done()
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
}
