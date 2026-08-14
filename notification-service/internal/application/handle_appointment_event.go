package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/julialeu/clinic-booking-system/notification-service/internal/domain/notification"
)

type AppointmentEventHandler struct {
	processed notification.ProcessedEvents
	directory notification.PatientDirectory
	sender    notification.Sender
}

func NewAppointmentEventHandler(
	processed notification.ProcessedEvents,
	directory notification.PatientDirectory,
	sender notification.Sender,
) *AppointmentEventHandler {
	return &AppointmentEventHandler{
		processed: processed,
		directory: directory,
		sender:    sender,
	}
}

func (h *AppointmentEventHandler) Handle(
	ctx context.Context,
	reference notification.EventReference,
	payload []byte,
) error {
	body, wanted := composerFor(reference.EventType)
	if !wanted {
		return nil
	}

	var event notification.AppointmentEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("decoding %s: %w", reference.EventType, err)
	}

	fresh, err := h.processed.MarkProcessed(ctx, reference)
	if err != nil {
		return err
	}
	if !fresh {
		log.Printf("skipping duplicate event %s at offset %d", reference.EventType, reference.Offset)
		return nil
	}

	contact, err := h.directory.Lookup(ctx, event.PatientId)
	if err != nil {
		return fmt.Errorf("looking up patient %s: %w", event.PatientId, err)
	}

	message, err := notification.New(
		contact.Phone,
		notification.ChannelWhatsApp,
		body(notification.AppointmentDetails{
			PatientName: contact.Name,
			StartsAt:    event.StartsAt,
			TypeName:    "fisioterapia",
		}),
	)
	if err != nil {
		return fmt.Errorf("building notification: %w", err)
	}

	if err := h.sender.Send(ctx, message); err != nil {
		return fmt.Errorf("sending notification: %w", err)
	}
	return nil
}

type composer func(notification.AppointmentDetails) string

// composerFor devuelve la plantilla del evento, o false si este
// servicio no envía notificación para ese tipo.
func composerFor(eventType string) (composer, bool) {
	switch eventType {
	case notification.EventAppointmentConfirmed:
		return notification.ComposeConfirmation, true
	case notification.EventAppointmentCancelled:
		return notification.ComposeCancellation, true
	default:
		return nil, false
	}
}
