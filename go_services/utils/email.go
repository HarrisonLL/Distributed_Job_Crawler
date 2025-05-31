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
	Company        string
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

	companyColors := map[string]string{
		"amazon":         "#F79B1B",
		"google":         "#F4B400",
		"meta":           "#4267B2",
		"salesforce":     "#00A1E0",
		"uber":           "#333333",
		"linkedin_posts": "#0077B5",
		"microsoft":      "#107C10",
	}

	var messageBody strings.Builder
	messageBody.WriteString(fmt.Sprintf("<h3>Hello %s,</h3>\n", username))
	messageBody.WriteString(fmt.Sprintf("<p>We found new jobs for <strong>%s</strong>:</p>\n", strings.Title(jobType)))

	for company, jobs := range jobsByCompany {
		if company == "linkedin_posts" {
			continue
		}
		color := companyColors[strings.ToLower(company)]
		messageBody.WriteString(fmt.Sprintf("<h3 style=\"font-size: 20px;\"><span style=\"color: %s; font-size: 20px;\">■</span> <i>%s</i></h3>\n", color, strings.Title(company)))
		messageBody.WriteString("<ul>\n")

		for i, job := range jobs {
			descriptionPreview := job.Description
			qualificationsPreview := job.Qualifications
			if company == "amazon" {
				descriptionPreview = truncateText(job.Description, 600)
				qualificationsPreview = truncateText(job.Qualifications, 600)
			}

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

	if jobs, ok := jobsByCompany["linkedin_posts"]; ok && len(jobs) > 0 {
		color := companyColors["linkedin_posts"]
		messageBody.WriteString(fmt.Sprintf("<h3 style=\"font-size: 20px;\"><span style=\"color: %s; font-size: 20px;\">■</span> <i>LinkedIn Posts</i></h3>\n", color))
		// Display a table
		messageBody.WriteString(`<table style="border-collapse: collapse; width: 100%; font-size: 15px;">
			<tr>
				<th style="border: 1px solid #ccc; padding: 8px; background-color: #f2f2f2;">Title</th>
				<th style="border: 1px solid #ccc; padding: 8px; background-color: #f2f2f2;">Company</th>
				<th style="border: 1px solid #ccc; padding: 8px; background-color: #f2f2f2;">Posting Date</th>
			</tr>`)

		for _, job := range jobs {
			messageBody.WriteString(fmt.Sprintf(`
			<tr>
				<td style="border: 1px solid #ccc; padding: 8px;"><a href="%s">%s</a></td>
				<td style="border: 1px solid #ccc; padding: 8px;">%s</td>
				<td style="border: 1px solid #ccc; padding: 8px;">%s</td>
			</tr>`,
				job.URL,
				job.Title,
				job.Company,
				job.Qualifications,
			))
		}
		messageBody.WriteString("</table>\n")
		messageBody.WriteString("<hr style=\"border: 1.5px solid #646262; margin: 20px 20px;\">\n")
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
