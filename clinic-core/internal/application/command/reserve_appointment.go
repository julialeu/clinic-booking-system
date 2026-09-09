package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/appointment"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/shared"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/platform/telemetry"
)

var ErrSlotNotAvailable = errors.New("reserve appointment: the requested slot is not available")

var tracer = otel.Tracer("clinic-core/application/command")

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
	ctx, span := tracer.Start(ctx, "ReserveAppointment",
		trace.WithAttributes(
			attribute.String("patient.id", cmd.PatientId),
			attribute.String("appointment.type", cmd.TypeName),
			attribute.String("slot.starts_at", cmd.StartsAt.Format(time.RFC3339)),
		),
	)
	defer span.End()

	started := time.Now()
	var reserved *appointment.Appointment

	patientId, err := appointment.NewPatientId(cmd.PatientId)
	if err != nil {
		return failReservation(span, started, "invalid_patient", err)
	}

	slot, err := appointment.NewTimeSlot(cmd.StartsAt, cmd.StartsAt.Add(cmd.TypeDuration))
	if err != nil {
		return failReservation(span, started, "invalid_slot", err)
	}

	price, err := appointment.NewMoney(cmd.PriceCents, cmd.PriceCurrency)
	if err != nil {
		return failReservation(span, started, "invalid_price", err)
	}

	appointmentType, err := appointment.NewAppointmentType(
		cmd.TypeName,
		cmd.TypeDuration,
		cmd.TypeColor,
		price,
	)
	if err != nil {
		return failReservation(span, started, "invalid_type", err)
	}

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
		reason := "internal"
		if errors.Is(err, ErrSlotNotAvailable) {
			reason = "slot_taken"
		}
		return failReservation(span, started, reason, err)
	}

	span.SetAttributes(attribute.String("appointment.id", reserved.Id().Value()))
	telemetry.AppointmentsReserved.Inc()
	telemetry.CommandDuration.
		WithLabelValues("reserve_appointment", "success").
		Observe(time.Since(started).Seconds())

	return reserved.Id(), nil
}

func failReservation(
	span trace.Span,
	started time.Time,
	reason string,
	err error,
) (appointment.AppointmentId, error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, reason)

	telemetry.AppointmentsRejected.WithLabelValues(reason).Inc()
	telemetry.CommandDuration.
		WithLabelValues("reserve_appointment", "failure").
		Observe(time.Since(started).Seconds())

	return appointment.AppointmentId{}, err
}
