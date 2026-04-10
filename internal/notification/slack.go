package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SlackNotifier sends alerts via Slack webhooks.
type SlackNotifier struct {
	webhookURL string
	client    *http.Client
}

// NewSlackNotifier creates a new Slack notifier.
func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendAlert sends an alert via Slack webhook.
func (s *SlackNotifier) SendAlert(ctx context.Context, alert Alert) error {
	message := s.formatAlert(alert)

	payload := map[string]interface{}{
		"text": message,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack webhook error: %s", string(body))
	}

	return nil
}

// formatAlert formats an alert for Slack.
func (s *SlackNotifier) formatAlert(alert Alert) string {
	emoji := ":white_check_mark:"
	if alert.Status == "down" {
		emoji = ":x:"
	}

	timestamp := time.Unix(alert.Timestamp, 0).Format("2006-01-02 15:04:05")

	message := fmt.Sprintf(
		"%s *Service Alert*\n\n"+
			"• *Service:* %s\n"+
			"• *Status:* %s\n"+
			"• *IP:* %s\n"+
			"• *Time:* %s",
		emoji, alert.ServiceName, alert.Status, alert.IPAddress, timestamp,
	)

	if alert.Error != "" {
		message += fmt.Sprintf("\n• *Error:* `%s`", alert.Error)
	}

	return message
}
