package messaging

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/shared"
)

var _ shared.EventPublisher = (*KafkaPublisher)(nil)

type KafkaPublisher struct {
	client *kgo.Client
}

func NewKafkaPublisher(brokers []string) (*KafkaPublisher, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ProducerBatchCompression(kgo.SnappyCompression()),
		kgo.RequiredAcks(kgo.AllISRAcks()),
	)
	if err != nil {
		return nil, fmt.Errorf("creating kafka client: %w", err)
	}
	return &KafkaPublisher{client: client}, nil
}

func (p *KafkaPublisher) Publish(ctx context.Context, messages ...shared.Message) error {
	if len(messages) == 0 {
		return nil
	}

	records := make([]*kgo.Record, 0, len(messages))
	for _, message := range messages {
		records = append(records, toRecord(message))
	}

	if err := p.client.ProduceSync(ctx, records...).FirstErr(); err != nil {
		return fmt.Errorf("producing to kafka: %w", err)
	}
	return nil
}

func (p *KafkaPublisher) Close() {
	p.client.Close()
}

func toRecord(message shared.Message) *kgo.Record {
	headers := make([]kgo.RecordHeader, 0, len(message.Headers))
	for key, value := range message.Headers {
		headers = append(headers, kgo.RecordHeader{Key: key, Value: []byte(value)})
	}

	return &kgo.Record{
		Topic:   message.Topic,
		Key:     []byte(message.Key),
		Value:   message.Payload,
		Headers: headers,
	}
}
