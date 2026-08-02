package appointment

import (
	"errors"
	"testing"
	"time"
)

func testNow() time.Time {
	return time.Date(2026, 8, 3, 9, 0, 0, 0, time.UTC)
}

func testPatientId(t *testing.T) PatientId {
	t.Helper()
	id, err := NewPatientId("3f2504e0-4f89-11d3-9a0c-0305e82c3301")
	if err != nil {
		t.Fatalf("unexpected error building patient id: %v", err)
	}
	return id
}

func testType(t *testing.T, duration time.Duration) AppointmentType {
	t.Helper()
	price, err := NewMoney(4500, "EUR")
	if err != nil {
		t.Fatalf("unexpected error building money: %v", err)
	}
	appointmentType, err := NewAppointmentType("Primera visita", duration, "#3B82F6", price)
	if err != nil {
		t.Fatalf("unexpected error building appointment type: %v", err)
	}
	return appointmentType
}

func testSlot(t *testing.T, start time.Time, duration time.Duration) TimeSlot {
	t.Helper()
	slot, err := NewTimeSlot(start, start.Add(duration))
	if err != nil {
		t.Fatalf("unexpected error building slot: %v", err)
	}
	return slot
}

func reserveTestAppointment(t *testing.T, now time.Time) *Appointment {
	t.Helper()
	duration := 60 * time.Minute
	appointment, err := Reserve(
		testPatientId(t),
		testSlot(t, now.Add(2*time.Hour), duration),
		testType(t, duration),
		now,
	)
	if err != nil {
		t.Fatalf("unexpected error reserving appointment: %v", err)
	}
	return appointment
}

func TestReserveCreatesAppointmentInReservedState(t *testing.T) {
	now := testNow()
	appointment := reserveTestAppointment(t, now)

	if appointment.Status() != StatusReserved {
		t.Errorf("expected status reserved, got %v", appointment.Status())
	}

	events := appointment.PullDomainEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 domain event, got %d", len(events))
	}
	if _, ok := events[0].(AppointmentReserved); !ok {
		t.Errorf("expected AppointmentReserved event, got %T", events[0])
	}
}

func TestReserveRejectsPastSlots(t *testing.T) {
	now := testNow()
	duration := 60 * time.Minute

	_, err := Reserve(
		testPatientId(t),
		testSlot(t, now.Add(-2*time.Hour), duration),
		testType(t, duration),
		now,
	)

	if !errors.Is(err, ErrPastAppointment) {
		t.Errorf("expected ErrPastAppointment, got %v", err)
	}
}

func TestReserveRejectsDurationMismatch(t *testing.T) {
	now := testNow()

	_, err := Reserve(
		testPatientId(t),
		testSlot(t, now.Add(2*time.Hour), 30*time.Minute),
		testType(t, 60*time.Minute),
		now,
	)

	if !errors.Is(err, ErrDurationMismatch) {
		t.Errorf("expected ErrDurationMismatch, got %v", err)
	}
}

func TestConfirmWithinReservationWindow(t *testing.T) {
	now := testNow()
	appointment := reserveTestAppointment(t, now)
	appointment.PullDomainEvents()

	if err := appointment.Confirm(now.Add(3 * time.Minute)); err != nil {
		t.Fatalf("unexpected error confirming: %v", err)
	}

	if appointment.Status() != StatusConfirmed {
		t.Errorf("expected status confirmed, got %v", appointment.Status())
	}

	events := appointment.PullDomainEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 domain event, got %d", len(events))
	}
	if _, ok := events[0].(AppointmentConfirmed); !ok {
		t.Errorf("expected AppointmentConfirmed event, got %T", events[0])
	}
}

func TestConfirmFailsAfterReservationWindow(t *testing.T) {
	now := testNow()
	appointment := reserveTestAppointment(t, now)

	err := appointment.Confirm(now.Add(6 * time.Minute))

	if !errors.Is(err, ErrReservationLapsed) {
		t.Errorf("expected ErrReservationLapsed, got %v", err)
	}
	if appointment.Status() != StatusReserved {
		t.Errorf("status should remain reserved, got %v", appointment.Status())
	}
}

func TestCannotConfirmTwice(t *testing.T) {
	now := testNow()
	appointment := reserveTestAppointment(t, now)

	if err := appointment.Confirm(now.Add(time.Minute)); err != nil {
		t.Fatalf("unexpected error on first confirm: %v", err)
	}

	err := appointment.Confirm(now.Add(2 * time.Minute))
	if !errors.Is(err, ErrNotReserved) {
		t.Errorf("expected ErrNotReserved, got %v", err)
	}
}

