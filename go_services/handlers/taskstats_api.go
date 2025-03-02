package handlers

import (
	"go_services/database"
	"go_services/models"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

type JobStats struct {
	TimePeriod string `json:"time_period"`
	Company    string `json:"company"`
	JobType    string `json:"job_type"`
	JobCount   int    `json:"job_count"`
}

// GET request to fetch task statistics
func GetTaskStats(c *gin.Context) {
	var tasks []models.Task
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	now := time.Now()

	var startDate, endDate time.Time
	if startDateStr != "" {
		parsedStartDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDate = parsedStartDate
		}
	} else {
		startDate = now.Truncate(24 * time.Hour)
	}

	if endDateStr != "" {
		parsedEndDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDate = parsedEndDate.Add(24 * time.Hour)
		}
	} else {
		endDate = now.Add(24 * time.Hour)
	}

	startDateStr = startDate.Format("2006-01-02 15:04")
	endDateStr = endDate.Format("2006-01-02 15:04")
	taskQuery := database.DB.Where("date_time >= ? and date_time < ?", startDateStr, endDateStr)
	if queryErr := taskQuery.Find(&tasks).Error; queryErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch tasks: " + queryErr.Error(),
		})
		return
	}

	statsMap := make(map[string]map[string]map[string]*JobStats)

	// build grouped nested map
	var dateGroupFormat string
	daysDiff := endDate.Sub(startDate).Hours() / 24
	switch {
	case daysDiff <= 1:
		dateGroupFormat = "2006-01-02 15:04"
	case daysDiff > 1 && daysDiff <= 7:
		dateGroupFormat = "2006-01-02"
	case daysDiff > 7 && daysDiff <= 30:
		dateGroupFormat = "2006-W02"
	case daysDiff > 30:
		dateGroupFormat = "2006-01"
	}

	for _, task := range tasks {
		dateTime, err := time.Parse(time.RFC3339, task.DateTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		dateHour := dateTime.Format(dateGroupFormat)
		if _, ok := statsMap[dateHour]; !ok {
			statsMap[dateHour] = make(map[string]map[string]*JobStats)
		}
		if _, ok := statsMap[dateHour][task.Company]; !ok {
			statsMap[dateHour][task.Company] = make(map[string]*JobStats)
		}
		if stats, ok := statsMap[dateHour][task.Company][task.JobType]; !ok {
			statsMap[dateHour][task.Company][task.JobType] = &JobStats{
				TimePeriod: dateHour,
				Company:    task.Company,
				JobType:    task.JobType,
				JobCount:   task.NumbersOfJobs,
			}
		} else {
			stats.JobCount += task.NumbersOfJobs
		}
	}

	var statsList []*JobStats
	for _, companyStats := range statsMap {
		for _, jobTypeStats := range companyStats {
			for _, stats := range jobTypeStats {
				statsList = append(statsList, stats)
			}
		}
	}
	// sort with comparator
	sort.Slice(statsList, func(i, j int) bool {
		if statsList[i].TimePeriod != statsList[j].TimePeriod {
			return statsList[i].TimePeriod > statsList[j].TimePeriod
		}
		if statsList[i].Company != statsList[j].Company {
			return statsList[i].Company < statsList[j].Company
		}
		return statsList[i].JobType < statsList[j].JobType
	})
	c.JSON(http.StatusOK, statsList)
}
