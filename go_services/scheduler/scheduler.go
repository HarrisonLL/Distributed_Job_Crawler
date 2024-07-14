package scheduler

import (
	"bytes"
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
						"task_id":  taskID,
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
}

func RetryTaskScheduler(task models.Task) {
	mode := os.Getenv("MODE")
	// Wait for 1 hour before retrying
	time.Sleep(1 * time.Hour)

	newTaskID := uuid.New().String()
	args := models.JSONMap{
		"retry":          true,
		"parent_task_id": task.TaskID,
		"task_id":        newTaskID,
	}

	envVars := []string{
		fmt.Sprintf("MONGOURL=%s", os.Getenv("MONGOURL")),
	}
	htmlPath := os.Getenv("HTML_PATH")

	if mode == "docker" {
		volumeMappings := []string{
			htmlPath + ":/app/html_data",
		}
		cmd := []string{
			"--retry=true",
			fmt.Sprintf("--task_id=%s", newTaskID),
			fmt.Sprintf("--parent_task_id=%s", task.TaskID),
		}
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
		pythonCmd := exec.Command("python3", "main.py",
			"--retry=true",
			fmt.Sprintf("--task_id=%s", newTaskID),
			fmt.Sprintf("--parent_task_id=%s", task.TaskID))
		pythonCmd.Env = append(os.Environ(), envVars...)
		pythonCmd.Dir = pythonCmdDir
		var stderr bytes.Buffer
		pythonCmd.Stderr = &stderr
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

// To-DO:
// Add helper functions in any condition when processes ends with error
// update db with status = 3
