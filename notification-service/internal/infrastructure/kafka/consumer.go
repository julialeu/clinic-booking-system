package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/julialeu/clinic-booking-system/notification-service/internal/domain/notification"
)

type EventHandler interface {
	Handle(ctx context.Context, reference notification.EventReference, payload []byte) error
}

type Consumer struct {
	client  *kgo.Client
	handler EventHandler
}

func NewConsumer(brokers []string, topic, group string, handler EventHandler) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumeTopics(topic),
		kgo.ConsumerGroup(group),
		// Commit manual: confirmamos solo lo que hemos procesado.
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		return nil, fmt.Errorf("creating kafka consumer: %w", err)
	}
	return &Consumer{client: client, handler: handler}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		fetches := c.client.PollFetches(ctx)
		if ctx.Err() != nil {
			return nil
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			for _, err := range errs {
				log.Printf("consumer: fetch error on %s: %v", err.Topic, err.Err)
			}
			continue
		}

		if err := c.processFetches(ctx, fetches); err != nil {
			log.Printf("consumer: %v", err)
			continue
		}

		if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
			log.Printf("consumer: committing offsets: %v", err)
		}
	}
}

func (c *Consumer) processFetches(ctx context.Context, fetches kgo.Fetches) error {
	var failed error

	fetches.EachRecord(func(record *kgo.Record) {
		reference := notification.EventReference{
			Topic:     record.Topic,
			Partition: record.Partition,
			Offset:    record.Offset,
			EventType: headerValue(record, "event_type"),
		}

		if err := c.handler.Handle(ctx, reference, record.Value); err != nil {
			failed = fmt.Errorf("handling offset %d: %w", record.Offset, err)
		}
	})

	return failed
}

func headerValue(record *kgo.Record, key string) string {
	for _, header := range record.Headers {
		if header.Key == key {
			return string(header.Value)
		}
	}
	return ""
}

func (c *Consumer) Close() {
	c.client.Close()
}
