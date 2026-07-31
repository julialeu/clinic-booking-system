package appointment

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AppointmentId struct {
	value string
}

var ErrInvalidAppointmentId = errors.New("invalid appointment id: must be a valid UUID")

func NewAppointmentId() AppointmentId {
	return AppointmentId{value: uuid.NewString()}
}

// string existente
// Valida que sea un UUID correcto
func AppointmentIdFromString(value string) (AppointmentId, error) {
	if _, err := uuid.Parse(value); err != nil {
		return AppointmentId{}, ErrInvalidAppointmentId
	}
	return AppointmentId{value: value}, nil
}

func (a AppointmentId) Value() string {
	return a.value
}

func (a AppointmentId) Equals(other AppointmentId) bool {
	return a.value == other.value
}

type TimeSlot struct {
	start time.Time
	end   time.Time
}

var ErrInvalidSlot = errors.New("invalid slot: end must be after start")

func NewTimeSlot(start, end time.Time) (TimeSlot, error) {
	if !end.After(start) {
		return TimeSlot{}, ErrInvalidSlot
	}
	return TimeSlot{start: start, end: end}, nil
}

func (a TimeSlot) Start() time.Time {
	return a.start
}

func (a TimeSlot) End() time.Time {
	return a.end
}

func (a TimeSlot) IsPast(now time.Time) bool {
	return a.end.Before(now)
}

func (a TimeSlot) Equals(other TimeSlot) bool {
	return a.start.Equal(other.start) && a.end.Equal(other.end)
}

type Money struct {
	amountInCents int
	currency      string
}

var ErrNegativeAmount = errors.New("invalid money: amount cannot be negative")

func NewMoney(amountInCents int, currency string) (Money, error) {
	if amountInCents < 0 {
		return Money{}, ErrNegativeAmount
	}
	return Money{amountInCents: amountInCents, currency: currency}, nil
}

func (m Money) AmountInCents() int {
	return m.amountInCents
}

func (m Money) Currency() string {
	return m.currency
}

func (m Money) Equals(other Money) bool {
	return m.amountInCents == other.amountInCents && m.currency == other.currency
}

// AppointmentStatus representa el estado del ciclo de vida de una cita.
type AppointmentStatus int

const (
	// StatusReserved: bloqueo temporal. El paciente ha seleccionado
	// la franja pero aún no ha confirmado. Expira a los 5 minutos.
	StatusReserved AppointmentStatus = iota

	StatusConfirmed

	StatusCancelled

	StatusCompleted
)

func (s AppointmentStatus) String() string {
	switch s {
	case StatusReserved:
		return "reserved"
	case StatusConfirmed:
		return "confirmed"
	case StatusCancelled:
		return "cancelled"
	case StatusCompleted:
		return "completed"
	default:
		return "unknown"
	}
}

// PatientId identifica al paciente que reserva la cita.
type PatientId struct {
	value string
}

var ErrInvalidPatientId = errors.New("invalid patient id: must be a valid UUID")

func NewPatientId(value string) (PatientId, error) {
	if _, err := uuid.Parse(value); err != nil {
		return PatientId{}, ErrInvalidPatientId
	}
	return PatientId{value: value}, nil
}

func (p PatientId) Value() string {
	return p.value
}

func (p PatientId) Equals(other PatientId) bool {
	return p.value == other.value
}

// ============================================================
// AppointmentType — Value Object que define un tipo de cita
// (primera visita, seguimiento, urgencia...) con su duración,
// color y precio. Configurable según requisitos del cliente.
// ============================================================

type AppointmentType struct {
	name             string
	expectedDuration time.Duration
	color            string
	price            Money
}

var (
	ErrEmptyTypeName       = errors.New("invalid appointment type: name cannot be empty")
	ErrNonPositiveDuration = errors.New("invalid appointment type: duration must be positive")
)

func NewAppointmentType(
	name string,
	expectedDuration time.Duration,
	color string,
	price Money,
) (AppointmentType, error) {
	if name == "" {
		return AppointmentType{}, ErrEmptyTypeName
	}
	if expectedDuration <= 0 {
		return AppointmentType{}, ErrNonPositiveDuration
	}
	return AppointmentType{
		name:             name,
		expectedDuration: expectedDuration,
		color:            color,
		price:            price,
	}, nil
}

func (at AppointmentType) Name() string {
	return at.name
}

func (at AppointmentType) ExpectedDuration() time.Duration {
	return at.expectedDuration
}

func (at AppointmentType) Color() string {
	return at.color
}

func (at AppointmentType) Price() Money {
	return at.price
}

// Duration devuelve cuánto dura la franja (ej: 30 * time.Minute).
// Es información derivada: no se almacena, se calcula.
func (a TimeSlot) Duration() time.Duration {
	return a.end.Sub(a.start)
}

// Overlaps indica si esta franja se solapa con otra.

func (a TimeSlot) Overlaps(other TimeSlot) bool {
	return a.start.Before(other.end) && other.start.Before(a.end)
}
