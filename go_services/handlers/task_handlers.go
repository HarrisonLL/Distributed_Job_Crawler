package handlers

import (
	"fmt"
	"go_services/database"
	"go_services/models"
	"go_services/utils"
	"log"
	"net/http"
	"os"
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

// POST requests to create a new task
func CreateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	if err := database.DB.Model(&task).Updates(updatedTask).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedTask)

	if task.Status == models.Completed {
		go scheduleRetryTask(task)
	}
}

// scheduleRetryTask schedules a retry for the failed jobs of a completed task
func scheduleRetryTask(task models.Task) {
	time.Sleep(1 * time.Hour)

	envVars := []string{
		fmt.Sprintf("TASKID=%s", retryTaskID),
		fmt.Sprintf("MONGOURL=%s", os.Getenv("MONGOURL")),
	}
	htmlPath := os.Getenv("HTML_PATH")
	volumeMappings := []string{
		htmlPath + ":/app/html_data",
	}
	cmd := []string{
		"--job_type", task.Args["job_type"].(string),
		"--location", task.Args["location"].(string),
		"--company", task.Args["company"].(string),
		"--retry=true",
	}

	// Start the crawler work for the retry task
	containerID, err := utils.RunDockerContainer(task.Args["company"].(string), envVars, volumeMappings, cmd)
	if err != nil {
		log.Printf("Failed to start retry crawler for task %s: %v", task.TaskID, err)
	} else {
		log.Printf("Started retry container %s for task %s", containerID, task.TaskID)
		retryTaskID := uuid.New().String()
		retryTask := models.Task{
			TaskID:         retryTaskID,
			ContainerID:    containerID,
			DateTime:       time.Now().Format(time.RFC3339),
			Args:           task.Args,
			Status:         models.Started,
			SuccessJobIDs:  []string{},
			FailedJobIDs:   task.FailedJobIDs,
			CompletionRate: 0.0,
			IsRetryTask:    true,
			ParentTaskID:   task.TaskID,
		}
		if err := database.DB.Create(&retryTask).Error; err != nil {
			log.Printf("Failed to create retry task: %v", err)
			return
		}
	}
}
