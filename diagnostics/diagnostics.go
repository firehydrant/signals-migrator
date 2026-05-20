package diagnostics

import (
	"fmt"
	"io"

	"github.com/firehydrant/signals-migrator/store"
)

// Write renders the migration diagnostics report to w.
// It reports schedules that will have incomplete member coverage due to missing
// FireHydrant users, grouped by schedule and rotation.
// Returns without writing if there are no skips.
func Write(w io.Writer, skips []store.ListRotationMemberSkipsRow) error {
	if len(skips) == 0 {
		return nil
	}

	type rotKey struct{ schedule, rotation string }
	grouped := make(map[rotKey][]store.ListRotationMemberSkipsRow)
	var order []rotKey
	seenKeys := make(map[rotKey]bool)
	seenSchedules := make(map[string]bool)
	seenUsers := make(map[string]bool)

	for _, s := range skips {
		k := rotKey{s.ScheduleName, s.RotationName}
		if !seenKeys[k] {
			order = append(order, k)
			seenKeys[k] = true
		}
		grouped[k] = append(grouped[k], s)
		seenSchedules[s.ScheduleName] = true
		seenUsers[s.UserID] = true
	}

	lines := []string{
		"DIAGNOSTICS: Incomplete Schedule Coverage",
		"==========================================",
		"",
		"The following schedules will NOT have 100% member coverage because",
		"one or more users are not matched to a FireHydrant user.",
		"",
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}

	var lastSchedule string
	for _, k := range order {
		if k.schedule != lastSchedule {
			if _, err := fmt.Fprintf(w, "  Schedule: %q\n", k.schedule); err != nil {
				return err
			}
			lastSchedule = k.schedule
		}
		if _, err := fmt.Fprintf(w, "    Rotation: %q\n", k.rotation); err != nil {
			return err
		}
		for _, s := range grouped[k] {
			id := s.UserEmail
			if id == "" {
				id = s.UserID
			}
			if _, err := fmt.Fprintf(w, "      - %s (ID: %s) — %s\n", id, s.UserID, s.Reason); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "%d schedule(s) affected, %d user(s) missing.\n", len(seenSchedules), len(seenUsers)); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, "To fix: ensure these users exist in FireHydrant and re-run the migration.")
	return err
}
