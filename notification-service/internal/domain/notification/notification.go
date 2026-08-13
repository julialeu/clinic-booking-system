package notification

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrEmptyRecipient = errors.New("notification: recipient cannot be empty")
	ErrEmptyBody      = errors.New("notification: body cannot be empty")
)

type Channel string

const (
	ChannelWhatsApp Channel = "whatsapp"
	ChannelSMS      Channel = "sms"
	ChannelEmail    Channel = "email"
)

type Notification struct {
	recipient string
	channel   Channel
	body      string
}

func New(recipient string, channel Channel, body string) (Notification, error) {
	if recipient == "" {
		return Notification{}, ErrEmptyRecipient
	}
	if body == "" {
		return Notification{}, ErrEmptyBody
	}
	return Notification{recipient: recipient, channel: channel, body: body}, nil
}

func (n Notification) Recipient() string { return n.recipient }
func (n Notification) Channel() Channel  { return n.channel }
func (n Notification) Body() string      { return n.body }

// AppointmentDetails son los datos que necesita una plantilla para
// componer el mensaje de una cita.
type AppointmentDetails struct {
	PatientName string
	StartsAt    time.Time
	TypeName    string
}

// ComposeConfirmation genera el mensaje de confirmación de una cita.
func ComposeConfirmation(details AppointmentDetails) string {
	return fmt.Sprintf(
		"Hola %s, tu cita de %s está confirmada para el %s. Si necesitas cambiarla, responde a este mensaje.",
		details.PatientName,
		details.TypeName,
		formatSpanish(details.StartsAt),
	)
}

// ComposeReminder genera el recordatorio previo a la cita.
func ComposeReminder(details AppointmentDetails) string {
	return fmt.Sprintf(
		"Hola %s, te recordamos tu cita de %s el %s. ¡Te esperamos!",
		details.PatientName,
		details.TypeName,
		formatSpanish(details.StartsAt),
	)
}

// ComposeCancellation genera el aviso de cancelación.
func ComposeCancellation(details AppointmentDetails) string {
	return fmt.Sprintf(
		"Hola %s, tu cita del %s ha sido cancelada. Puedes reservar otra cuando quieras.",
		details.PatientName,
		formatSpanish(details.StartsAt),
	)
}

var weekdays = [...]string{
	"domingo", "lunes", "martes", "miércoles", "jueves", "viernes", "sábado",
}

var months = [...]string{
	"", "enero", "febrero", "marzo", "abril", "mayo", "junio",
	"julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre",
}

func formatSpanish(t time.Time) string {
	return fmt.Sprintf("%s %d de %s a las %02d:%02d",
		weekdays[t.Weekday()],
		t.Day(),
		months[t.Month()],
		t.Hour(),
		t.Minute(),
	)
}
