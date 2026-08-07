package command

import (
	"context"
	"fmt"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/appointment"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/shared"
)

type CancelAppointment struct {
	AppointmentId string
}

type CancelAppointmentHandler struct {
	repository  appointment.Repository
	outbox      shared.OutboxRepository
	transaction shared.TransactionManager
	clock       Clock
}

func NewCancelAppointmentHandler(
	repository appointment.Repository,
	outbox shared.OutboxRepository,
	transaction shared.TransactionManager,
	clock Clock,
) *CancelAppointmentHandler {
	return &CancelAppointmentHandler{
		repository:  repository,
		outbox:      outbox,
		transaction: transaction,
		clock:       clock,
	}
}

func (h *CancelAppointmentHandler) Handle(ctx context.Context, cmd CancelAppointment) error {
	id, err := appointment.AppointmentIdFromString(cmd.AppointmentId)
	if err != nil {
		return err
	}

	return h.transaction.WithinTransaction(ctx, func(txCtx context.Context) error {
		found, err := h.repository.FindById(txCtx, id)
		if err != nil {
			return err
		}

		if err := found.Cancel(h.clock.Now()); err != nil {
			return err
		}

		if err := h.repository.Save(txCtx, found); err != nil {
			return fmt.Errorf("saving cancelled appointment: %w", err)
		}

		return recordEvents(txCtx, h.outbox, found)
	})
}
