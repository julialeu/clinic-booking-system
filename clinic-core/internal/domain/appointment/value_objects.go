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

type AppointmentSlot struct {
	start time.Time
	end   time.Time
}

var ErrInvalidSlot = errors.New("invalid slot: end must be after start")

func NewAppointmentSlot(start, end time.Time) (AppointmentSlot, error) {
	if !end.After(start) {
		return AppointmentSlot{}, ErrInvalidSlot
	}
	return AppointmentSlot{start: start, end: end}, nil
}

func (a AppointmentSlot) Start() time.Time {
	return a.start
}

func (a AppointmentSlot) End() time.Time {
	return a.end
}

func (a AppointmentSlot) IsPast(now time.Time) bool {
	return a.end.Before(now)
}

func (a AppointmentSlot) Equals(other AppointmentSlot) bool {
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
