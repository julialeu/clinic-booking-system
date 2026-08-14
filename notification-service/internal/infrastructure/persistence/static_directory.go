package persistence

import (
	"context"

	"github.com/julialeu/clinic-booking-system/notification-service/internal/domain/notification"
)

var _ notification.PatientDirectory = (*StaticPatientDirectory)(nil)

// StaticPatientDirectory devuelve datos de prueba mientras el servicio
// de Patient Management no está disponible.
type StaticPatientDirectory struct{}

func NewStaticPatientDirectory() *StaticPatientDirectory {
	return &StaticPatientDirectory{}
}

func (d *StaticPatientDirectory) Lookup(
	_ context.Context,
	patientId string,
) (notification.PatientContact, error) {
	return notification.PatientContact{
		Name:  "Paciente " + patientId[:8],
		Phone: "+34600000000",
	}, nil
}
