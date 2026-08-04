package query

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WeeklyAgenda struct {
	AnyDayOfWeek time.Time
}

type AgendaAppointment struct {
	AppointmentId string
	PatientId     string
	StartsAt      time.Time
	EndsAt        time.Time
	TypeName      string
	TypeColor     string
	Status        string
}

type AgendaDay struct {
	Date         time.Time
	Appointments []AgendaAppointment
}

type WeeklyAgendaResult struct {
	WeekStart time.Time
	WeekEnd   time.Time
	Days      []AgendaDay
}

type WeeklyAgendaHandler struct {
	pool *pgxpool.Pool
}

func NewWeeklyAgendaHandler(pool *pgxpool.Pool) *WeeklyAgendaHandler {
	return &WeeklyAgendaHandler{pool: pool}
}

const weeklyAgendaSQL = `
SELECT id, patient_id, starts_at, ends_at, type_name, type_color, status
FROM appointments
WHERE starts_at >= $1
  AND starts_at <  $2
  AND status IN (0, 1)
ORDER BY starts_at`

func (h *WeeklyAgendaHandler) Handle(
	ctx context.Context,
	q WeeklyAgenda,
) (WeeklyAgendaResult, error) {
	weekStart := startOfWeek(q.AnyDayOfWeek)
	weekEnd := weekStart.AddDate(0, 0, 7)

	rows, err := h.pool.Query(ctx, weeklyAgendaSQL, weekStart, weekEnd)
	if err != nil {
		return WeeklyAgendaResult{}, fmt.Errorf("querying weekly agenda: %w", err)
	}
	defer rows.Close()

	byDay := make(map[time.Time][]AgendaAppointment)

	for rows.Next() {
		var item AgendaAppointment
		var status int16

		err := rows.Scan(
			&item.AppointmentId,
			&item.PatientId,
			&item.StartsAt,
			&item.EndsAt,
			&item.TypeName,
			&item.TypeColor,
			&status,
		)
		if err != nil {
			return WeeklyAgendaResult{}, fmt.Errorf("scanning agenda row: %w", err)
		}

		item.Status = statusLabel(status)
		day := truncateToDay(item.StartsAt)
		byDay[day] = append(byDay[day], item)
	}

	if err := rows.Err(); err != nil {
		return WeeklyAgendaResult{}, fmt.Errorf("iterating agenda rows: %w", err)
	}

	return WeeklyAgendaResult{
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Days:      buildDays(weekStart, byDay),
	}, nil
}

// startOfWeek devuelve el lunes de la semana a la que pertenece la fecha.
func startOfWeek(day time.Time) time.Time {
	day = truncateToDay(day)

	offset := int(day.Weekday()) - int(time.Monday)
	if offset < 0 {
		offset += 7 // domingo pertenece a la semana que empezó el lunes anterior
	}
	return day.AddDate(0, 0, -offset)
}

func truncateToDay(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

// buildDays devuelve siempre los 7 días, incluidos los vacíos,
// para que el cliente pueda pintar la rejilla sin comprobaciones.
func buildDays(weekStart time.Time, byDay map[time.Time][]AgendaAppointment) []AgendaDay {
	days := make([]AgendaDay, 0, 7)

	for i := 0; i < 7; i++ {
		date := weekStart.AddDate(0, 0, i)
		appointments := byDay[date]
		if appointments == nil {
			appointments = []AgendaAppointment{}
		}
		days = append(days, AgendaDay{Date: date, Appointments: appointments})
	}
	return days
}

func statusLabel(status int16) string {
	switch status {
	case 0:
		return "reserved"
	case 1:
		return "confirmed"
	case 2:
		return "cancelled"
	case 3:
		return "completed"
	default:
		return "unknown"
	}
}
