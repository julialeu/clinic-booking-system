package notification

import "time"

const (
	EventPatientRegistered     = "patient.registered"
	EventPatientContactChanged = "patient.contact_changed"
)

type PatientEvent struct {
	PatientId  string    `json:"patientId"`
	FirstName  string    `json:"firstName"`
	FullName   string    `json:"fullName"`
	Phone      string    `json:"phone"`
	OccurredOn time.Time `json:"occurredOn"`
}
