package query_test

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/application/query"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("clinic_core_test"),
		tcpostgres.WithUsername("clinic"),
		tcpostgres.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("starting postgres container: %v", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("getting connection string: %v", err)
	}

	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("creating test pool: %v", err)
	}

	path := filepath.Join("..", "..", "..", "migrations", "000001_create_appointments_table.up.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("reading migration: %v", err)
	}
	if _, err := testPool.Exec(ctx, string(content)); err != nil {
		log.Fatalf("applying migration: %v", err)
	}

	code := m.Run()

	testPool.Close()
	if err := testcontainers.TerminateContainer(container); err != nil {
		log.Printf("terminating container: %v", err)
	}
	os.Exit(code)
}

func truncate(t *testing.T) {
	t.Helper()
	if _, err := testPool.Exec(context.Background(), "TRUNCATE appointments"); err != nil {
		t.Fatalf("truncating: %v", err)
	}
}

const insertSQL = `
INSERT INTO appointments (
    id, patient_id, starts_at, ends_at, status, reserved_until,
    type_name, type_duration_min, type_color, price_cents, price_currency
) VALUES ($1, $2, $3, $4, $5, NULL, 'Primera visita', 60, '#3B82F6', 4500, 'EUR')`

func insertAppointment(t *testing.T, id string, startsAt time.Time, status int16) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), insertSQL,
		id,
		"3f2504e0-4f89-11d3-9a0c-0305e82c3301",
		startsAt,
		startsAt.Add(time.Hour),
		status,
	)
	if err != nil {
		t.Fatalf("inserting appointment: %v", err)
	}
}

func totalAppointments(result query.WeeklyAgendaResult) int {
	total := 0
	for _, day := range result.Days {
		total += len(day.Appointments)
	}
	return total
}

