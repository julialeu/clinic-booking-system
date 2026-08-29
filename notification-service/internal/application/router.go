package application

import (
	"context"
	"fmt"

	"github.com/julialeu/clinic-booking-system/notification-service/internal/domain/notification"
)

type Handler interface {
	Handle(ctx context.Context, reference notification.EventReference, payload []byte) error
}

type TopicRouter struct {
	handlers map[string]Handler
}

func NewTopicRouter(handlers map[string]Handler) *TopicRouter {
	return &TopicRouter{handlers: handlers}
}

func (r *TopicRouter) Handle(
	ctx context.Context,
	reference notification.EventReference,
	payload []byte,
) error {
	handler, found := r.handlers[reference.Topic]
	if !found {
		return fmt.Errorf("no handler registered for topic %s", reference.Topic)
	}

	return handler.Handle(ctx, reference, payload)
}
