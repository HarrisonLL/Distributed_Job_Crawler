package utils

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

type JobDetail struct {
	Title       string
	Description string
	URL         string
}

func SendEmail(username, email string, company string, jobs []JobDetail) error {
	from := os.Getenv("SENDER_MAIL")
	password := os.Getenv("SENDER_PASS")

	to := []string{
		email,
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	var messageBody strings.Builder
	messageBody.WriteString(fmt.Sprintf("<h3>Hello %s,</h3>\n", username))
	messageBody.WriteString(fmt.Sprintf("<p>We found new jobs from <strong>%s</strong> for you:</p>\n", company))

	for _, job := range jobs {
		messageBody.WriteString(fmt.Sprintf("<p><strong>Job Title:</strong> %s<br>\n", job.Title))
		messageBody.WriteString(fmt.Sprintf("<strong>Description:</strong> %s<br>\n", job.Description))
		messageBody.WriteString(fmt.Sprintf("<strong>URL:</strong> <a href=\"%s\">%s</a></p>\n", job.URL, job.URL))
		messageBody.WriteString("<hr>\n")
	}

	message := []byte("Subject: Job Alert!\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n" + messageBody.String())

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		return err
	}

	log.Printf("Email sent to: %s", email)
	return nil
}
