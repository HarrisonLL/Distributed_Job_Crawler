package models

type User struct {
	Username          string `json:"username" gorm:"unique;not null"`
	Email             string `json:"email" gorm:"unique;not null"`
	JobType           string `json:"job_type"`
	YOE               string `json:"yoe"`
	Company           string `json:"company"`
	Password          string `json:"-" gorm:"not null"`
	EmailSubscription bool   `json:"email_subscription"`
}
