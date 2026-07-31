package appointment

import "time"

type AppointmentReserved struct {
	AppointmentId string
	PatientId     string
	StartsAt      time.Time
	EndsAt        time.Time
	ReservedUntil time.Time
	OccurredOn    time.Time
}

type AppointmentConfirmed struct {
	AppointmentId string
	PatientId     string
	StartsAt      time.Time
	EndsAt        time.Time
	OccurredOn    time.Time
}

type AppointmentCancelled struct {
	AppointmentId string
	PatientId     string
	StartsAt      time.Time
	OccurredOn    time.Time
}

type AppointmentCompleted struct {
	AppointmentId string
	PatientId     string
	OccurredOn    time.Time
}

type AppointmentExpired struct {
	AppointmentId string
	PatientId     string
	OccurredOn    time.Time
}
