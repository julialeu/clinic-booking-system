package command

import (
	"context"
	"fmt"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/appointment"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/shared"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/infrastructure/messaging"
)

// recordEvents persiste en el outbox los eventos emitidos por el
// aggregate. Debe llamarse dentro de la misma transacción que lo guardó.
func recordEvents(
	ctx context.Context,
	outbox shared.OutboxRepository,
	a *appointment.Appointment,
) error {
	events, err := messaging.ToOutboxEvents(a.PullDomainEvents())
	if err != nil {
		return fmt.Errorf("translating domain events: %w", err)
	}

	if err := outbox.Save(ctx, events...); err != nil {
		return fmt.Errorf("saving outbox events: %w", err)
	}
	return nil
}
