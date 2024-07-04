package main

import (
	"encoding/json"
	"job-scheduler/database"
	"job-scheduler/handlers"
	"job-scheduler/models"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
)

func main() {
	// Initialize Database
	database.Init()

	// Initialize Router
	router := gin.Default()
	router.GET("/tasks", handlers.GetTasks)
	router.POST("/tasks", handlers.CreateTask)
	router.PATCH("/tasks", handlers.UpdateTask)

	// Initialize Scheduler
	s := gocron.NewScheduler(time.UTC)
	// Schedule the crawling task to run every 6 hours
	s.Every(6).Hours().Do(scheduleCrawling)
	s.StartAsync()

	// Start Server
	router.Run(":8080")
}

// scheduleCrawling is the function that runs every 6 hours to retry failed tasks
func scheduleCrawling() {
	// Fetch tasks with failed status
	var tasks []models.Task
	database.DB.Where("status = ?", "failed").Find(&tasks)

	for _, task := range tasks {
		// Retry logic for each failed task
		log.Println("Retrying task:", task.TaskID)
		// Update the task status to retrying or in-progress
		task.Status = "retrying"
		database.DB.Save(&task)

		// Simulate sending the task to workers
		sendTaskToWorker(task)
	}
}

// sendTaskToWorker simulates sending the task to a worker
func sendTaskToWorker(task models.Task) {
	argsJSON, err := json.Marshal(task.Args)
	if err != nil {
		log.Fatalf("Failed to marshal task args: %v", err)
	}

	log.Printf("Sending task %s to worker with args: %s\n", task.TaskID, string(argsJSON))
	// Here you would add the actual code to send the task to the worker
	// For now, we just simulate a successful job run
	task.Status = "success"
	task.SuccessRate = 1.0
	task.SuccessJobIDs = append(task.SuccessJobIDs, "example_job_id_1")
	task.FailedJobIDs = nil
	database.DB.Save(&task)
}
