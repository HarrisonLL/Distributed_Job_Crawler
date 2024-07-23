package utils

import (
	"log"
	"net/smtp"
	"os"
	"strings"
)

func SendEmail(username, email string, jobIDs []string) error {
	from := os.Getenv("SENDER_MAIL")
	password := os.Getenv("SENDER_PASS")

	to := []string{
		email,
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	jobIDList := strings.Join(jobIDs, ", ")

	message := []byte("Subject: Job Alert!\r\n\r\n" +
		"Hello " + username + ",\n\n" +
		"New jobs found for you: " + jobIDList)

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		return err
	}

	log.Printf("Email sent to: %s", email)
	return nil
}
