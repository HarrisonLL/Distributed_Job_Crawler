package handlers

import (
	"fmt"
	"go_services/database"
	"go_services/models"
	"go_services/utils"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET requests to fetch all tasks
func GetTasks(c *gin.Context) {
	var tasks []models.Task
	if err := database.DB.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// GET request to fetch a task by ID
func GetTaskByID(c *gin.Context) {
	taskID := c.Param("task_id")
	var task models.Task
	if err := database.DB.First(&task, "task_id = ?", taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

// PATCH requests to update an existing task
func UpdateTask(c *gin.Context) {
	taskID := c.Param("task_id")
	var task models.Task

	if err := database.DB.First(&task, "task_id = ?", taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var updatedTask models.Task
	if err := c.ShouldBindJSON(&updatedTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if updatedTask.Status != 0 {
		task.Status = updatedTask.Status
	}
	if updatedTask.CompletionRate != 0 {
		task.CompletionRate = updatedTask.CompletionRate
	}
	if len(updatedTask.SuccessJobIDs) > 0 {
		task.SuccessJobIDs = updatedTask.SuccessJobIDs
	}
	if len(updatedTask.FailedJobIDs) > 0 {
		task.FailedJobIDs = updatedTask.FailedJobIDs
	}

	if err := database.DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)

	if task.Status == models.Completed && len(task.FailedJobIDs) > 0 {
		go scheduleRetryTask(task)
	}
}

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
		containerID, err := utils.RunDockerContainer(task.ContainerID, envVars, volumeMappings, cmd)
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
			}
		}
	}
}
