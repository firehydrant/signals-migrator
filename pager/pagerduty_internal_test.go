package pager

import (
	"testing"
	"time"
)

func TestRollVirtualStartForward(t *testing.T) {
	mustParse := func(s string) time.Time {
		t.Helper()
		v, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatalf("parsing %q: %s", s, err)
		}
		return v
	}

	now := mustParse("2026-05-08T12:00:00-07:00")
	day := 24 * time.Hour
	week := 7 * day

	tests := []struct {
		name         string
		virtualStart string
		turn         time.Duration
		want         string
	}{
		{
			name:         "already within window leaves untouched",
			virtualStart: "2026-04-28T09:30:00-07:00",
			turn:         day,
			want:         "2026-04-28T09:30:00-07:00",
		},
		{
			name:         "weekly rotation rolls forward to within window preserving weekday",
			virtualStart: "2023-06-02T14:00:00-07:00", // Friday
			turn:         week,
			want:         "2026-04-17T14:00:00-07:00", // first Friday at-or-after cutoff
		},
		{
			name:         "daily rotation rolls forward by days",
			virtualStart: "2025-02-10T03:00:00-08:00",
			turn:         day,
			want:         "2026-04-14T03:00:00-08:00", // first daily anchor at-or-after cutoff
		},
		{
			name:         "future virtual_start untouched",
			virtualStart: "2026-06-01T00:00:00-07:00",
			turn:         day,
			want:         "2026-06-01T00:00:00-07:00",
		},
		{
			name:         "zero turn returns input unchanged",
			virtualStart: "2020-01-01T00:00:00Z",
			turn:         0,
			want:         "2020-01-01T00:00:00Z",
		},
		{
			name:         "monthly turn jumps past the cutoff in a single step",
			virtualStart: "2025-02-10T03:00:00-08:00",
			turn:         30 * day,
			want:         "2026-05-06T03:00:00-08:00",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := rollVirtualStartForward(mustParse(tc.virtualStart), tc.turn, now)
			if !got.Equal(mustParse(tc.want)) {
				t.Fatalf("got %s, want %s", got.Format(time.RFC3339), tc.want)
			}
			cutoff := now.Add(-fhStartTimeWindow)
			if tc.turn > 0 && got.Before(cutoff) && !got.Equal(mustParse(tc.virtualStart)) {
				t.Fatalf("rolled value %s is still before cutoff %s", got, cutoff)
			}
		})
	}
}
