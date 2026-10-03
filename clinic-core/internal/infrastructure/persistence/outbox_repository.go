package persistence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/shared"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/platform/postgres"
)

var _ shared.OutboxRepository = (*OutboxRepository)(nil)

type OutboxRepository struct {
	pool *pgxpool.Pool
}

func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{pool: pool}
}

const insertOutboxSQL = `
INSERT INTO outbox_events (
    aggregate_type, aggregate_id, event_type, payload, occurred_on, trace_context
) VALUES ($1, $2, $3, $4, $5, $6)`

func (r *OutboxRepository) Save(ctx context.Context, events ...shared.OutboxEvent) error {
	if len(events) == 0 {
		return nil
	}

	querier := postgres.QuerierFrom(ctx, r.pool)

	for _, event := range events {
		traceContext, err := json.Marshal(event.TraceContext)
		if err != nil {
			return fmt.Errorf("marshalling trace context: %w", err)
		}

		_, err = querier.Exec(ctx, insertOutboxSQL,
			event.AggregateType,
			event.AggregateId,
			event.EventType,
			event.Payload,
			event.OccurredOn,
			traceContext,
		)
		if err != nil {
			return fmt.Errorf("saving outbox event %s: %w", event.EventType, err)
		}
	}
	return nil
}
