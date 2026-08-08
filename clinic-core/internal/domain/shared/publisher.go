package shared

import "context"

type Message struct {
	Topic   string
	Key     string
	Payload []byte
	Headers map[string]string
}

type EventPublisher interface {
	Publish(ctx context.Context, messages ...Message) error
	Close()
}
