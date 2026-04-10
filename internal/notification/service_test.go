package notification

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNotifier is a mock implementation of Notifier.
type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) SendAlert(ctx context.Context, alert Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func TestNotificationService_SendAlert(t *testing.T) {
	mock1 := new(MockNotifier)
	mock2 := new(MockNotifier)
	mock3 := new(MockNotifier)

	alert := Alert{
		ServiceName: "Test Service",
		ServiceID:   "svc-123",
		Status:      "down",
		IPAddress:   "192.168.1.1",
		Error:       "Connection timeout",
		Timestamp:   time.Now().Unix(),
	}

	mock1.On("SendAlert", mock.Anything, alert).Return(nil)
	mock2.On("SendAlert", mock.Anything, alert).Return(nil)
	mock3.On("SendAlert", mock.Anything, alert).Return(nil)

	service := NewNotificationService(mock1, mock2, mock3)
	err := service.SendAlert(context.Background(), alert)

	assert.NoError(t, err)
	mock1.AssertExpectations(t)
	mock2.AssertExpectations(t)
	mock3.AssertExpectations(t)
}

func TestNotificationService_SendAlert_ContinueOnError(t *testing.T) {
	mock1 := new(MockNotifier)
	mock2 := new(MockNotifier)
	mock3 := new(MockNotifier)

	alert := Alert{
		ServiceName: "Test Service",
		ServiceID:   "svc-123",
		Status:      "down",
	}

	mock1.On("SendAlert", mock.Anything, alert).Return(assert.AnError)
	mock2.On("SendAlert", mock.Anything, alert).Return(assert.AnError)
	mock3.On("SendAlert", mock.Anything, alert).Return(nil)

	service := NewNotificationService(mock1, mock2, mock3)
	err := service.SendAlert(context.Background(), alert)

	// Service should continue even if some notifiers fail
	assert.NoError(t, err)
	mock1.AssertExpectations(t)
	mock2.AssertExpectations(t)
	mock3.AssertExpectations(t)
}

func TestNotificationService_NoNotifiers(t *testing.T) {
	service := NewNotificationService()
	alert := Alert{
		ServiceName: "Test Service",
		ServiceID:   "svc-123",
		Status:      "down",
	}

	err := service.SendAlert(context.Background(), alert)
	assert.NoError(t, err)
}

func TestTelegramFormatter(t *testing.T) {
	notifier := &TelegramNotifier{
		botToken: "test-token",
		chatID:   "123456",
	}

	upAlert := Alert{
		ServiceName: "My Service",
		Status:     "up",
		IPAddress:   "10.0.0.1",
		Timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC).Unix(),
	}

	message := notifier.formatAlert(upAlert)
	assert.Contains(t, message, "🟢")
	assert.Contains(t, message, "My Service")
	assert.Contains(t, message, "up")
	assert.Contains(t, message, "10.0.0.1")

	downAlert := Alert{
		ServiceName: "My Service",
		Status:     "down",
		IPAddress:   "10.0.0.1",
		Error:       "Connection refused",
		Timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC).Unix(),
	}

	message = notifier.formatAlert(downAlert)
	assert.Contains(t, message, "🔴")
	assert.Contains(t, message, "down")
	assert.Contains(t, message, "Connection refused")
}

func TestSlackFormatter(t *testing.T) {
	notifier := &SlackNotifier{
		webhookURL: "https://hooks.slack.com/test",
	}

	upAlert := Alert{
		ServiceName: "My Service",
		Status:     "up",
		IPAddress:   "10.0.0.1",
		Timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC).Unix(),
	}

	message := notifier.formatAlert(upAlert)
	assert.Contains(t, message, ":white_check_mark:")
	assert.Contains(t, message, "My Service")
	assert.Contains(t, message, "up")

	downAlert := Alert{
		ServiceName: "My Service",
		Status:     "down",
		IPAddress:   "10.0.0.1",
		Error:       "Connection refused",
		Timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC).Unix(),
	}

	message = notifier.formatAlert(downAlert)
	assert.Contains(t, message, ":x:")
	assert.Contains(t, message, "down")
	assert.Contains(t, message, "Connection refused")
}

func TestEmailFormatter(t *testing.T) {
	config := EmailConfig{
		FromEmail: "alerts@example.com",
		FromName:  "App Monitor",
		ToEmails:  []string{"admin@example.com"},
	}
	notifier := NewEmailNotifier(config)

	upAlert := Alert{
		ServiceName: "My Service",
		Status:     "up",
		IPAddress:   "10.0.0.1",
		Timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC).Unix(),
	}

	message := notifier.formatAlert(upAlert)
	assert.Contains(t, message, "✅")
	assert.Contains(t, message, "My Service")
	assert.Contains(t, message, "up")
	assert.Contains(t, message, "From: App Monitor <alerts@example.com>")
	assert.Contains(t, message, "To: alerts@example.com")

	downAlert := Alert{
		ServiceName: "My Service",
		Status:     "down",
		IPAddress:   "10.0.0.1",
		Error:       "Connection refused",
		Timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC).Unix(),
	}

	message = notifier.formatAlert(downAlert)
	assert.Contains(t, message, "❌")
	assert.Contains(t, message, "down")
	assert.Contains(t, message, "Connection refused")
}

func TestHtmlEscape(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<script>", "&lt;script&gt;"},
		{"&test", "&amp;test"},
		{"<>&", "&lt;&gt;&amp;"},
		{"normal text", "normal text"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := htmlEscape(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
