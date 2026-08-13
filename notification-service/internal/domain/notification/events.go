package notification

import "time"

const (
	EventAppointmentReserved  = "appointment.reserved"
	EventAppointmentConfirmed = "appointment.confirmed"
	EventAppointmentCancelled = "appointment.cancelled"
	EventAppointmentCompleted = "appointment.completed"
	EventAppointmentExpired   = "appointment.expired"
)

type AppointmentEvent struct {
	AppointmentId string    `json:"AppointmentId"`
	PatientId     string    `json:"PatientId"`
	StartsAt      time.Time `json:"StartsAt"`
	EndsAt        time.Time `json:"EndsAt"`
	OccurredOn    time.Time `json:"OccurredOn"`
}
