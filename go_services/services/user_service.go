package services

import (
	"go_services/handlers"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func StartUserSvc() {
	router := gin.Default()
	// Static web pages
	router.Delims("[[", "]]")
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// User api
	router.POST("/api/v1/register", handlers.RegisterUser)
	router.POST("/api/v1/login", handlers.LoginUser)
	router.GET("/api/v1/user_profile", handlers.GetUserProfile)
	router.PATCH("/api/v1/update_profile", handlers.UpdateUserProfile)
	router.GET("/api/v1/task_stats", handlers.ExternalGetTaskStats)

	if err := router.Run(":8090"); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}
