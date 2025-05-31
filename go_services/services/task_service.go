package services

import (
	"go_services/handlers"
	"log"

	"github.com/gin-gonic/gin"
)

func StartTaskSvc() {
	router := gin.Default()
	// Task api
	router.GET("/api/v1/tasks", handlers.GetTasks)
	router.GET("/api/v1/tasks/:task_id", handlers.GetTaskByID)
	router.PATCH("/api/v1/tasks/:task_id", handlers.UpdateTask)
	router.GET("/api/v1/tasks/stats", handlers.GetTaskStats)
	router.POST("/api/v1/compose_email_task", handlers.ComposeEmailTaskHandler)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}
