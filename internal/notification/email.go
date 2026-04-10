package notification

import (
	"context"
	"fmt"
	"net/smtp"
	"time"
)

// EmailConfig holds email configuration.
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	ToEmails     []string
}

// EmailNotifier sends alerts via email.
type EmailNotifier struct {
	config EmailConfig
}

// NewEmailNotifier creates a new email notifier.
func NewEmailNotifier(config EmailConfig) *EmailNotifier {
	return &EmailNotifier{
		config: config,
	}
}

// SendAlert sends an alert via email.
func (e *EmailNotifier) SendAlert(ctx context.Context, alert Alert) error {
	if len(e.config.ToEmails) == 0 {
		return nil
	}

	body := e.formatAlert(alert)
	smtpAddr := fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort)

	auth := smtp.PlainAuth("", e.config.SMTPUsername, e.config.SMTPPassword, e.config.SMTPHost)

	for _, toEmail := range e.config.ToEmails {
		err := smtp.SendMail(smtpAddr, auth, e.config.FromEmail, []string{toEmail}, []byte(body))
		if err != nil {
			return fmt.Errorf("send email to %s: %w", toEmail, err)
		}
	}

	return nil
}

// formatAlert formats an alert for email.
func (e *EmailNotifier) formatAlert(alert Alert) string {
	emoji := "✅"
	if alert.Status == "down" {
		emoji = "❌"
	}

	timestamp := time.Unix(alert.Timestamp, 0).Format("2006-01-02 15:04:05 MST")

	message := fmt.Sprintf(
		"From: %s <%s>\n"+
			"To: %s\n"+
			"Subject: %s\n"+
			"MIME-Version: 1.0\n"+
			"Content-Type: text/plain; charset=utf-8\n\n",
		e.config.FromName, e.config.FromEmail, e.config.FromEmail,
		fmt.Sprintf("[%s] Service Alert: %s", alert.Status, alert.ServiceName),
	)

	message += fmt.Sprintf(
		"%s Service Alert\n\n"+
			"Service: %s\n"+
			"Status: %s\n"+
			"IP Address: %s\n"+
			"Timestamp: %s\n",
		emoji, alert.ServiceName, alert.Status, alert.IPAddress, timestamp,
	)

	if alert.Error != "" {
		message += fmt.Sprintf("Error: %s\n", alert.Error)
	}

	message += "\n\n---\nApp Monitoring System"

	return message
}
