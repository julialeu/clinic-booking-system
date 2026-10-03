package kafka

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/julialeu/clinic-booking-system/notification-service/internal/domain/notification"
)

var tracer = otel.Tracer("notification-service/infrastructure/kafka")

type EventHandler interface {
	Handle(ctx context.Context, reference notification.EventReference, payload []byte) error
}

type Consumer struct {
	client  *kgo.Client
	handler EventHandler
}

func NewConsumer(brokers []string, topics []string, group string, handler EventHandler) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumeTopics(topics...),
		kgo.ConsumerGroup(group),
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
		if err := c.processRecord(ctx, record); err != nil {
			failed = fmt.Errorf("handling offset %d: %w", record.Offset, err)
		}
	})

	return failed
}

func (c *Consumer) processRecord(ctx context.Context, record *kgo.Record) error {
	reference := notification.EventReference{
		Topic:     record.Topic,
		Partition: record.Partition,
		Offset:    record.Offset,
		EventType: headerValue(record, "event_type"),
	}

	originCtx := otel.GetTextMapPropagator().Extract(ctx, headerCarrier(record))

	ctx, span := tracer.Start(originCtx, "Consume "+reference.EventType,
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.source", record.Topic),
			attribute.Int64("messaging.kafka.offset", record.Offset),
			attribute.String("event.type", reference.EventType),
		),
	)
	defer span.End()

	if err := c.handler.Handle(ctx, reference, record.Value); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "handler failed")
		return err
	}

	return nil
}

func headerCarrier(record *kgo.Record) propagation.MapCarrier {
	carrier := propagation.MapCarrier{}
	for _, header := range record.Headers {
		carrier[header.Key] = strings.Trim(string(header.Value), `"`)
	}
	return carrier
}

func headerValue(record *kgo.Record, key string) string {
	for _, header := range record.Headers {
		if header.Key == key {
			return strings.Trim(string(header.Value), `"`)
		}
	}
	return ""
}

func (c *Consumer) Close() {
	c.client.Close()
}