func TestCancelReleasesTheSlot(t *testing.T) {
	now := testNow()
	appointment := reserveTestAppointment(t, now)
	appointment.PullDomainEvents()

	if err := appointment.Cancel(now.Add(time.Minute)); err != nil {
		t.Fatalf("unexpected error cancelling: %v", err)
	}

	if appointment.Status() != StatusCancelled {
		t.Errorf("expected status cancelled, got %v", appointment.Status())
	}

	events := appointment.PullDomainEvents()
	if _, ok := events[0].(AppointmentCancelled); !ok {
		t.Errorf("expected AppointmentCancelled event, got %T", events[0])
	}
}

func TestCannotCancelTwice(t *testing.T) {
	now := testNow()
	appointment := reserveTestAppointment(t, now)

	if err := appointment.Cancel(now.Add(time.Minute)); err != nil {
		t.Fatalf("unexpected error on first cancel: %v", err)
	}

	err := appointment.Cancel(now.Add(2 * time.Minute))
	if !errors.Is(err, ErrAlreadyCancelled) {
		t.Errorf("expected ErrAlreadyCancelled, got %v", err)
	}
}

func TestCompleteRequiresConfirmedAppointment(t *testing.T) {
	now := testNow()
	appointment := reserveTestAppointment(t, now)

	err := appointment.Complete(now.Add(3 * time.Hour))
	if !errors.Is(err, ErrNotConfirmed) {
		t.Errorf("expected ErrNotConfirmed, got %v", err)
	}
}

func TestExpireCancelsLapsedReservation(t *testing.T) {
	now := testNow()
	appointment := reserveTestAppointment(t, now)
	appointment.PullDomainEvents()

	if err := appointment.Expire(now.Add(6 * time.Minute)); err != nil {
		t.Fatalf("unexpected error expiring: %v", err)
	}

	if appointment.Status() != StatusCancelled {
		t.Errorf("expected status cancelled, got %v", appointment.Status())
	}

	events := appointment.PullDomainEvents()
	if _, ok := events[0].(AppointmentExpired); !ok {
		t.Errorf("expected AppointmentExpired event, got %T", events[0])
	}
}

func TestExpireFailsWithinReservationWindow(t *testing.T) {
	now := testNow()
	appointment := reserveTestAppointment(t, now)

	err := appointment.Expire(now.Add(2 * time.Minute))
	if !errors.Is(err, ErrNotReserved) {
		t.Errorf("expected ErrNotReserved, got %v", err)
	}
}

func TestCannotCancelACompletedAppointment(t *testing.T) {
	now := testNow()
	appointment := reserveTestAppointment(t, now)

	if err := appointment.Confirm(now.Add(time.Minute)); err != nil {
		t.Fatalf("unexpected error confirming: %v", err)
	}

	afterSession := appointment.Slot().End().Add(time.Minute)
	if err := appointment.Complete(afterSession); err != nil {
		t.Fatalf("unexpected error completing: %v", err)
	}

	err := appointment.Cancel(afterSession)
	if !errors.Is(err, ErrAlreadyCompleted) {
		t.Errorf("expected ErrAlreadyCompleted, got %v", err)
	}
}

func TestCompleteFailsBeforeSessionEnds(t *testing.T) {
	now := testNow()
	appointment := reserveTestAppointment(t, now)

	if err := appointment.Confirm(now.Add(time.Minute)); err != nil {
		t.Fatalf("unexpected error confirming: %v", err)
	}

	duringSession := appointment.Slot().Start().Add(10 * time.Minute)
	err := appointment.Complete(duringSession)

	if !errors.Is(err, ErrNotFinished) {
		t.Errorf("expected ErrNotFinished, got %v", err)
	}
	if appointment.Status() != StatusConfirmed {
		t.Errorf("status should remain confirmed, got %v", appointment.Status())
	}
}

func TestCompleteSucceedsAfterSessionEnds(t *testing.T) {
	now := testNow()
	appointment := reserveTestAppointment(t, now)

	if err := appointment.Confirm(now.Add(time.Minute)); err != nil {
		t.Fatalf("unexpected error confirming: %v", err)
	}
	appointment.PullDomainEvents()

	afterSession := appointment.Slot().End().Add(time.Minute)
	if err := appointment.Complete(afterSession); err != nil {
		t.Fatalf("unexpected error completing: %v", err)
	}

	if appointment.Status() != StatusCompleted {
		t.Errorf("expected status completed, got %v", appointment.Status())
	}

	events := appointment.PullDomainEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 domain event, got %d", len(events))
	}
	if _, ok := events[0].(AppointmentCompleted); !ok {
		t.Errorf("expected AppointmentCompleted event, got %T", events[0])
	}
}
