package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/appointment"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/platform/postgres"
)

var _ appointment.Repository = (*AppointmentRepository)(nil)

type AppointmentRepository struct {
	pool *pgxpool.Pool
}

func NewAppointmentRepository(pool *pgxpool.Pool) *AppointmentRepository {
	return &AppointmentRepository{pool: pool}
}

const upsertAppointmentSQL = `
INSERT INTO appointments (
    id, patient_id, starts_at, ends_at, status, reserved_until,
    type_name, type_duration_min, type_color, price_cents, price_currency
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (id) DO UPDATE SET
    starts_at      = EXCLUDED.starts_at,
    ends_at        = EXCLUDED.ends_at,
    status         = EXCLUDED.status,
    reserved_until = EXCLUDED.reserved_until,
    updated_at     = now()`

func (r *AppointmentRepository) Save(ctx context.Context, a *appointment.Appointment) error {
	_, err := postgres.QuerierFrom(ctx, r.pool).Exec(ctx, upsertAppointmentSQL,
		a.Id().Value(),
		a.PatientId().Value(),
		a.Slot().Start(),
		a.Slot().End(),
		int16(a.Status()),
		nullableTime(a.ReservedUntil()),
		a.Type().Name(),
		int32(a.Type().ExpectedDuration()/time.Minute),
		a.Type().Color(),
		int32(a.Type().Price().AmountInCents()),
		a.Type().Price().Currency(),
	)
	if err != nil {
		return fmt.Errorf("saving appointment: %w", err)
	}
	return nil
}

const selectColumns = `
    id, patient_id, starts_at, ends_at, status, reserved_until,
    type_name, type_duration_min, type_color, price_cents, price_currency`

func (r *AppointmentRepository) FindById(
	ctx context.Context,
	id appointment.AppointmentId,
) (*appointment.Appointment, error) {
	query := `SELECT ` + selectColumns + ` FROM appointments WHERE id = $1`

	row := postgres.QuerierFrom(ctx, r.pool).QueryRow(ctx, query, id.Value())

	result, err := scanAppointment(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appointment.ErrAppointmentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("finding appointment: %w", err)
	}
	return result, nil
}

// FindOverlapping busca citas activas que pisen la franja indicada.
// Usa FOR UPDATE para bloquear las filas durante la transacción y
// evitar que dos reservas simultáneas pasen la comprobación.
func (r *AppointmentRepository) FindOverlapping(
	ctx context.Context,
	slot appointment.TimeSlot,
) ([]*appointment.Appointment, error) {
	query := `
SELECT ` + selectColumns + `
FROM appointments
WHERE status IN ($1, $2)
  AND starts_at < $3
  AND ends_at   > $4
FOR UPDATE`

	rows, err := postgres.QuerierFrom(ctx, r.pool).Query(ctx, query,
		int16(appointment.StatusReserved),
		int16(appointment.StatusConfirmed),
		slot.End(),
		slot.Start(),
	)
	if err != nil {
		return nil, fmt.Errorf("finding overlapping appointments: %w", err)
	}
	defer rows.Close()

	return collectAppointments(rows)
}

func (r *AppointmentRepository) FindExpiredReservations(
	ctx context.Context,
	now time.Time,
) ([]*appointment.Appointment, error) {
	query := `
SELECT ` + selectColumns + `
FROM appointments
WHERE status = $1
  AND reserved_until < $2`

	rows, err := postgres.QuerierFrom(ctx, r.pool).Query(ctx, query, int16(appointment.StatusReserved), now)
	if err != nil {
		return nil, fmt.Errorf("finding expired reservations: %w", err)
	}
	defer rows.Close()

	return collectAppointments(rows)
}
