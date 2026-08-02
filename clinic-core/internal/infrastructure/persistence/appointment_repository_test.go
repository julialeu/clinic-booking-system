package persistence

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/appointment"
)

func testType(t *testing.T, duration time.Duration) appointment.AppointmentType {
	t.Helper()
	price, err := appointment.NewMoney(4500, "EUR")
	if err != nil {
		t.Fatalf("building money: %v", err)
	}
	appointmentType, err := appointment.NewAppointmentType("Primera visita", duration, "#3B82F6", price)
	if err != nil {
		t.Fatalf("building appointment type: %v", err)
	}
	return appointmentType
}

func testSlot(t *testing.T, start time.Time, duration time.Duration) appointment.TimeSlot {
	t.Helper()
	slot, err := appointment.NewTimeSlot(start, start.Add(duration))
	if err != nil {
		t.Fatalf("building slot: %v", err)
	}
	return slot
}

func newTestAppointment(t *testing.T, start time.Time, duration time.Duration) *appointment.Appointment {
	t.Helper()
	patientId, err := appointment.NewPatientId("3f2504e0-4f89-11d3-9a0c-0305e82c3301")
	if err != nil {
		t.Fatalf("building patient id: %v", err)
	}

	result, err := appointment.Reserve(
		patientId,
		testSlot(t, start, duration),
		testType(t, duration),
		start.Add(-time.Hour),
	)
	if err != nil {
		t.Fatalf("reserving appointment: %v", err)
	}
	return result
}

func TestSaveAndFindById(t *testing.T) {
	truncateAppointments(t)
	ctx := context.Background()
	repo := NewAppointmentRepository(testPool)

	start := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	original := newTestAppointment(t, start, time.Hour)

	if err := repo.Save(ctx, original); err != nil {
		t.Fatalf("saving appointment: %v", err)
	}

	found, err := repo.FindById(ctx, original.Id())
	if err != nil {
		t.Fatalf("finding appointment: %v", err)
	}

	if !found.Id().Equals(original.Id()) {
		t.Errorf("id mismatch: got %s", found.Id().Value())
	}
	if found.Status() != original.Status() {
		t.Errorf("expected status %v, got %v", original.Status(), found.Status())
	}
	if !found.Slot().Start().Equal(original.Slot().Start()) {
		t.Errorf("start mismatch: got %v", found.Slot().Start())
	}
	if found.Type().Price().AmountInCents() != 4500 {
		t.Errorf("expected price 4500, got %d", found.Type().Price().AmountInCents())
	}
}

func TestFindByIdReturnsNotFound(t *testing.T) {
	truncateAppointments(t)
	ctx := context.Background()
	repo := NewAppointmentRepository(testPool)

	missingId, err := appointment.AppointmentIdFromString("11111111-2222-3333-4444-555555555555")
	if err != nil {
		t.Fatalf("building id: %v", err)
	}

	_, err = repo.FindById(ctx, missingId)
	if !errors.Is(err, appointment.ErrAppointmentNotFound) {
		t.Errorf("expected ErrAppointmentNotFound, got %v", err)
	}
}

func TestSaveUpdatesExistingAppointment(t *testing.T) {
	truncateAppointments(t)
	ctx := context.Background()
	repo := NewAppointmentRepository(testPool)

	start := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	original := newTestAppointment(t, start, time.Hour)

	if err := repo.Save(ctx, original); err != nil {
		t.Fatalf("saving appointment: %v", err)
	}

	if err := original.Confirm(start.Add(-59 * time.Minute)); err != nil {
		t.Fatalf("confirming appointment: %v", err)
	}
	if err := repo.Save(ctx, original); err != nil {
		t.Fatalf("saving confirmed appointment: %v", err)
	}

	found, err := repo.FindById(ctx, original.Id())
	if err != nil {
		t.Fatalf("finding appointment: %v", err)
	}
	if found.Status() != appointment.StatusConfirmed {
		t.Errorf("expected confirmed, got %v", found.Status())
	}

	var count int
	row := testPool.QueryRow(ctx, "SELECT count(*) FROM appointments")
	if err := row.Scan(&count); err != nil {
		t.Fatalf("counting appointments: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestFindOverlapping(t *testing.T) {
	truncateAppointments(t)
	ctx := context.Background()
	repo := NewAppointmentRepository(testPool)

	start := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	existing := newTestAppointment(t, start, time.Hour)
	if err := repo.Save(ctx, existing); err != nil {
		t.Fatalf("saving appointment: %v", err)
	}

	cases := []struct {
		name        string
		slotStart   time.Time
		duration    time.Duration
		expectFound bool
	}{
		{"same slot", start, time.Hour, true},
		{"starts during", start.Add(30 * time.Minute), time.Hour, true},
		{"ends during", start.Add(-30 * time.Minute), time.Hour, true},
		{"contained", start.Add(15 * time.Minute), 30 * time.Minute, true},
		{"right before", start.Add(-time.Hour), time.Hour, false},
		{"right after", start.Add(time.Hour), time.Hour, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := repo.FindOverlapping(ctx, testSlot(t, tc.slotStart, tc.duration))
			if err != nil {
				t.Fatalf("finding overlapping: %v", err)
			}
			if found := len(results) > 0; found != tc.expectFound {
				t.Errorf("expected found=%v, got %d results", tc.expectFound, len(results))
			}
		})
	}
}

func TestFindOverlappingIgnoresCancelled(t *testing.T) {
	truncateAppointments(t)
	ctx := context.Background()
	repo := NewAppointmentRepository(testPool)

	start := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	existing := newTestAppointment(t, start, time.Hour)

	if err := existing.Cancel(start.Add(-30 * time.Minute)); err != nil {
		t.Fatalf("cancelling appointment: %v", err)
	}
	if err := repo.Save(ctx, existing); err != nil {
		t.Fatalf("saving appointment: %v", err)
	}

	results, err := repo.FindOverlapping(ctx, testSlot(t, start, time.Hour))
	if err != nil {
		t.Fatalf("finding overlapping: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("cancelled appointments should not block the slot, got %d", len(results))
	}
}

func TestFindExpiredReservations(t *testing.T) {
	truncateAppointments(t)
	ctx := context.Background()
	repo := NewAppointmentRepository(testPool)

	start := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	reserved := newTestAppointment(t, start, time.Hour)
	if err := repo.Save(ctx, reserved); err != nil {
		t.Fatalf("saving appointment: %v", err)
	}

	beforeExpiry := reserved.ReservedUntil().Add(-time.Minute)
	results, err := repo.FindExpiredReservations(ctx, beforeExpiry)
	if err != nil {
		t.Fatalf("finding expired: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no expired reservations, got %d", len(results))
	}

	afterExpiry := reserved.ReservedUntil().Add(time.Minute)
	results, err = repo.FindExpiredReservations(ctx, afterExpiry)
	if err != nil {
		t.Fatalf("finding expired: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 expired reservation, got %d", len(results))
	}
}
