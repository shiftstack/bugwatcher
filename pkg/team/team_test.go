package team

import (
	"strings"
	"testing"
	"time"
)

func TestLoadLeaveAvailability(t *testing.T) {
	people, err := Load(strings.NewReader(`
- kerberos: with-leave
  bug_triage: true
  leave:
  - start: 2026-06-01
    end: 2026-08-31
  - start: 2026-09-15
    end: 2026-09-15
- kerberos: without-leave
  bug_triage: true
`))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(people) != 2 {
		t.Fatalf("Load returned %d people, want 2", len(people))
	}

	tests := []struct {
		name      string
		timestamp string
		available bool
	}{
		{"before leave", "2026-05-31T23:59:59Z", true},
		{"start at midnight UTC", "2026-06-01T00:00:00Z", false},
		{"local date before start but UTC on leave", "2026-05-31T20:00:00-04:00", false},
		{"during leave", "2026-07-15T12:00:00Z", false},
		{"end at midnight UTC", "2026-08-31T00:00:00Z", false},
		{"last instant of end date", "2026-08-31T23:59:59.999999999Z", false},
		{"after leave", "2026-09-01T00:00:00Z", true},
		{"second leave lasts one day", "2026-09-15T12:00:00Z", false},
		{"after second leave", "2026-09-16T00:00:00Z", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			at, err := time.Parse(time.RFC3339Nano, tt.timestamp)
			if err != nil {
				t.Fatal(err)
			}
			if got := people[0].IsAvailable(at); got != tt.available {
				t.Errorf("IsAvailable(%s) with leave = %t, want %t", tt.timestamp, got, tt.available)
			}
			if !people[1].IsAvailable(at) {
				t.Errorf("IsAvailable(%s) without leave = false, want true", tt.timestamp)
			}
		})
	}
}
