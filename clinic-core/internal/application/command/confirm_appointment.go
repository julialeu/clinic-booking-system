package command

import (
	"context"
	"fmt"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/appointment"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/shared"
)

type ConfirmAppointment struct {
	AppointmentId string
}

type ConfirmAppointmentHandler struct {
	repository  appointment.Repository
	transaction shared.TransactionManager
	clock       Clock
}

func NewConfirmAppointmentHandler(
	repository appointment.Repository,
	transaction shared.TransactionManager,
	clock Clock,
) *ConfirmAppointmentHandler {
	return &ConfirmAppointmentHandler{
		repository:  repository,
		transaction: transaction,
		clock:       clock,
	}
}

func (h *ConfirmAppointmentHandler) Handle(ctx context.Context, cmd ConfirmAppointment) error {
	id, err := appointment.AppointmentIdFromString(cmd.AppointmentId)
	if err != nil {
		return err
	}

	return h.transaction.WithinTransaction(ctx, func(txCtx context.Context) error {
		found, err := h.repository.FindById(txCtx, id)
		if err != nil {
			return err
		}

		if err := found.Confirm(h.clock.Now()); err != nil {
			return err
		}

		if err := h.repository.Save(txCtx, found); err != nil {
			return fmt.Errorf("saving confirmed appointment: %w", err)
		}
		return nil
	})
}
