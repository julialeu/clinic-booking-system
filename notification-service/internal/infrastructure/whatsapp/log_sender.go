package whatsapp

import (
	"context"
	"log"

	"github.com/julialeu/clinic-booking-system/notification-service/internal/domain/notification"
)

var _ notification.Sender = (*LogSender)(nil)

type LogSender struct{}

func NewLogSender() *LogSender {
	return &LogSender{}
}

func (s *LogSender) Send(_ context.Context, message notification.Notification) error {
	log.Printf("[%s → %s] %s", message.Channel(), message.Recipient(), message.Body())
	return nil
}
