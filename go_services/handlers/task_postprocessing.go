package handlers

import (
	"fmt"
	"go_services/database"
	"go_services/models"
	"go_services/utils"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/google/uuid"
)

func scheduleRetryTask(task models.Task) {
	mode := os.Getenv("MODE")
	// Wait for 1 hour before retrying
	time.Sleep(1 * time.Hour)

	newTaskID := uuid.New().String()
	args := task.Args
	args["retry"] = true
	args["parentTaskID"] = task.TaskID

	envVars := []string{
		fmt.Sprintf("MONGOURL=%s", os.Getenv("MONGOURL")),
	}
	htmlPath := os.Getenv("HTML_PATH")
	volumeMappings := []string{
		htmlPath + ":/app/html_data",
	}
	cmd := []string{
		"--retry=true",
		fmt.Sprintf("--task_id=%s", newTaskID),
		fmt.Sprintf("--parent_task_id=%s", task.TaskID),
	}

	if mode == "docker" {
		containerID, err := utils.RunDockerContainer(task.ContainerID, envVars, volumeMappings, cmd, false)
		if err != nil {
			log.Printf("Failed to start retry crawler for task %s: %v", task.TaskID, err)
		} else {
			log.Printf("Started retry container %s for task %s", containerID, task.TaskID)
			err = database.CreateTask(newTaskID, containerID, args, true, task.TaskID)
			if err != nil {
				log.Printf("Failed to create retry task for task %s: %v", task.TaskID, err)
			}
		}

	} else {
		pythonCmdDir := os.Getenv("PYTHONFILEPATH")
		pythonCmd := exec.Command("python", "main.py",
			"--retry=true",
			fmt.Sprintf("--task_id=%s", newTaskID),
			fmt.Sprintf("--parent_task_id=%s", task.TaskID))
		pythonCmd.Env = append(os.Environ(), envVars...)
		pythonCmd.Dir = pythonCmdDir
		if err := pythonCmd.Start(); err != nil {
			log.Printf("Failed to start retry crawler for task %s: %v", task.TaskID, err)
		} else {
			log.Printf("Started Python retry crawler for task %s", task.TaskID)
			err = database.CreateTask(newTaskID, "", args, true, task.TaskID)
			if err != nil {
				log.Printf("Failed to create retry task for task %s: %v", task.TaskID, err)
				return
			}
			// Wait till process finishes and release its resouce
			if err := pythonCmd.Wait(); err != nil {
				log.Printf("Python crawler for task %s finished with error: %v", task.TaskID, err)
			} else {
				log.Printf("Python crawler for task %s finished successfully", task.TaskID)
			}

		}

	}
}
