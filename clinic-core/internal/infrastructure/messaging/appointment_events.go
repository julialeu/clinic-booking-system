package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/appointment"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/shared"
)

const appointmentAggregate = "appointment"

const (
	EventAppointmentReserved  = "appointment.reserved"
	EventAppointmentConfirmed = "appointment.confirmed"
	EventAppointmentCancelled = "appointment.cancelled"
	EventAppointmentCompleted = "appointment.completed"
	EventAppointmentExpired   = "appointment.expired"
)

func ToOutboxEvents(domainEvents []any) ([]shared.OutboxEvent, error) {
	result := make([]shared.OutboxEvent, 0, len(domainEvents))

	for _, domainEvent := range domainEvents {
		outboxEvent, err := toOutboxEvent(domainEvent)
		if err != nil {
			return nil, err
		}
		result = append(result, outboxEvent)
	}
	return result, nil
}

func toOutboxEvent(domainEvent any) (shared.OutboxEvent, error) {
	switch event := domainEvent.(type) {

	case appointment.AppointmentReserved:
		return build(EventAppointmentReserved, event.AppointmentId, event.OccurredOn, event)

	case appointment.AppointmentConfirmed:
		return build(EventAppointmentConfirmed, event.AppointmentId, event.OccurredOn, event)

	case appointment.AppointmentCancelled:
		return build(EventAppointmentCancelled, event.AppointmentId, event.OccurredOn, event)

	case appointment.AppointmentCompleted:
		return build(EventAppointmentCompleted, event.AppointmentId, event.OccurredOn, event)

	case appointment.AppointmentExpired:
		return build(EventAppointmentExpired, event.AppointmentId, event.OccurredOn, event)

	default:
		return shared.OutboxEvent{}, fmt.Errorf("unknown domain event type: %T", domainEvent)
	}
}

func build(
	eventType string,
	aggregateId string,
	occurredOn time.Time,
	payload any,
) (shared.OutboxEvent, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return shared.OutboxEvent{}, fmt.Errorf("marshalling %s: %w", eventType, err)
	}

	return shared.OutboxEvent{
		AggregateType: appointmentAggregate,
		AggregateId:   aggregateId,
		EventType:     eventType,
		Payload:       encoded,
		OccurredOn:    occurredOn,
	}, nil
}
