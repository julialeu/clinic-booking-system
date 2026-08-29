package command

import (
	"context"
	"fmt"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/appointment"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/shared"
)

type ExpireReservationsHandler struct {
	repository  appointment.Repository
	outbox      shared.OutboxRepository
	transaction shared.TransactionManager
	clock       Clock
}

func NewExpireReservationsHandler(
	repository appointment.Repository,
	outbox shared.OutboxRepository,
	transaction shared.TransactionManager,
	clock Clock,
) *ExpireReservationsHandler {
	return &ExpireReservationsHandler{
		repository:  repository,
		outbox:      outbox,
		transaction: transaction,
		clock:       clock,
	}
}

// Handle libera las franjas cuyo bloqueo temporal ha vencido y
// devuelve cuántas reservas expiraron.
func (h *ExpireReservationsHandler) Handle(ctx context.Context) (int, error) {
	now := h.clock.Now()
	expired := 0

	err := h.transaction.WithinTransaction(ctx, func(txCtx context.Context) error {
		lapsed, err := h.repository.FindExpiredReservations(txCtx, now)
		if err != nil {
			return fmt.Errorf("finding expired reservations: %w", err)
		}

		for _, reservation := range lapsed {
			if err := reservation.Expire(now); err != nil {
				continue
			}

			if err := h.repository.Save(txCtx, reservation); err != nil {
				return fmt.Errorf("saving expired reservation: %w", err)
			}

			if err := recordEvents(txCtx, h.outbox, reservation); err != nil {
				return err
			}

			expired++
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return expired, nil
}
