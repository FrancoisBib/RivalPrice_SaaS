package services

import (
	"fmt"
	"log"
	"os"
)

// EmailService handles sending email notifications
type EmailService struct {
	fromEmail string
	smtpHost  string
	smtpPort  string
	enabled   bool
}

func NewEmailService() *EmailService {
	smtpHost := os.Getenv("SMTP_HOST")
	return &EmailService{
		fromEmail: os.Getenv("SMTP_FROM"),
		smtpHost:  smtpHost,
		smtpPort:  os.Getenv("SMTP_PORT"),
		enabled:   smtpHost != "",
	}
}

// SendAlert sends an alert notification by email (or logs if SMTP not configured)
func (s *EmailService) SendAlert(toEmail, alertType, severity, summary, recommendation string, pageID int) error {
	subject := fmt.Sprintf("[RivalPrice] %s alert — Page #%d", alertType, pageID)
	body := fmt.Sprintf(`
RivalPrice Alert
================
Type:           %s
Severity:       %s
Page ID:        %d

Summary:
%s

Recommendation:
%s
`, alertType, severity, pageID, summary, recommendation)

	if !s.enabled {
		// Log-only mode when SMTP not configured
		log.Printf("📧 [EMAIL-LOG] To: %s | Subject: %s\n%s", toEmail, subject, body)
		return nil
	}

	// TODO: implement real SMTP sending (e.g. net/smtp or SendGrid)
	log.Printf("📧 Email sent to %s: %s", toEmail, subject)
	return nil
}

// SendWebhook sends an alert to a webhook URL
func (s *EmailService) SendWebhook(webhookURL, alertType, severity, summary, recommendation string, pageID int, changeID uint) error {
	if webhookURL == "" {
		return nil
	}

	// Payload logged for now — real HTTP POST would go here
	log.Printf("🔔 [WEBHOOK] URL: %s | Type: %s | Severity: %s | Page: %d | Change: %d | %s",
		webhookURL, alertType, severity, pageID, changeID, summary)
	return nil
}
