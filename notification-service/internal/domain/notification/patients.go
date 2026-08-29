package notification

import (
	"context"
	"errors"
)

var ErrPatientNotFound = errors.New("notification: patient not found")

type PatientContact struct {
	Name  string
	Phone string
}

type PatientDirectory interface {
	Lookup(ctx context.Context, patientId string) (PatientContact, error)
}

type PatientProjection interface {
	Upsert(ctx context.Context, contact ProjectedPatient) error
}

type ProjectedPatient struct {
	PatientId string
	FirstName string
	FullName  string
	Phone     string
}
