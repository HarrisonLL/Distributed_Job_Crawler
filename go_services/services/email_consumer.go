package services

import (
	"context"
	"encoding/json"
	"fmt"
	"go_services/utils"
	"log"
	"os"
	"strings"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

func initMongoDB() {
	var err error
	mongoURI := os.Getenv("MONGOURL")
	mongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
}

func fetchJobDetailsFromMongo(jobIDs []string, company string) ([]utils.JobDetail, error) {
	dbName := fmt.Sprintf("%s_jobcrawler", strings.ToLower(company))
	collection := mongoClient.Database(dbName).Collection("jobs")

	var jobs []utils.JobDetail

	for _, jobID := range jobIDs {
		if strings.ToLower(company) == "linkedin_posts" {
			var doc struct {
				JobPosts []struct {
					Title       string `bson:"title"`
					Company     string `bson:"company"`
					Link        string `bson:"link"`
					PostingDate string `bson:"posting_date"`
				} `bson:"job_posts"`
			}
			err := collection.FindOne(context.TODO(), bson.M{"id": jobID}).Decode(&doc)
			if err != nil {
				log.Printf("Failed to fetch LinkedIn job list for id %s: %v", jobID, err)
				continue
			}
			for _, post := range doc.JobPosts {
				jobs = append(jobs, utils.JobDetail{
					Title:          post.Title,
					URL:            post.Link,
					Description:    "",
					Qualifications: post.PostingDate,
					Company:        post.Company,
				})
			}
			jobs = utils.DedupAndSortLinkedInJobs(jobs)
		} else {
			var job struct {
				Title          string `bson:"title"`
				Description    string `bson:"description"`
				Qualifications string `bson:"qualifications"`
				URL            string `bson:"url"`
			}
			err := collection.FindOne(context.TODO(), bson.M{"id": jobID}).Decode(&job)
			if err != nil {
				log.Printf("Failed to fetch job details for jobID %s: %v", jobID, err)
				continue
			}
			jobs = append(jobs, utils.JobDetail{
				Title:          job.Title,
				Description:    job.Description,
				Qualifications: job.Qualifications,
				URL:            job.URL,
			})
		}
	}
	return jobs, nil
}

func StartEmailConsumer() {
	initMongoDB()
	conn, err := amqp.Dial(os.Getenv("MQ_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"email_tasks",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var emailData map[string]interface{}
			if err := json.Unmarshal(d.Body, &emailData); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}
			username := emailData["username"].(string)
			email := emailData["email"].(string)
			jobType := emailData["jobType"].(string)

			jobsByCompany := make(map[string][]utils.JobDetail)

			for key, val := range emailData {
				if key == "username" || key == "email" || key == "jobType" {
					continue
				}
				company := key
				jobIDsInterface := val.([]interface{})
				var jobIDs []string
				for _, id := range jobIDsInterface {
					jobIDs = append(jobIDs, id.(string))
				}

				jobs, err := fetchJobDetailsFromMongo(jobIDs, company)
				if err != nil {
					log.Printf("Failed to fetch job details for %s: %v", company, err)
					continue
				}
				jobsByCompany[company] = jobs
			}

			err := utils.SendEmailByJobType(username, email, jobType, jobsByCompany)
			if err != nil {
				log.Printf("Failed to send email: %v", err)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages ...")
	<-forever
}
