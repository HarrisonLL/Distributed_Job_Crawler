package models

type JobType struct {
	JobTypeName     string `gorm:"primaryKey" json:"job_type_name"` // Job type name
	CompanyName     string `gorm:"primaryKey" json:"company_name"`  // Company name
	DockerImageName string `json:"docker_image_name"`               // Docker image name
	DockerImageID   string `json:"docker_image_id"`                 // Docker image ID
	PullDate        string `json:"pull_date"`                       // Date the image was pulled
}
