package pager

import (
	"reflect"
	"testing"
	"time"

	"github.com/PagerDuty/go-pagerduty"
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
		wantStart    string
		wantOffset   int
	}{
		{
			name:         "less than one turn elapsed leaves anchor and cursor untouched",
			virtualStart: "2026-05-08T09:30:00-07:00",
			turn:         day,
			wantStart:    "2026-05-08T09:30:00-07:00",
			wantOffset:   0,
		},
		{
			name:         "weekly rotation rolls one cycle forward when virtual_start is 10 days old",
			virtualStart: "2026-04-28T09:30:00-07:00",
			turn:         week,
			wantStart:    "2026-05-05T09:30:00-07:00",
			wantOffset:   1,
		},
		{
			name:         "weekly rotation rolls multiple years forward preserving weekday",
			virtualStart: "2023-06-02T14:00:00-07:00", // Friday
			turn:         week,
			wantStart:    "2026-05-01T14:00:00-07:00", // most recent Friday cycle boundary at-or-before now
			wantOffset:   152,
		},
		{
			name:         "daily rotation rolls to most recent daily anchor",
			virtualStart: "2025-02-10T03:00:00-08:00",
			turn:         day,
			wantStart:    "2026-05-08T03:00:00-08:00",
			wantOffset:   452,
		},
		{
			name:         "future virtual_start untouched",
			virtualStart: "2026-06-01T00:00:00-07:00",
			turn:         day,
			wantStart:    "2026-06-01T00:00:00-07:00",
			wantOffset:   0,
		},
		{
			name:         "zero turn returns input unchanged",
			virtualStart: "2020-01-01T00:00:00Z",
			turn:         0,
			wantStart:    "2020-01-01T00:00:00Z",
			wantOffset:   0,
		},
		{
			name:         "monthly turn rolls forward in 30-day steps",
			virtualStart: "2025-02-10T03:00:00-08:00",
			turn:         30 * day,
			wantStart:    "2026-05-06T03:00:00-08:00",
			wantOffset:   15,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotStart, gotOffset := rollVirtualStartForward(mustParse(tc.virtualStart), tc.turn, now)
			if !gotStart.Equal(mustParse(tc.wantStart)) {
				t.Errorf("start: got %s, want %s", gotStart.Format(time.RFC3339), tc.wantStart)
			}
			if gotOffset != tc.wantOffset {
				t.Errorf("offset: got %d, want %d", gotOffset, tc.wantOffset)
			}
		})
	}
}

func TestRotateUsersForward(t *testing.T) {
	user := func(id string) pagerduty.UserReference {
		return pagerduty.UserReference{User: pagerduty.APIObject{ID: id}}
	}
	ids := func(users []pagerduty.UserReference) []string {
		out := make([]string, len(users))
		for i, u := range users {
			out[i] = u.User.ID
		}
		return out
	}

	tests := []struct {
		name   string
		input  []string
		cursor int
		want   []string
	}{
		{
			name:   "cursor zero leaves order unchanged",
			input:  []string{"a", "b", "c"},
			cursor: 0,
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "negative cursor leaves order unchanged",
			input:  []string{"a", "b", "c"},
			cursor: -1,
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "cursor one shifts head by one",
			input:  []string{"a", "b", "c"},
			cursor: 1,
			want:   []string{"b", "c", "a"},
		},
		{
			name:   "cursor two shifts head by two",
			input:  []string{"a", "b", "c"},
			cursor: 2,
			want:   []string{"c", "a", "b"},
		},
		{
			name:   "cursor equal to length wraps to no-op",
			input:  []string{"a", "b", "c"},
			cursor: 3,
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "cursor greater than length wraps",
			input:  []string{"a", "b", "c"},
			cursor: 4,
			want:   []string{"b", "c", "a"},
		},
		{
			name:   "cursor much greater than length wraps via modulo",
			input:  []string{"a", "b", "c", "d", "e", "f", "g"},
			cursor: 152, // matches the multi-year case in TestRollVirtualStartForward
			want:   []string{"f", "g", "a", "b", "c", "d", "e"},
		},
		{
			name:   "single user stays in place regardless of cursor",
			input:  []string{"only"},
			cursor: 99,
			want:   []string{"only"},
		},
		{
			name:   "empty input stays empty",
			input:  []string{},
			cursor: 5,
			want:   []string{},
		},
		{
			name:   "two users alternate",
			input:  []string{"a", "b"},
			cursor: 1,
			want:   []string{"b", "a"},
		},
		{
			name:   "weekly layer with seven users and cursor of one puts second-in-PD user first",
			input:  []string{"user_a", "user_b", "user_c", "user_d", "user_e", "user_f", "user_g"},
			cursor: 1,
			want:   []string{"user_b", "user_c", "user_d", "user_e", "user_f", "user_g", "user_a"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := make([]pagerduty.UserReference, len(tc.input))
			for i, id := range tc.input {
				input[i] = user(id)
			}

			// Snapshot the input so we can prove the helper never mutates it.
			inputSnapshot := append([]pagerduty.UserReference{}, input...)

			got := rotateUsersForward(input, tc.cursor)

			if !reflect.DeepEqual(ids(got), tc.want) {
				t.Errorf("rotated order: got %v, want %v", ids(got), tc.want)
			}
			if !reflect.DeepEqual(input, inputSnapshot) {
				t.Errorf("input was mutated: got %v, want %v", ids(input), ids(inputSnapshot))
			}
		})
	}
}
