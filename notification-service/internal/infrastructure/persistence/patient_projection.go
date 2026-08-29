package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/julialeu/clinic-booking-system/notification-service/internal/domain/notification"
)

var (
	_ notification.PatientProjection = (*PatientRepository)(nil)
	_ notification.PatientDirectory  = (*PatientRepository)(nil)
)

// PatientRepository implementa a la vez la escritura de la proyección
// y su lectura, que son las dos caras del mismo dato.
type PatientRepository struct {
	pool *pgxpool.Pool
}

func NewPatientRepository(pool *pgxpool.Pool) *PatientRepository {
	return &PatientRepository{pool: pool}
}

const upsertPatientSQL = `
INSERT INTO patients (id, first_name, full_name, phone)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET
    first_name = EXCLUDED.first_name,
    full_name  = EXCLUDED.full_name,
    phone      = EXCLUDED.phone,
    updated_at = now()`

func (r *PatientRepository) Upsert(
	ctx context.Context,
	contact notification.ProjectedPatient,
) error {
	_, err := r.pool.Exec(ctx, upsertPatientSQL,
		contact.PatientId,
		contact.FirstName,
		contact.FullName,
		contact.Phone,
	)
	if err != nil {
		return fmt.Errorf("upserting patient projection: %w", err)
	}
	return nil
}

func (r *PatientRepository) Lookup(
	ctx context.Context,
	patientId string,
) (notification.PatientContact, error) {
	row := r.pool.QueryRow(ctx,
		"SELECT first_name, phone FROM patients WHERE id = $1",
		patientId,
	)

	var contact notification.PatientContact
	err := row.Scan(&contact.Name, &contact.Phone)

	if errors.Is(err, pgx.ErrNoRows) {
		return notification.PatientContact{}, notification.ErrPatientNotFound
	}
	if err != nil {
		return notification.PatientContact{}, fmt.Errorf("looking up patient: %w", err)
	}

	return contact, nil
}
