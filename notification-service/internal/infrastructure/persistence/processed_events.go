package persistence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/julialeu/clinic-booking-system/notification-service/internal/domain/notification"
)

var _ notification.ProcessedEvents = (*ProcessedEventsRepository)(nil)

type ProcessedEventsRepository struct {
	pool *pgxpool.Pool
}

func NewProcessedEventsRepository(pool *pgxpool.Pool) *ProcessedEventsRepository {
	return &ProcessedEventsRepository{pool: pool}
}

const markProcessedSQL = `
INSERT INTO processed_events (topic, partition_id, offset_id, event_type)
VALUES ($1, $2, $3, $4)
ON CONFLICT (topic, partition_id, offset_id) DO NOTHING`

// MarkProcessed intenta registrar el evento. Si la fila ya existía,
// no se insertó nada y devolvemos false: es un duplicado.
func (r *ProcessedEventsRepository) MarkProcessed(
	ctx context.Context,
	reference notification.EventReference,
) (bool, error) {
	tag, err := r.pool.Exec(ctx, markProcessedSQL,
		reference.Topic,
		reference.Partition,
		reference.Offset,
		reference.EventType,
	)
	if err != nil {
		return false, fmt.Errorf("marking event as processed: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}
