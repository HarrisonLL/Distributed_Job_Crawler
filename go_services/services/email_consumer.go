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
	collection := mongoClient.Database(fmt.Sprintf("%s_jobcrawler", strings.ToLower(company))).Collection("jobs")
	var jobs []utils.JobDetail
	for _, jobID := range jobIDs {
		var job struct {
			Title       string `bson:"title"`
			Description string `bson:"description"`
			URL         string `bson:"url"`
		}
		err := collection.FindOne(context.TODO(), bson.M{"id": jobID}).Decode(&job)
		if err != nil {
			log.Printf("Failed to fetch job details for jobID %s: %v", jobID, err)
			continue
		}
		jobs = append(jobs, utils.JobDetail{
			Title:       job.Title,
			Description: job.Description,
			URL:         job.URL,
		})
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
			var emailData struct {
				Username string   `json:"username"`
				Email    string   `json:"email"`
				JobIDs   []string `json:"jobIDs"`
				Company  string   `json:"company"`
			}

			if err := json.Unmarshal(d.Body, &emailData); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			jobs, err := fetchJobDetailsFromMongo(emailData.JobIDs, emailData.Company)
			if err != nil {
				log.Printf("Failed to fetch job details from MongoDB: %v", err)
				continue
			}

			err = utils.SendEmail(emailData.Username, emailData.Email, emailData.Company, jobs)
			if err != nil {
				log.Printf("Failed to send email: %v", err)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
