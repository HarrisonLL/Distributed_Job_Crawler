package cli

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type JobType struct {
	JobTypeName     string `gorm:"primaryKey" json:"job_type_name"`
	CompanyName     string `gorm:"primaryKey" json:"company_name"`
	DockerImageName string `json:"docker_image_name"`
	DockerImageID   string `json:"docker_image_id"`
	PullDate        string `json:"pull_date"`
}

func connectDB() (*gorm.DB, error) {
	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		return nil, fmt.Errorf("POSTGRES_URL is not set!")
	}
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func pullDockerImage(imageName string) (string, error) {
	fmt.Println("Pulling docker image: ", imageName)
	cmd := exec.Command("docker", "pull", imageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("!! Failed to pull Docker image: %s \n %s", err, string(output))
	}
	cmd = exec.Command("docker", "inspect", "--format={{.Id}}", imageName)
	imageIDBytes, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("!! Failed to get Docker image ID: %s", err)
	}
	imageID := string(imageIDBytes)
	log.Println("Successfully pulled Docker image:", imageName)
	return imageID, nil
}

func insertUpdateJobType(jobType JobType) {
	db, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	var jobRecord JobType
	result := db.Where("job_type_name = ? and company_name = ?", jobType.JobTypeName, jobType.CompanyName).First(&jobRecord)
	if result.RowsAffected == 0 {
		imageID, err := pullDockerImage(jobType.DockerImageName)
		if err != nil {
			log.Fatal(err)
		}
		db.Create(&JobType{
			JobTypeName:     jobType.JobTypeName,
			CompanyName:     jobType.CompanyName,
			DockerImageName: jobType.DockerImageName,
			DockerImageID:   imageID,
			PullDate:        time.Now().Format("2006-01-02 15:04:05"),
		})
		log.Println("Job type inserted successfully.")

	} else {
		if jobRecord.DockerImageName == jobType.DockerImageName {
			fmt.Println("Job type already exists, skipping update.")
			return
		}
		imageID, err := pullDockerImage(jobType.DockerImageName)
		if err != nil {
			log.Fatal(err)
		}
		db.Model(&jobRecord).Updates(JobType{
			DockerImageName: jobType.DockerImageName,
			DockerImageID:   imageID,
			PullDate:        time.Now().Format("2006-01-02 15:04:05"),
		})
		log.Println("Job type updated successfully.")
	}
}

func deleteJobType(jobType JobType) {
	db, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	var jobRecord JobType
	result := db.Where("job_type_name = ? and company_name = ?", jobType.JobTypeName, jobType.CompanyName).First(&jobRecord)
	if result.RowsAffected == 0 {
		fmt.Println("Job type does not exist, skipping delete.")
		return
	}
	db.Delete(&jobRecord)
	log.Println("Job type deleted successfully.")
}
