package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/appointment"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/shared"
)

var ErrSlotNotAvailable = errors.New("reserve appointment: the requested slot is not available")

type ReserveAppointment struct {
	PatientId     string
	StartsAt      time.Time
	TypeName      string
	TypeDuration  time.Duration
	TypeColor     string
	PriceCents    int
	PriceCurrency string
}

type Clock interface {
	Now() time.Time
}

type ReserveAppointmentHandler struct {
	repository  appointment.Repository
	outbox      shared.OutboxRepository
	transaction shared.TransactionManager
	clock       Clock
}

func NewReserveAppointmentHandler(
	repository appointment.Repository,
	outbox shared.OutboxRepository,
	transaction shared.TransactionManager,
	clock Clock,
) *ReserveAppointmentHandler {
	return &ReserveAppointmentHandler{
		repository:  repository,
		outbox:      outbox,
		transaction: transaction,
		clock:       clock,
	}
}

func (h *ReserveAppointmentHandler) Handle(
	ctx context.Context,
	cmd ReserveAppointment,
) (appointment.AppointmentId, error) {
	var reserved *appointment.Appointment

	patientId, err := appointment.NewPatientId(cmd.PatientId)
	if err != nil {
		return appointment.AppointmentId{}, err
	}

	slot, err := appointment.NewTimeSlot(cmd.StartsAt, cmd.StartsAt.Add(cmd.TypeDuration))
	if err != nil {
		return appointment.AppointmentId{}, err
	}

	price, err := appointment.NewMoney(cmd.PriceCents, cmd.PriceCurrency)
	if err != nil {
		return appointment.AppointmentId{}, err
	}

	appointmentType, err := appointment.NewAppointmentType(
		cmd.TypeName,
		cmd.TypeDuration,
		cmd.TypeColor,
		price,
	)
	if err != nil {
		return appointment.AppointmentId{}, err
	}

	// La comprobación de disponibilidad, el guardado y la escritura de
	// eventos ocurren en la misma transacción: FindOverlapping bloquea
	// las filas con FOR UPDATE y solo las libera al hacer commit.
	err = h.transaction.WithinTransaction(ctx, func(txCtx context.Context) error {
		overlapping, err := h.repository.FindOverlapping(txCtx, slot)
		if err != nil {
			return fmt.Errorf("checking slot availability: %w", err)
		}
		if len(overlapping) > 0 {
			return ErrSlotNotAvailable
		}

		reserved, err = appointment.Reserve(patientId, slot, appointmentType, h.clock.Now())
		if err != nil {
			return err
		}

		if err := h.repository.Save(txCtx, reserved); err != nil {
			return fmt.Errorf("saving appointment: %w", err)
		}

		return recordEvents(txCtx, h.outbox, reserved)
	})
	if err != nil {
		return appointment.AppointmentId{}, err
	}

	return reserved.Id(), nil
}
