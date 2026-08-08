package messaging

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/shared"
)

const (
	appointmentsTopic = "clinic.appointments"
	defaultBatchSize  = 100
	defaultInterval   = 2 * time.Second
)

type RelayConfig struct {
	BatchSize    int
	PollInterval time.Duration
}

func DefaultRelayConfig() RelayConfig {
	return RelayConfig{
		BatchSize:    defaultBatchSize,
		PollInterval: defaultInterval,
	}
}

type OutboxRelay struct {
	pool      *pgxpool.Pool
	publisher shared.EventPublisher
	config    RelayConfig
}

func NewOutboxRelay(
	pool *pgxpool.Pool,
	publisher shared.EventPublisher,
	config RelayConfig,
) *OutboxRelay {
	return &OutboxRelay{pool: pool, publisher: publisher, config: config}
}

func (r *OutboxRelay) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			log.Println("outbox relay: polling")
			published, err := r.processBatch(ctx)
			if err != nil {
				log.Printf("outbox relay: %v", err)
				continue
			}
			if published > 0 {
				log.Printf("outbox relay: published %d events", published)
			}
		}
	}
}

type pendingEvent struct {
	id            int64
	aggregateType string
	aggregateId   string
	eventType     string
	payload       []byte
	occurredOn    time.Time
}

const selectPendingSQL = `
SELECT id, aggregate_type, aggregate_id, event_type, payload, occurred_on
FROM outbox_events
WHERE published_at IS NULL
ORDER BY id
LIMIT $1
FOR UPDATE SKIP LOCKED`

const markPublishedSQL = `
UPDATE outbox_events
SET published_at = now()
WHERE id = ANY($1)`

const recordFailureSQL = `
UPDATE outbox_events
SET attempts = attempts + 1, last_error = $2
WHERE id = ANY($1)`

func (r *OutboxRelay) processBatch(ctx context.Context) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	events, err := fetchPending(ctx, tx, r.config.BatchSize)
	if err != nil {
		return 0, err
	}
	if len(events) == 0 {
		return 0, nil
	}

	messages := make([]shared.Message, 0, len(events))
	ids := make([]int64, 0, len(events))

	for _, event := range events {
		messages = append(messages, shared.Message{
			Topic:   appointmentsTopic,
			Key:     event.aggregateId,
			Payload: event.payload,
			Headers: map[string]string{
				"event_type":     event.eventType,
				"aggregate_type": event.aggregateType,
				"occurred_on":    event.occurredOn.Format(time.RFC3339),
			},
		})
		ids = append(ids, event.id)
	}

	// Sin timeout, un broker que no responde cuelga el relay indefinidamente.
	publishCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := r.publisher.Publish(publishCtx, messages...); err != nil {
		// El fallo se registra en su propia transacción, porque esta
		// va a revertirse para liberar los bloqueos.
		r.recordFailure(ctx, ids, err)
		return 0, fmt.Errorf("publishing batch: %w", err)
	}

	if _, err := tx.Exec(ctx, markPublishedSQL, ids); err != nil {
		return 0, fmt.Errorf("marking events as published: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("committing batch: %w", err)
	}

	return len(events), nil
}

func fetchPending(ctx context.Context, tx pgx.Tx, limit int) ([]pendingEvent, error) {
	rows, err := tx.Query(ctx, selectPendingSQL, limit)
	if err != nil {
		return nil, fmt.Errorf("fetching pending events: %w", err)
	}
	defer rows.Close()

	events := make([]pendingEvent, 0, limit)
	for rows.Next() {
		var event pendingEvent
		err := rows.Scan(
			&event.id,
			&event.aggregateType,
			&event.aggregateId,
			&event.eventType,
			&event.payload,
			&event.occurredOn,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning pending event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating pending events: %w", err)
	}
	return events, nil
}

func (r *OutboxRelay) recordFailure(ctx context.Context, ids []int64, cause error) {
	if _, err := r.pool.Exec(ctx, recordFailureSQL, ids, cause.Error()); err != nil {
		log.Printf("outbox relay: could not record failure: %v", err)
	}
}