func TestWeeklyAgendaAlwaysReturnsSevenDays(t *testing.T) {
	truncate(t)
	handler := query.NewWeeklyAgendaHandler(testPool)

	// Miércoles 2 de septiembre de 2026
	result, err := handler.Handle(context.Background(), query.WeeklyAgenda{
		AnyDayOfWeek: time.Date(2026, 9, 2, 15, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Days) != 7 {
		t.Errorf("expected 7 days, got %d", len(result.Days))
	}
	if result.WeekStart.Weekday() != time.Monday {
		t.Errorf("week should start on Monday, got %v", result.WeekStart.Weekday())
	}
}

func TestWeeklyAgendaWeekStartIsMonday(t *testing.T) {
	truncate(t)
	handler := query.NewWeeklyAgendaHandler(testPool)

	// El lunes de esa semana es el 31 de agosto de 2026
	expected := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		day  time.Time
	}{
		{"monday", time.Date(2026, 8, 31, 9, 0, 0, 0, time.UTC)},
		{"wednesday", time.Date(2026, 9, 2, 9, 0, 0, 0, time.UTC)},
		{"saturday", time.Date(2026, 9, 5, 9, 0, 0, 0, time.UTC)},
		{"sunday", time.Date(2026, 9, 6, 9, 0, 0, 0, time.UTC)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := handler.Handle(context.Background(), query.WeeklyAgenda{
				AnyDayOfWeek: tc.day,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !result.WeekStart.Equal(expected) {
				t.Errorf("expected week start %v, got %v", expected, result.WeekStart)
			}
		})
	}
}

func TestWeeklyAgendaGroupsByDay(t *testing.T) {
	truncate(t)
	handler := query.NewWeeklyAgendaHandler(testPool)

	monday := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	insertAppointment(t, "11111111-1111-1111-1111-111111111111", monday.Add(9*time.Hour), 1)
	insertAppointment(t, "22222222-2222-2222-2222-222222222222", monday.Add(11*time.Hour), 1)
	insertAppointment(t, "33333333-3333-3333-3333-333333333333", monday.AddDate(0, 0, 2).Add(10*time.Hour), 0)

	result, err := handler.Handle(context.Background(), query.WeeklyAgenda{AnyDayOfWeek: monday})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(result.Days[0].Appointments); got != 2 {
		t.Errorf("expected 2 appointments on Monday, got %d", got)
	}
	if got := len(result.Days[1].Appointments); got != 0 {
		t.Errorf("expected 0 appointments on Tuesday, got %d", got)
	}
	if got := len(result.Days[2].Appointments); got != 1 {
		t.Errorf("expected 1 appointment on Wednesday, got %d", got)
	}
}

func TestWeeklyAgendaOrdersAppointmentsByTime(t *testing.T) {
	truncate(t)
	handler := query.NewWeeklyAgendaHandler(testPool)

	monday := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	insertAppointment(t, "22222222-2222-2222-2222-222222222222", monday.Add(16*time.Hour), 1)
	insertAppointment(t, "11111111-1111-1111-1111-111111111111", monday.Add(9*time.Hour), 1)

	result, err := handler.Handle(context.Background(), query.WeeklyAgenda{AnyDayOfWeek: monday})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	appointments := result.Days[0].Appointments
	if len(appointments) != 2 {
		t.Fatalf("expected 2 appointments, got %d", len(appointments))
	}
	if !appointments[0].StartsAt.Before(appointments[1].StartsAt) {
		t.Error("appointments should be ordered by start time")
	}
}

func TestWeeklyAgendaExcludesCancelledAndCompleted(t *testing.T) {
	truncate(t)
	handler := query.NewWeeklyAgendaHandler(testPool)

	monday := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	insertAppointment(t, "11111111-1111-1111-1111-111111111111", monday.Add(9*time.Hour), 0)
	insertAppointment(t, "22222222-2222-2222-2222-222222222222", monday.Add(10*time.Hour), 1)
	insertAppointment(t, "33333333-3333-3333-3333-333333333333", monday.Add(11*time.Hour), 2)
	insertAppointment(t, "44444444-4444-4444-4444-444444444444", monday.Add(12*time.Hour), 3)

	result, err := handler.Handle(context.Background(), query.WeeklyAgenda{AnyDayOfWeek: monday})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := totalAppointments(result); got != 2 {
		t.Errorf("expected 2 active appointments, got %d", got)
	}
}

func TestWeeklyAgendaExcludesOtherWeeks(t *testing.T) {
	truncate(t)
	handler := query.NewWeeklyAgendaHandler(testPool)

	monday := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	insertAppointment(t, "11111111-1111-1111-1111-111111111111", monday.Add(-time.Hour), 1)
	insertAppointment(t, "22222222-2222-2222-2222-222222222222", monday.Add(10*time.Hour), 1)
	insertAppointment(t, "33333333-3333-3333-3333-333333333333", monday.AddDate(0, 0, 7).Add(10*time.Hour), 1)

	result, err := handler.Handle(context.Background(), query.WeeklyAgenda{AnyDayOfWeek: monday})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := totalAppointments(result); got != 1 {
		t.Errorf("expected 1 appointment in this week, got %d", got)
	}
}

func TestWeeklyAgendaMapsStatusLabels(t *testing.T) {
	truncate(t)
	handler := query.NewWeeklyAgendaHandler(testPool)

	monday := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	insertAppointment(t, "11111111-1111-1111-1111-111111111111", monday.Add(9*time.Hour), 0)
	insertAppointment(t, "22222222-2222-2222-2222-222222222222", monday.Add(10*time.Hour), 1)

	result, err := handler.Handle(context.Background(), query.WeeklyAgenda{AnyDayOfWeek: monday})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	appointments := result.Days[0].Appointments
	if appointments[0].Status != "reserved" {
		t.Errorf("expected 'reserved', got %q", appointments[0].Status)
	}
	if appointments[1].Status != "confirmed" {
		t.Errorf("expected 'confirmed', got %q", appointments[1].Status)
	}
}
