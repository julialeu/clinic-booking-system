package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

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

func ToOutboxEvents(ctx context.Context, domainEvents []any) ([]shared.OutboxEvent, error) {
	result := make([]shared.OutboxEvent, 0, len(domainEvents))

	for _, domainEvent := range domainEvents {
		outboxEvent, err := toOutboxEvent(ctx, domainEvent)
		if err != nil {
			return nil, err
		}
		result = append(result, outboxEvent)
	}
	return result, nil
}

func toOutboxEvent(ctx context.Context, domainEvent any) (shared.OutboxEvent, error) {
	switch event := domainEvent.(type) {

	case appointment.AppointmentReserved:
		return build(ctx, EventAppointmentReserved, event.AppointmentId, event.OccurredOn, event)

	case appointment.AppointmentConfirmed:
		return build(ctx, EventAppointmentConfirmed, event.AppointmentId, event.OccurredOn, event)

	case appointment.AppointmentCancelled:
		return build(ctx, EventAppointmentCancelled, event.AppointmentId, event.OccurredOn, event)

	case appointment.AppointmentCompleted:
		return build(ctx, EventAppointmentCompleted, event.AppointmentId, event.OccurredOn, event)

	case appointment.AppointmentExpired:
		return build(ctx, EventAppointmentExpired, event.AppointmentId, event.OccurredOn, event)

	default:
		return shared.OutboxEvent{}, fmt.Errorf("unknown domain event type: %T", domainEvent)
	}
}

func build(
	ctx context.Context,
	eventType string,
	aggregateId string,
	occurredOn time.Time,
	payload any,
) (shared.OutboxEvent, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return shared.OutboxEvent{}, fmt.Errorf("marshalling %s: %w", eventType, err)
	}

	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	return shared.OutboxEvent{
		AggregateType: appointmentAggregate,
		AggregateId:   aggregateId,
		EventType:     eventType,
		Payload:       encoded,
		OccurredOn:    occurredOn,
		TraceContext:  carrier,
	}, nil
}
