package main

import (
	"flag"
	"go_services/cli"
	"go_services/database"
	"go_services/handlers"
	"go_services/services"
	"log"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
)

func startWeb() {
	router := gin.Default()
	// Static web pages
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Task api
	router.GET("/api/v1/tasks", handlers.GetTasks)
	router.GET("/api/v1/tasks/:task_id", handlers.GetTaskByID)
	router.PATCH("/api/v1/tasks/:task_id", handlers.UpdateTask)

	// User api
	router.POST("/api/v1/register", handlers.RegisterUser)
	router.POST("/api/v1/login", handlers.LoginUser)
	router.GET("/api/v1/user_profile", handlers.GetUserProfile)
	router.PATCH("/api/v1/update_profile", handlers.UpdateUserProfile)
	router.GET("/api/v1/task_stats", handlers.GetTaskStats)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}

func startScheduler() {
	s := gocron.NewScheduler(time.UTC)
	s.Every(2).Hours().Do(services.CrawlerTaskBase)
	s.StartBlocking()
}

func main() {
	serviceMode := flag.String("service", "", "Service mode: web, scheduler, emailConsumer, CLI")
	flag.Parse()

	switch *serviceMode {
	case "web":
		database.Init()
		startWeb()
	case "scheduler":
		database.Init()
		startScheduler()
	case "emailConsumer":
		services.StartEmailConsumer()
	case "CLI":
		database.Init()
		cli.RunCLI()
	default:
		log.Fatalf("Invalid service mode: %s", *serviceMode)
	}

}
