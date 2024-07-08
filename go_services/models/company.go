package models

type Company struct {
	CompanyName     string `gorm:"primaryKey" json:"company_name"` // Company name
	DockerImageName string `json:"docker_image_name"`              // Docker image name
	DockerImageID   string `json:"docker_image_id"`                // Docker image ID
	PullDate        string `json:"pull_date"`                      // Date the image was pulled
}
