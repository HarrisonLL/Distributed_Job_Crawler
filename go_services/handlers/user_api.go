package handlers

import (
	"encoding/json"
	"fmt"
	"go_services/config"
	"go_services/database"
	"go_services/models"
	"go_services/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type inputLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type UpdateProfileInput struct {
	Email             string `json:"email"`
	Company           string `json:"company"`
	JobType           string `json:"job_type"` // Comma-separated values
	YOE               string `json:"yoe"`      // Comma-separated values
	EmailSubscription bool   `json:"email_subscription"`
}

func LoginUser(c *gin.Context) {
	var input inputLogin
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user models.User
	query := database.DB.Where("username = ?", input.Username).First(&user)
	if err := query.Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	if !utils.CheckPassword(user.Password, input.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	token, err := utils.GenerateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
	})
}

func RegisterUser(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	if err := database.DB.Where("username = ?", input.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
		return
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Username: input.Username,
		Password: hashedPassword,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully!"})
}

func GetUserProfile(c *gin.Context) {
	tokenUsername, err := utils.GetUsernameFromToken(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var user models.User
	if err := database.DB.Where("username = ?", tokenUsername).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username":           user.Username,
		"email":              user.Email,
		"job_type":           user.JobType,
		"company":            user.Company,
		"yoe":                user.YOE,
		"email_subscription": user.EmailSubscription,
	})
}

func UpdateUserProfile(c *gin.Context) {
	tokenUsername, err := utils.GetUsernameFromToken(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("username = ?", tokenUsername).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	updatedFields := map[string]interface{}{}
	updatedFields["email"] = input.Email
	updatedFields["job_type"] = input.JobType
	updatedFields["company"] = input.Company
	updatedFields["yoe"] = input.YOE
	updatedFields["email_subscription"] = input.EmailSubscription

	if len(updatedFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	fmt.Println(tokenUsername)
	if err := database.DB.Model(&user).Where("username = ?", tokenUsername).Updates(updatedFields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func ExternalGetTaskStats(c *gin.Context) {
	var stats []JobStats
	gsURL, errMsg := config.GetURL("GS_URL")

	if errMsg != "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get GS_URL: %s", errMsg)})
		return
	}

	query := c.Request.URL.RawQuery
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/tasks/stats?%s", gsURL, query))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to contact Task Service: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Task Service responded with %d", resp.StatusCode)})
		return
	}

	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to parse task stats: %v", err)})
		return
	}

	c.JSON(http.StatusOK, stats)
}
