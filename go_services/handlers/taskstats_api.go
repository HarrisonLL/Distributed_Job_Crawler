package handlers

import (
	"encoding/json"
	"go_services/database"
	"go_services/models"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

type TaskStats struct {
	TimePeriod   string `json:"time_period"`
	CompanyName  string `json:"company_name"`
	JobTypeName  string `json:"job_type_name"`
	SuccessCount int    `json:"success_count"`
	FailureCount int    `json:"failure_count"`
}

// GET request to fetch task statistics
func GetTaskStats(c *gin.Context) {
	var tasks []models.Task
	if err := database.DB.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Sort tasks by DateTime in descending order
	sort.Slice(tasks, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, tasks[i].DateTime)
		timeJ, _ := time.Parse(time.RFC3339, tasks[j].DateTime)
		return timeI.After(timeJ)
	})

	// Limit to the most recent 24 tasks
	if len(tasks) > 24 {
		tasks = tasks[:24]
	}

	statsMap := make(map[string]map[string]map[string]*TaskStats)

	for _, task := range tasks {
		var args struct {
			Company string `json:"company"`
			JobType string `json:"job_type"`
		}

		// Serialize JSONMap to JSON bytes
		argsBytes, err := json.Marshal(task.Args)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := json.Unmarshal(argsBytes, &args); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		dateTime, err := time.Parse(time.RFC3339, task.DateTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		dateHour := dateTime.Format("2006-01-02 15")

		if _, ok := statsMap[dateHour]; !ok {
			statsMap[dateHour] = make(map[string]map[string]*TaskStats)
		}
		if _, ok := statsMap[dateHour][args.Company]; !ok {
			statsMap[dateHour][args.Company] = make(map[string]*TaskStats)
		}
		if _, ok := statsMap[dateHour][args.Company][args.JobType]; !ok {
			statsMap[dateHour][args.Company][args.JobType] = &TaskStats{
				TimePeriod:  dateHour,
				CompanyName: args.Company,
				JobTypeName: args.JobType,
			}
		}

		stats := statsMap[dateHour][args.Company][args.JobType]
		stats.SuccessCount += len(task.SuccessJobIDs)
		stats.FailureCount += len(task.FailedJobIDs)
	}

	var statsList []*TaskStats
	for _, companyStats := range statsMap {
		for _, jobTypeStats := range companyStats {
			for _, stats := range jobTypeStats {
				statsList = append(statsList, stats)
			}
		}
	}

	c.JSON(http.StatusOK, statsList)
}
