package handlers

import (
	"go_services/database"
	"go_services/models"
	"go_services/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
		task.NumbersOfJobs = len(updatedTask.SuccessJobIDs)
	}

	if err := database.DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)

	if task.Status == models.Completed && len(task.SuccessJobIDs) > 0 {
		// send user email
		var users []models.User
		if err := database.DB.Where("email_subscription = ?", true).Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		taskCompany := strings.ToLower(task.Company)
		for _, user := range users {
			userCompanies := strings.ToLower(user.Company)
			if user.JobType == task.JobType && strings.Contains(userCompanies, taskCompany) {
				if len(task.SuccessJobIDs) == 0 {
					continue
				}
				emailData := map[string]interface{}{
					"username": user.Username,
					"email":    user.Email,
					"company":  task.Company,
					"jobIDs":   task.SuccessJobIDs,
				}
				services.StartEmailProducer(emailData)
			}
		}
	}

}
