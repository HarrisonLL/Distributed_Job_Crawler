package services

import (
	"go_services/handlers"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func StartWeb() {
	router := gin.Default()
	// Static web pages
	router.Delims("[[", "]]")
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Task api
	router.GET("/api/v1/tasks", handlers.GetTasks)
	router.GET("/api/v1/tasks/:task_id", handlers.GetTaskByID)
	router.PATCH("/api/v1/tasks/:task_id", handlers.UpdateTask)
	router.POST("/api/v1/compose_email_task", handlers.ComposeEmailTaskHandler)

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
