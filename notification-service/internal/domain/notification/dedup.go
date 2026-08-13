package notification

import "context"

type EventReference struct {
	Topic     string
	Partition int32
	Offset    int64
	EventType string
}

type ProcessedEvents interface {
	MarkProcessed(ctx context.Context, reference EventReference) (bool, error)
}
