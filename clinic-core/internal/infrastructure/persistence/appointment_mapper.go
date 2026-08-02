package persistence

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/appointment"
)

type appointmentRow struct {
	id              string
	patientId       string
	startsAt        time.Time
	endsAt          time.Time
	status          int16
	reservedUntil   *time.Time
	typeName        string
	typeDurationMin int32
	typeColor       string
	priceCents      int32
	priceCurrency   string
}

type scannable interface {
	Scan(dest ...any) error
}

func scanAppointment(row scannable) (*appointment.Appointment, error) {
	var r appointmentRow

	err := row.Scan(
		&r.id, &r.patientId, &r.startsAt, &r.endsAt, &r.status, &r.reservedUntil,
		&r.typeName, &r.typeDurationMin, &r.typeColor, &r.priceCents, &r.priceCurrency,
	)
	if err != nil {
		return nil, err
	}

	return toDomain(r)
}

func collectAppointments(rows pgx.Rows) ([]*appointment.Appointment, error) {
	results := make([]*appointment.Appointment, 0)

	for rows.Next() {
		result, err := scanAppointment(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning appointment row: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating appointment rows: %w", err)
	}
	return results, nil
}

func toDomain(r appointmentRow) (*appointment.Appointment, error) {
	id, err := appointment.AppointmentIdFromString(r.id)
	if err != nil {
		return nil, fmt.Errorf("mapping appointment id: %w", err)
	}

	patientId, err := appointment.NewPatientId(r.patientId)
	if err != nil {
		return nil, fmt.Errorf("mapping patient id: %w", err)
	}

	slot, err := appointment.NewTimeSlot(r.startsAt, r.endsAt)
	if err != nil {
		return nil, fmt.Errorf("mapping time slot: %w", err)
	}

	price, err := appointment.NewMoney(int(r.priceCents), r.priceCurrency)
	if err != nil {
		return nil, fmt.Errorf("mapping price: %w", err)
	}

	appointmentType, err := appointment.NewAppointmentType(
		r.typeName,
		time.Duration(r.typeDurationMin)*time.Minute,
		r.typeColor,
		price,
	)
	if err != nil {
		return nil, fmt.Errorf("mapping appointment type: %w", err)
	}

	var reservedUntil time.Time
	if r.reservedUntil != nil {
		reservedUntil = *r.reservedUntil
	}

	return appointment.Reconstitute(
		id,
		patientId,
		slot,
		appointmentType,
		appointment.AppointmentStatus(r.status),
		reservedUntil,
	), nil
}

func nullableTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
