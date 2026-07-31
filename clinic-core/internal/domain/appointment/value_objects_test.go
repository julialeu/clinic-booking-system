package appointment

import (
	"testing"
	"time"
)

func TestNewTimeSlotRejectsInvalidRanges(t *testing.T) {
	base := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)

	cases := []struct {
		name  string
		start time.Time
		end   time.Time
	}{
		{"end before start", base.Add(time.Hour), base},
		{"end equals start", base, base},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewTimeSlot(tc.start, tc.end)
			if err == nil {
				t.Fatalf("expected an error, got nil")
			}
		})
	}
}

func TestTimeSlotDuration(t *testing.T) {
	start := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	slot, err := NewTimeSlot(start, start.Add(45*time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := slot.Duration(); got != 45*time.Minute {
		t.Errorf("expected 45m, got %v", got)
	}
}

func TestTimeSlotOverlaps(t *testing.T) {
	at := func(hour, minute int) time.Time {
		return time.Date(2026, 8, 3, hour, minute, 0, 0, time.UTC)
	}

	mustSlot := func(start, end time.Time) TimeSlot {
		slot, err := NewTimeSlot(start, end)
		if err != nil {
			t.Fatalf("unexpected error building slot: %v", err)
		}
		return slot
	}

	existing := mustSlot(at(10, 0), at(11, 0))

	cases := []struct {
		name     string
		other    TimeSlot
		expected bool
	}{
		{"identical slot", mustSlot(at(10, 0), at(11, 0)), true},
		{"starts during existing", mustSlot(at(10, 30), at(11, 30)), true},
		{"ends during existing", mustSlot(at(9, 30), at(10, 30)), true},
		{"fully contained", mustSlot(at(10, 15), at(10, 45)), true},
		{"fully contains", mustSlot(at(9, 0), at(12, 0)), true},
		{"ends exactly when existing starts", mustSlot(at(9, 0), at(10, 0)), false},
		{"starts exactly when existing ends", mustSlot(at(11, 0), at(12, 0)), false},
		{"completely before", mustSlot(at(8, 0), at(9, 0)), false},
		{"completely after", mustSlot(at(12, 0), at(13, 0)), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := existing.Overlaps(tc.other); got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}
