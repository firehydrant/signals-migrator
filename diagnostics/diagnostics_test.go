package diagnostics_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/firehydrant/signals-migrator/diagnostics"
	"github.com/firehydrant/signals-migrator/store"
)

// errWriter is an io.Writer that always returns an error.
type errWriter struct{ err error }

func (e errWriter) Write(_ []byte) (int, error) { return 0, e.err }

func TestWrite_NoSkips(t *testing.T) {
	var b strings.Builder
	if err := diagnostics.Write(&b, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if b.Len() != 0 {
		t.Errorf("expected no output for empty skips, got: %q", b.String())
	}
}

func TestWrite_SingleSkip(t *testing.T) {
	skips := []store.ListRotationMemberSkipsRow{
		{
			ScheduleName: "CS-on-call",
			RotationName: "Layer 1",
			RotationID:   "P3BRVNT",
			UserID:       "P1VTA5W",
			UserEmail:    "",
			Reason:       "missing_fh_user",
		},
	}

	var b strings.Builder
	if err := diagnostics.Write(&b, skips); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	out := b.String()
	assertContains(t, out, `Schedule: "CS-on-call"`)
	assertContains(t, out, `Rotation: "Layer 1"`)
	assertContains(t, out, "P1VTA5W")
	assertContains(t, out, "missing_fh_user")
	assertContains(t, out, "1 schedule(s) affected, 1 user(s) missing.")
}

func TestWrite_EmailFallsBackToUserID(t *testing.T) {
	skips := []store.ListRotationMemberSkipsRow{
		{ScheduleName: "S", RotationName: "R", UserID: "PXYZ", UserEmail: "", Reason: "missing_fh_user"},
	}

	var b strings.Builder
	if err := diagnostics.Write(&b, skips); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	out := b.String()
	// When email is empty, the user line should show the ID in both the display and ID positions.
	if !strings.Contains(out, "- PXYZ (ID: PXYZ)") {
		t.Errorf("expected user ID fallback in output, got:\n%s", out)
	}
}

func TestWrite_EmailShownWhenPresent(t *testing.T) {
	skips := []store.ListRotationMemberSkipsRow{
		{ScheduleName: "S", RotationName: "R", UserID: "PXYZ", UserEmail: "jane@example.com", Reason: "missing_fh_user"},
	}

	var b strings.Builder
	if err := diagnostics.Write(&b, skips); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	out := b.String()
	if !strings.Contains(out, "- jane@example.com (ID: PXYZ)") {
		t.Errorf("expected email in output, got:\n%s", out)
	}
}

func TestWrite_MultipleSchedulesAndRotations(t *testing.T) {
	skips := []store.ListRotationMemberSkipsRow{
		{ScheduleName: "Infra", RotationName: "Primary", UserID: "PA", UserEmail: "a@example.com", Reason: "missing_fh_user"},
		{ScheduleName: "Infra", RotationName: "Primary", UserID: "PB", UserEmail: "b@example.com", Reason: "missing_fh_user"},
		{ScheduleName: "Infra", RotationName: "Secondary", UserID: "PC", UserEmail: "c@example.com", Reason: "missing_fh_user"},
		{ScheduleName: "Platform", RotationName: "Primary", UserID: "PA", UserEmail: "a@example.com", Reason: "missing_fh_user"},
	}

	var b strings.Builder
	if err := diagnostics.Write(&b, skips); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	out := b.String()
	assertContains(t, out, `Schedule: "Infra"`)
	assertContains(t, out, `Schedule: "Platform"`)
	// PA appears in two schedules but is one distinct user.
	assertContains(t, out, "2 schedule(s) affected, 3 user(s) missing.")
}

func TestWrite_RotationGroupingPreservesOrder(t *testing.T) {
	// SQL returns rows ordered by schedule name, rotation name, user email.
	// Verify that the report groups correctly and doesn't duplicate schedule headers.
	skips := []store.ListRotationMemberSkipsRow{
		{ScheduleName: "Alpha", RotationName: "R1", UserID: "P1", UserEmail: "a@example.com", Reason: "missing_fh_user"},
		{ScheduleName: "Alpha", RotationName: "R1", UserID: "P2", UserEmail: "b@example.com", Reason: "missing_fh_user"},
		{ScheduleName: "Alpha", RotationName: "R2", UserID: "P3", UserEmail: "c@example.com", Reason: "missing_fh_user"},
	}

	var b strings.Builder
	if err := diagnostics.Write(&b, skips); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	out := b.String()
	// "Alpha" should appear exactly once as a schedule header.
	if count := strings.Count(out, `Schedule: "Alpha"`); count != 1 {
		t.Errorf(`expected Schedule: "Alpha" to appear once, got %d times`, count)
	}
	assertContains(t, out, "1 schedule(s) affected, 3 user(s) missing.")
}

func TestWrite_WriterError(t *testing.T) {
	sentinel := errors.New("disk full")
	skips := []store.ListRotationMemberSkipsRow{
		{ScheduleName: "S", RotationName: "R", UserID: "P1", Reason: "missing_fh_user"},
	}
	err := diagnostics.Write(errWriter{sentinel}, skips)
	if !errors.Is(err, sentinel) {
		t.Errorf("expected writer error to be propagated, got: %v", err)
	}
}

func TestWrite_SameUserInMultipleRotations(t *testing.T) {
	// PA is skipped in two different rotations across two schedules.
	// They should appear in the report twice (once per rotation) but count as 1 missing user.
	skips := []store.ListRotationMemberSkipsRow{
		{ScheduleName: "Infra", RotationName: "Primary", UserID: "PA", UserEmail: "a@example.com", Reason: "missing_fh_user"},
		{ScheduleName: "Platform", RotationName: "Primary", UserID: "PA", UserEmail: "a@example.com", Reason: "missing_fh_user"},
	}

	var b strings.Builder
	if err := diagnostics.Write(&b, skips); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	out := b.String()
	assertContains(t, out, "2 schedule(s) affected, 1 user(s) missing.")
	if count := strings.Count(out, "a@example.com"); count != 2 {
		t.Errorf("expected user to appear in report twice (once per rotation), got %d times", count)
	}
}

func TestWrite_DuplicateSkipInSameRotation(t *testing.T) {
	// The DB PRIMARY KEY prevents this at the store layer, but Write accepts a plain slice.
	// Duplicates should not inflate the user count or collapse the lines.
	skips := []store.ListRotationMemberSkipsRow{
		{ScheduleName: "S", RotationName: "R", UserID: "PA", UserEmail: "a@example.com", Reason: "missing_fh_user"},
		{ScheduleName: "S", RotationName: "R", UserID: "PA", UserEmail: "a@example.com", Reason: "missing_fh_user"},
	}

	var b strings.Builder
	if err := diagnostics.Write(&b, skips); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	out := b.String()
	// Distinct user count should be 1, not 2.
	assertContains(t, out, "1 schedule(s) affected, 1 user(s) missing.")
	// Both lines still appear (Write doesn't deduplicate rows, the DB constraint does).
	if count := strings.Count(out, "a@example.com"); count != 2 {
		t.Errorf("expected duplicate rows to appear as-is in report, got %d lines", count)
	}
}

func assertContains(t *testing.T, output, substr string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Errorf("expected output to contain %q\nfull output:\n%s", substr, output)
	}
}
