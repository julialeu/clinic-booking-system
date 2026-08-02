package appointment

import (
	"errors"
	"time"
)

// reservationWindow es el tiempo que una cita permanece bloqueada
// tras seleccionarla, antes de expirar si no se confirma.
const reservationWindow = 5 * time.Minute

var (
	ErrDurationMismatch  = errors.New("appointment: slot duration does not match the appointment type")
	ErrPastAppointment   = errors.New("appointment: cannot book a slot in the past")
	ErrNotReserved       = errors.New("appointment: only reserved appointments can be confirmed")
	ErrReservationLapsed = errors.New("appointment: the reservation window has expired")
	ErrAlreadyCancelled  = errors.New("appointment: appointment is already cancelled")
	ErrAlreadyCompleted  = errors.New("appointment: appointment is already completed")
	ErrAlreadyStarted    = errors.New("appointment: cannot cancel an appointment that already started")
	ErrNotConfirmed      = errors.New("appointment: only confirmed appointments can be completed")
	ErrNotFinished       = errors.New("appointment: cannot complete an appointment before it ends")
)

// Appointment es el Aggregate Root del dominio de citas.
// Protege las invariantes del ciclo de vida y registra los
// eventos de dominio que va produciendo.
type Appointment struct {
	id              AppointmentId
	patientId       PatientId
	slot            TimeSlot
	appointmentType AppointmentType
	status          AppointmentStatus
	reservedUntil   time.Time
	domainEvents    []any
}

// Reserve crea una cita nueva en estado reservado (bloqueo temporal).
func Reserve(
	patientId PatientId,
	slot TimeSlot,
	appointmentType AppointmentType,
	now time.Time,
) (*Appointment, error) {
	if slot.Start().Before(now) {
		return nil, ErrPastAppointment
	}
	if slot.Duration() != appointmentType.ExpectedDuration() {
		return nil, ErrDurationMismatch
	}

	appointment := &Appointment{
		id:              NewAppointmentId(),
		patientId:       patientId,
		slot:            slot,
		appointmentType: appointmentType,
		status:          StatusReserved,
		reservedUntil:   now.Add(reservationWindow),
		domainEvents:    []any{},
	}

	appointment.recordEvent(AppointmentReserved{
		AppointmentId: appointment.id.Value(),
		PatientId:     patientId.Value(),
		StartsAt:      slot.Start(),
		EndsAt:        slot.End(),
		ReservedUntil: appointment.reservedUntil,
		OccurredOn:    now,
	})

	return appointment, nil
}

// Reconstitute rehidrata un aggregate desde persistencia.
// No valida reglas de creación ni emite eventos: reconstruye
// un estado que ya existió.
func Reconstitute(
	id AppointmentId,
	patientId PatientId,
	slot TimeSlot,
	appointmentType AppointmentType,
	status AppointmentStatus,
	reservedUntil time.Time,
) *Appointment {
	return &Appointment{
		id:              id,
		patientId:       patientId,
		slot:            slot,
		appointmentType: appointmentType,
		status:          status,
		reservedUntil:   reservedUntil,
		domainEvents:    []any{},
	}
}

// === Transiciones de estado ===

func (a *Appointment) Confirm(now time.Time) error {
	if a.status != StatusReserved {
		return ErrNotReserved
	}
	if now.After(a.reservedUntil) {
		return ErrReservationLapsed
	}

	a.status = StatusConfirmed
	a.recordEvent(AppointmentConfirmed{
		AppointmentId: a.id.Value(),
		PatientId:     a.patientId.Value(),
		StartsAt:      a.slot.Start(),
		EndsAt:        a.slot.End(),
		OccurredOn:    now,
	})
	return nil
}

func (a *Appointment) Cancel(now time.Time) error {
	if a.status == StatusCancelled {
		return ErrAlreadyCancelled
	}
	if a.status == StatusCompleted {
		return ErrAlreadyCompleted
	}
	// No se cancela una cita cuya franja ya empezó.
	if !a.slot.Start().After(now) {
		return ErrAlreadyStarted
	}

	a.status = StatusCancelled
	a.recordEvent(AppointmentCancelled{
		AppointmentId: a.id.Value(),
		PatientId:     a.patientId.Value(),
		StartsAt:      a.slot.Start(),
		OccurredOn:    now,
	})
	return nil
}

func (a *Appointment) Complete(now time.Time) error {
	if a.status != StatusConfirmed {
		return ErrNotConfirmed
	}
	// Una sesión solo puede darse por realizada una vez terminada.
	if now.Before(a.slot.End()) {
		return ErrNotFinished
	}

	a.status = StatusCompleted
	a.recordEvent(AppointmentCompleted{
		AppointmentId: a.id.Value(),
		PatientId:     a.patientId.Value(),
		OccurredOn:    now,
	})
	return nil
}

func (a *Appointment) Expire(now time.Time) error {
	if !a.IsReservationExpired(now) {
		return ErrNotReserved
	}

	a.status = StatusCancelled
	a.recordEvent(AppointmentExpired{
		AppointmentId: a.id.Value(),
		PatientId:     a.patientId.Value(),
		OccurredOn:    now,
	})
	return nil
}

// === Consultas ===

func (a *Appointment) Id() AppointmentId {
	return a.id
}

func (a *Appointment) PatientId() PatientId {
	return a.patientId
}

func (a *Appointment) Slot() TimeSlot {
	return a.slot
}

func (a *Appointment) Type() AppointmentType {
	return a.appointmentType
}

func (a *Appointment) Status() AppointmentStatus {
	return a.status
}

func (a *Appointment) ReservedUntil() time.Time {
	return a.reservedUntil
}

func (a *Appointment) IsReservationExpired(now time.Time) bool {
	return a.status == StatusReserved && now.After(a.reservedUntil)
}

// === Eventos de dominio ===

func (a *Appointment) recordEvent(event any) {
	a.domainEvents = append(a.domainEvents, event)
}

// PullDomainEvents devuelve los eventos acumulados y vacía la lista.
func (a *Appointment) PullDomainEvents() []any {
	events := a.domainEvents
	a.domainEvents = []any{}
	return events
}
