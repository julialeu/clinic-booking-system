package appointment

import (
	"context"
	"errors"
	"time"
)

var ErrAppointmentNotFound = errors.New("appointment: not found")

type Repository interface {
	Save(ctx context.Context, appointment *Appointment) error

	FindById(ctx context.Context, id AppointmentId) (*Appointment, error)

	// FindOverlapping devuelve las citas activas cuya franja se solapa
	// con la indicada. Base de la prevención de dobles reservas.
	FindOverlapping(ctx context.Context, slot TimeSlot) ([]*Appointment, error)

	// FindExpiredReservations devuelve las reservas cuyo bloqueo temporal
	// ha vencido, para que el proceso de expiración las libere.
	FindExpiredReservations(ctx context.Context, now time.Time) ([]*Appointment, error)
}
