package main

import (
	"go_services/database"
	"go_services/handlers"
	"go_services/scheduler"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
)

func startWeb() {
	router := gin.Default()
	router.GET("/api/v1/tasks", handlers.GetTasks)
	router.PATCH("/api/v1/tasks/:task_id", handlers.UpdateTask)
	router.GET("/api/v1/tasks/:task_id", handlers.GetTaskByID)
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}

func main() {
	database.Init()

	// scheduler
	s := gocron.NewScheduler(time.UTC)
	s.Every(6).Hours().Do(scheduler.CrawlerTaskBase)
	s.StartAsync()

	// web
	go startWeb()
	select {}
}
