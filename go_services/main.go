package main

import (
	"flag"
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
	router.LoadHTMLGlob("static/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "stats.html", nil)
	})
	router.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})

	router.GET("/api/v1/tasks", handlers.GetTasks)
	router.GET("/api/v1/tasks/:task_id", handlers.GetTaskByID)
	router.PATCH("/api/v1/tasks/:task_id", handlers.UpdateTask)

	// Page api
	router.POST("/api/v1/register", handlers.RegisterUser)
	router.GET("/api/v1/task_stats", handlers.GetTaskStats)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}

func main() {
	serviceMode := flag.String("service", "webScheduler", "Service mode: webScheduler, emailConsumer")
	flag.Parse()

	switch *serviceMode {
	case "webScheduler":
		database.Init()
		// scheduler
		s := gocron.NewScheduler(time.UTC)
		s.Every(6).Hours().Do(services.CrawlerTaskBase)
		s.StartAsync()
		// web
		go startWeb()
		select {}
	case "emailConsumer":
		go services.StartEmailConsumer()
		select {}
	default:
		log.Fatalf("Invalid service mode: %s", *serviceMode)
	}

}
