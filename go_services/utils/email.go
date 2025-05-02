package utils

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"regexp"
	"strings"
)

type JobDetail struct {
	Title          string
	Description    string
	Qualifications string
	URL            string
}

func replaceNewLines(text string) string {
	// Collapse 2+ newlines to exactly 2
	re := regexp.MustCompile(`\n{2,}`)
	text = re.ReplaceAllString(text, "\n")
	text = strings.ReplaceAll(text, "\n", "<br>")
	return text
}

func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}

func SendEmailByJobType(username, email, jobType string, jobsByCompany map[string][]JobDetail) error {
	from := os.Getenv("SENDER_MAIL")
	password := os.Getenv("SENDER_PASS")

	to := []string{email}
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Color emoji square and mapping
	companyColors := map[string]string{
		"amazon":     "#F79B1B",
		"google":     "#F4B400",
		"meta":       "#4267B2",
		"salesforce": "#00A1E0",
		"uber":       "#333333",
	}

	var messageBody strings.Builder
	messageBody.WriteString(fmt.Sprintf("<h3>Hello %s,</h3>\n", username))
	messageBody.WriteString(fmt.Sprintf("<p>We found new jobs for <strong>%s</strong>:</p>\n", strings.Title(jobType)))

	for company, jobs := range jobsByCompany {
		color := companyColors[strings.ToLower(company)]
		messageBody.WriteString(fmt.Sprintf("<h3 style=\"font-size: 20px;\"><span style=\"color: %s; font-size: 20px;\">■</span> <i>%s</i></h3>\n", color, strings.Title(company)))
		messageBody.WriteString("<ul>\n")

		for i, job := range jobs {
			descriptionPreview := truncateText(job.Description, 300)
			qualificationsPreview := truncateText(job.Qualifications, 500)

			messageBody.WriteString("<li style=\"margin-bottom: 15px; font-size: 15px;\">\n")
			messageBody.WriteString(fmt.Sprintf("<strong>Job Title:</strong> <a href=\"%s\" style=\"font-size: 15px;\">%s</a><br>\n", job.URL, job.Title))

			messageBody.WriteString("<div style=\"margin-top: 6px;\"><strong>Description:</strong><br>\n")
			messageBody.WriteString(fmt.Sprintf("<div style=\"padding-left: 20px; font-size: 14px;\">%s</div></div>\n", replaceNewLines(descriptionPreview)))

			messageBody.WriteString("<div style=\"margin-top: 6px;\"><strong>Qualifications:</strong><br>\n")
			messageBody.WriteString(fmt.Sprintf("<div style=\"padding-left: 20px; font-size: 14px;\">%s</div></div>\n", replaceNewLines(qualificationsPreview)))

			messageBody.WriteString("</li>\n")

			if i != len(jobs)-1 {
				messageBody.WriteString("<hr style=\"border: 0.5px dashed #cccccc; margin: 10px 0;\">\n")
			}
		}
		messageBody.WriteString("</ul>\n")
		messageBody.WriteString("<hr style=\"border: 1.5px solid #646262; margin: 20px 0;\">\n")
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
