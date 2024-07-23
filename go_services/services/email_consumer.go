package services

import (
	"encoding/json"
	"go_services/utils"
	"log"
	"os"

	"github.com/streadway/amqp"
)

func StartEmailConsumer() {
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

	forever := make(chan bool) // Start a channel to sync goroutine threads

	go func() {
		for d := range msgs {
			var emailData map[string]interface{}
			if err := json.Unmarshal(d.Body, &emailData); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			log.Printf("Received a message: %s", d.Body)

			username, ok := emailData["username"].(string)
			if !ok {
				log.Printf("Failed to convert username")
				continue
			}
			email, ok := emailData["email"].(string)
			if !ok {
				log.Printf("Failed to convert email")
				continue
			}
			jobIDs, ok := emailData["jobIDs"].([]interface{})
			if !ok {
				log.Printf("Failed to convert jobIDs")
				continue
			}

			var jobIDStrings []string
			for _, jobID := range jobIDs {
				jobIDStr, ok := jobID.(string)
				if !ok {
					log.Printf("Failed to convert jobID")
					continue
				}
				jobIDStrings = append(jobIDStrings, jobIDStr)
			}

			err = utils.SendEmail(username, email, jobIDStrings)
			if err != nil {
				log.Printf("Failed to send email: %v", err)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
