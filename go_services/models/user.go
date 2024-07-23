package models

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	JobType  string `json:"job_type"`
	YOE      string `json:"yoe"`
	Company  string `json:"company"`
}
