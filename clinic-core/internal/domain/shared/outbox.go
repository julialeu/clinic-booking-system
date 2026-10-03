package shared

import (
	"context"
	"time"
)

type OutboxEvent struct {
	AggregateType string
	AggregateId   string
	EventType     string
	Payload       []byte
	OccurredOn    time.Time
	TraceContext  map[string]string
}

type OutboxRepository interface {
	Save(ctx context.Context, events ...OutboxEvent) error
}
