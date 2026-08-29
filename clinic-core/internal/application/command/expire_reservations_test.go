package command_test

import (
	"context"
	"testing"
	"time"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/application/command"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/infrastructure/persistence"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/platform/postgres"
)

func newExpiryHandler(now time.Time) *command.ExpireReservationsHandler {
	return command.NewExpireReservationsHandler(
		persistence.NewAppointmentRepository(testPool),
		persistence.NewOutboxRepository(testPool),
		postgres.NewTransactionManager(testPool),
		fixedClock{now: now},
	)
}

func TestExpiryReleasesLapsedReservations(t *testing.T) {
	truncate(t)
	ctx := context.Background()

	now := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	startsAt := now.Add(2 * time.Hour)

	// La reserva se crea a las 08:00 y su bloqueo vence a las 08:05.
	if _, err := newHandler(now).Handle(ctx, newCommand("3f2504e0-4f89-11d3-9a0c-0305e82c3301", startsAt)); err != nil {
		t.Fatalf("unexpected error reserving: %v", err)
	}

	expired, err := newExpiryHandler(now.Add(6 * time.Minute)).Handle(ctx)
	if err != nil {
		t.Fatalf("unexpected error expiring: %v", err)
	}

	if expired != 1 {
		t.Errorf("expected 1 expired reservation, got %d", expired)
	}
}

func TestExpiryLeavesActiveReservationsAlone(t *testing.T) {
	truncate(t)
	ctx := context.Background()

	now := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	startsAt := now.Add(2 * time.Hour)

	if _, err := newHandler(now).Handle(ctx, newCommand("3f2504e0-4f89-11d3-9a0c-0305e82c3301", startsAt)); err != nil {
		t.Fatalf("unexpected error reserving: %v", err)
	}

	expired, err := newExpiryHandler(now.Add(3 * time.Minute)).Handle(ctx)
	if err != nil {
		t.Fatalf("unexpected error expiring: %v", err)
	}

	if expired != 0 {
		t.Errorf("reservation is still within its window, got %d expired", expired)
	}
}

func TestExpiredSlotBecomesBookableAgain(t *testing.T) {
	truncate(t)
	ctx := context.Background()

	now := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	startsAt := now.Add(2 * time.Hour)

	if _, err := newHandler(now).Handle(ctx, newCommand("3f2504e0-4f89-11d3-9a0c-0305e82c3301", startsAt)); err != nil {
		t.Fatalf("unexpected error on first reservation: %v", err)
	}

	// Antes de expirar, la franja sigue ocupada.
	_, err := newHandler(now).Handle(ctx, newCommand("6ba7b810-9dad-11d1-80b4-00c04fd430c8", startsAt))
	if err == nil {
		t.Fatal("expected the slot to be unavailable before expiry")
	}

	later := now.Add(6 * time.Minute)
	if _, err := newExpiryHandler(later).Handle(ctx); err != nil {
		t.Fatalf("unexpected error expiring: %v", err)
	}

	// Tras expirar, otro paciente puede reservarla.
	if _, err := newHandler(later).Handle(ctx, newCommand("6ba7b810-9dad-11d1-80b4-00c04fd430c8", startsAt)); err != nil {
		t.Errorf("slot should be bookable after expiry, got %v", err)
	}
}

func TestExpiryEmitsDomainEvent(t *testing.T) {
	truncate(t)
	ctx := context.Background()

	now := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	startsAt := now.Add(2 * time.Hour)

	if _, err := newHandler(now).Handle(ctx, newCommand("3f2504e0-4f89-11d3-9a0c-0305e82c3301", startsAt)); err != nil {
		t.Fatalf("unexpected error reserving: %v", err)
	}

	if _, err := newExpiryHandler(now.Add(6 * time.Minute)).Handle(ctx); err != nil {
		t.Fatalf("unexpected error expiring: %v", err)
	}

	var count int
	row := testPool.QueryRow(ctx,
		"SELECT count(*) FROM outbox_events WHERE event_type = 'appointment.expired'")

	if err := row.Scan(&count); err != nil {
		t.Fatalf("counting expired events: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 appointment.expired event, got %d", count)
	}
}
