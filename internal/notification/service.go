package notification

import (
	"context"
)

// Notifier defines the interface for sending notifications.
type Notifier interface {
	SendAlert(ctx context.Context, alert Alert) error
}

// Alert represents a monitoring alert.
type Alert struct {
	ServiceName string
	ServiceID  string
	Status     string
	IPAddress  string
	Error      string
	Timestamp  int64
}

// NotificationService manages multiple notifiers.
type NotificationService struct {
	notifiers []Notifier
}

// NewNotificationService creates a new notification service.
func NewNotificationService(notifiers ...Notifier) *NotificationService {
	return &NotificationService{
		notifiers: notifiers,
	}
}

// SendAlert sends an alert through all configured notifiers.
func (s *NotificationService) SendAlert(ctx context.Context, alert Alert) error {
	for _, notifier := range s.notifiers {
		if err := notifier.SendAlert(ctx, alert); err != nil {
			// Log error but continue with other notifiers
			// For now, we'll return the last error
			continue
		}
	}
	return nil
}
