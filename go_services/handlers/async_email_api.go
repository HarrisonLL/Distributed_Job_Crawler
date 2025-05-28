package handlers

import (
	"go_services/database"
	"go_services/models"
	"go_services/utils"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ComposeEmailTaskRequest struct {
	TaskIDs []string `json:"task_ids"`
}

func composeEmailTask(taskIDs []string) {
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

		utils.ProduceEmailTasks(emailData)
	}
}

// Async POST API to handle sending user email
func ComposeEmailTaskHandler(c *gin.Context) {
	var req ComposeEmailTaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	go composeEmailTask(req.TaskIDs)

	c.JSON(http.StatusAccepted, gin.H{"message": "Email task processing started"})
}
