package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TelegramNotifier sends alerts via Telegram bot.
type TelegramNotifier struct {
	botToken string
	chatID   string
	client   *http.Client
}

// NewTelegramNotifier creates a new Telegram notifier.
func NewTelegramNotifier(botToken, chatID string) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: botToken,
		chatID:   chatID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendAlert sends an alert via Telegram.
func (t *TelegramNotifier) SendAlert(ctx context.Context, alert Alert) error {
	message := t.formatAlert(alert)

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

	payload := map[string]interface{}{
		"chat_id":    t.chatID,
		"text":       message,
		"parse_mode": "HTML",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API error: %s", string(body))
	}

	return nil
}

// formatAlert formats an alert for Telegram.
func (t *TelegramNotifier) formatAlert(alert Alert) string {
	emoji := "🟢"
	if alert.Status == "down" {
		emoji = "🔴"
	}

	timestamp := time.Unix(alert.Timestamp, 0).Format("2006-01-02 15:04:05")

	message := fmt.Sprintf(
		"%s <b>Service Alert</b>\n\n"+
			"<b>Service:</b> %s\n"+
			"<b>Status:</b> %s\n"+
			"<b>IP:</b> %s\n"+
			"<b>Time:</b> %s",
		emoji, alert.ServiceName, alert.Status, alert.IPAddress, timestamp,
	)

	if alert.Error != "" {
		message += fmt.Sprintf("\n<b>Error:</b> <code>%s</code>", htmlEscape(alert.Error))
	}

	return message
}

// GetUpdatesURL returns the URL for getting updates (for webhooks).
func (t *TelegramNotifier) GetUpdatesURL() string {
	return fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", t.botToken)
}

// SetWebhook sets a webhook for the Telegram bot.
func (t *TelegramNotifier) SetWebhook(ctx context.Context, webhookURL string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", t.botToken)

	params := url.Values{}
	params.Set("url", webhookURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("set webhook failed: status %d", resp.StatusCode)
	}

	return nil
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
