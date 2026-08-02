package command_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/application/command"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/infrastructure/persistence"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/platform/postgres"
)

var testPool *pgxpool.Pool

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

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

func newHandler(now time.Time) *command.ReserveAppointmentHandler {
	return command.NewReserveAppointmentHandler(
		persistence.NewAppointmentRepository(testPool),
		postgres.NewTransactionManager(testPool),
		fixedClock{now: now},
	)
}

func newCommand(patientId string, startsAt time.Time) command.ReserveAppointment {
	return command.ReserveAppointment{
		PatientId:     patientId,
		StartsAt:      startsAt,
		TypeName:      "Primera visita",
		TypeDuration:  time.Hour,
		TypeColor:     "#3B82F6",
		PriceCents:    4500,
		PriceCurrency: "EUR",
	}
}

func truncate(t *testing.T) {
	t.Helper()
	if _, err := testPool.Exec(context.Background(), "TRUNCATE appointments"); err != nil {
		t.Fatalf("truncating: %v", err)
	}
}

func countAppointments(t *testing.T) int {
	t.Helper()
	var count int
	row := testPool.QueryRow(context.Background(), "SELECT count(*) FROM appointments")
	if err := row.Scan(&count); err != nil {
		t.Fatalf("counting: %v", err)
	}
	return count
}

func TestReserveAppointmentSucceeds(t *testing.T) {
	truncate(t)
	now := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	startsAt := now.Add(2 * time.Hour)

	handler := newHandler(now)
	id, err := handler.Handle(context.Background(), newCommand("3f2504e0-4f89-11d3-9a0c-0305e82c3301", startsAt))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id.Value() == "" {
		t.Error("expected a non-empty appointment id")
	}
	if got := countAppointments(t); got != 1 {
		t.Errorf("expected 1 appointment, got %d", got)
	}
}

func TestReserveAppointmentRejectsOccupiedSlot(t *testing.T) {
	truncate(t)
	now := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	startsAt := now.Add(2 * time.Hour)
	handler := newHandler(now)
	ctx := context.Background()

	if _, err := handler.Handle(ctx, newCommand("3f2504e0-4f89-11d3-9a0c-0305e82c3301", startsAt)); err != nil {
		t.Fatalf("unexpected error on first reservation: %v", err)
	}

	_, err := handler.Handle(ctx, newCommand("6ba7b810-9dad-11d1-80b4-00c04fd430c8", startsAt))
	if !errors.Is(err, command.ErrSlotNotAvailable) {
		t.Errorf("expected ErrSlotNotAvailable, got %v", err)
	}

	if got := countAppointments(t); got != 1 {
		t.Errorf("expected 1 appointment, got %d", got)
	}
}

// Verifica que el bloqueo transaccional impide dobles reservas:
// varios pacientes compiten por la misma franja y solo uno gana.
func TestConcurrentReservationsOnSameSlot(t *testing.T) {
	truncate(t)
	now := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	startsAt := now.Add(2 * time.Hour)
	handler := newHandler(now)

	const attempts = 8
	patientIds := []string{
		"3f2504e0-4f89-11d3-9a0c-0305e82c3301",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"6ba7b811-9dad-11d1-80b4-00c04fd430c8",
		"6ba7b812-9dad-11d1-80b4-00c04fd430c8",
		"6ba7b814-9dad-11d1-80b4-00c04fd430c8",
		"7ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"8ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"9ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}

	var wg sync.WaitGroup
	results := make(chan error, attempts)
	start := make(chan struct{})

	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(patientId string) {
			defer wg.Done()
			<-start
			_, err := handler.Handle(context.Background(), newCommand(patientId, startsAt))
			results <- err
		}(patientIds[i])
	}

	// Libera todas las goroutines a la vez para maximizar la contención.
	close(start)
	wg.Wait()
	close(results)

	var succeeded, rejected int
	for err := range results {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, command.ErrSlotNotAvailable):
			rejected++
		default:
			t.Errorf("unexpected error: %v", err)
		}
	}

	if succeeded != 1 {
		t.Errorf("expected exactly 1 successful reservation, got %d", succeeded)
	}
	if rejected != attempts-1 {
		t.Errorf("expected %d rejections, got %d", attempts-1, rejected)
	}
	if got := countAppointments(t); got != 1 {
		t.Errorf("expected 1 row in database, got %d", got)
	}
}

func TestReserveAppointmentAllowsAdjacentSlots(t *testing.T) {
	truncate(t)
	now := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	startsAt := now.Add(2 * time.Hour)
	handler := newHandler(now)
	ctx := context.Background()

	if _, err := handler.Handle(ctx, newCommand("3f2504e0-4f89-11d3-9a0c-0305e82c3301", startsAt)); err != nil {
		t.Fatalf("unexpected error on first reservation: %v", err)
	}

	adjacent := startsAt.Add(time.Hour)
	if _, err := handler.Handle(ctx, newCommand("6ba7b810-9dad-11d1-80b4-00c04fd430c8", adjacent)); err != nil {
		t.Errorf("adjacent slot should be bookable, got %v", err)
	}

	if got := countAppointments(t); got != 2 {
		t.Errorf("expected 2 appointments, got %d", got)
	}
}

var _ = fmt.Sprintf
