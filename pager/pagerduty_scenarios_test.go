package pager_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/firehydrant/signals-migrator/pager"
	"github.com/firehydrant/signals-migrator/store"
	"github.com/firehydrant/signals-migrator/tfrender"
	"gotest.tools/v3/golden"
)

// pdScenarioSetup mirrors the setup() helper in pagerduty_test.go. Each scenario
// gets its own temp DB + httptest server pointed at its own apiserver/ fixture dir.
func pdScenarioSetup(t *testing.T) (context.Context, pager.Pager) {
	t.Helper()
	ctx := withTestDB(t)
	ts := pagerProviderHttpServer(t)
	pd := pager.NewPagerDutyWithURL("api-key-very-secret", ts.URL)
	return ctx, pd
}

// markAllForImport links every ext_user/ext_team to a corresponding FireHydrant
// row and marks the teams to-import, mirroring the idiom used by
// TestPagerDuty/TerraformOutputMatchesScheduleOrder. Without this, tfrender skips
// resources because nothing's flagged for import.
func markAllForImport(t *testing.T, ctx context.Context) {
	t.Helper()
	q := store.UseQueries(ctx)

	teams, err := q.ListExtTeams(ctx)
	if err != nil {
		t.Fatalf("error listing teams: %s", err)
	}
	for _, team := range teams {
		fhTeamID := fmt.Sprintf("fh-team-%s", team.ID)
		if err := q.InsertFhTeam(ctx, store.InsertFhTeamParams{
			ID:   fhTeamID,
			Name: team.Name,
			Slug: fhTeamID,
		}); err != nil {
			t.Fatalf("error creating FireHydrant team: %s", err)
		}
		if err := q.LinkExtTeam(ctx, store.LinkExtTeamParams{
			ID:       team.ID,
			FhTeamID: sql.NullString{String: fhTeamID, Valid: true},
		}); err != nil {
			t.Fatalf("error linking team: %s", err)
		}
		if err := q.MarkExtTeamToImport(ctx, team.ID); err != nil {
			t.Fatalf("error marking team for import: %s", err)
		}
	}

	users, err := q.ListExtUsers(ctx)
	if err != nil {
		t.Fatalf("error listing users: %s", err)
	}
	for _, user := range users {
		fhUserID := fmt.Sprintf("fh-user-%s", user.ID)
		if err := q.InsertFhUser(ctx, store.InsertFhUserParams{
			ID:    fhUserID,
			Email: user.Email,
			Name:  user.Name,
		}); err != nil {
			t.Fatalf("error creating FireHydrant user: %s", err)
		}
		if err := q.LinkExtUser(ctx, store.LinkExtUserParams{
			ID:       user.ID,
			FhUserID: sql.NullString{String: fhUserID, Valid: true},
		}); err != nil {
			t.Fatalf("error linking user: %s", err)
		}
	}
}

// assertGoldenTF renders the Terraform output and compares to
// testdata/<TestName>/render.golden.tf via gotest.tools/v3/golden (supports -update).
func assertGoldenTF(t *testing.T, ctx context.Context) {
	t.Helper()
	tfFile := filepath.Join(t.TempDir(), "render.tf")
	tfr, err := tfrender.New(tfFile)
	if err != nil {
		t.Fatalf("error creating Terraform renderer: %s", err)
	}
	if err := tfr.Write(ctx); err != nil {
		t.Fatalf("error writing Terraform file: %s", err)
	}
	got, err := os.ReadFile(tfFile)
	if err != nil {
		t.Fatalf("error reading Terraform file: %s", err)
	}
	golden.Assert(t, string(got), filepath.Join(t.Name(), "render.golden.tf"))
}

// loadAll runs the standard PagerDuty load pipeline used by every scenario.
func loadAll(t *testing.T, ctx context.Context, pd pager.Pager, withEP bool) {
	t.Helper()
	if err := pd.UseTeamInterface("team"); err != nil {
		t.Fatalf("error setting team interface: %s", err)
	}
	if err := pd.LoadUsers(ctx); err != nil {
		t.Fatalf("error loading users: %s", err)
	}
	if err := pd.LoadTeams(ctx); err != nil {
		t.Fatalf("error loading teams: %s", err)
	}
	if err := pd.LoadTeamMembers(ctx); err != nil {
		t.Fatalf("error loading team members: %s", err)
	}
	if err := pd.LoadSchedules(ctx); err != nil {
		t.Fatalf("error loading schedules: %s", err)
	}
	if withEP {
		if err := pd.LoadEscalationPolicies(ctx); err != nil {
			t.Fatalf("error loading escalation policies: %s", err)
		}
	}
}

// =============================================================================
// Scenario 1: Follow-The-Sun — one schedule, one team, four stacked layers.
// =============================================================================
func TestPagerDutyFollowTheSun(t *testing.T) {
	ctx, pd := pdScenarioSetup(t)
	loadAll(t, ctx, pd, false)
	q := store.UseQueries(ctx)

	rotations, err := q.ListExtRotationsByScheduleID(ctx, "S1FTS")
	if err != nil {
		t.Fatalf("error loading rotations: %s", err)
	}
	if len(rotations) != 4 {
		t.Fatalf("expected 4 rotations for S1FTS, got %d", len(rotations))
	}

	expected := []struct {
		id              string
		members         []string
		virtualStart    string
		restrictionLen  int
		firstStartDay   string
		firstEndDay     string
		firstStartTime  string
		firstEndTime    string
	}{
		{"L1FTSA", []string{"U1AVERY", "U1BLAKE", "U1CAM"}, "2025-06-02T00:00:00Z", 5, "monday", "monday", "00:00:00", "08:00:00"},
		{"L1FTSE", []string{"U1DREW", "U1EVAN", "U1FAYE", "U1GALE"}, "2025-06-02T08:00:00Z", 5, "monday", "monday", "08:00:00", "16:00:00"},
		{"L1FTSW", []string{"U1HARLAN", "U1INDRA", "U1JULES", "U1KAI", "U1LIU"}, "2025-06-02T16:00:00Z", 5, "monday", "tuesday", "16:00:00", "00:00:00"},
		// Weekend layer: first restriction is start_day_of_week=6 (Saturday under
		// PD's "Sunday=0" convention). The migrator's +1 offset (pagerduty.go:505)
		// preserves Saturday and wraps a 24h window into Sunday.
		{"L1FTSX", []string{"U1MORGAN", "U1NOOR"}, "2025-06-07T00:00:00Z", 2, "saturday", "sunday", "00:00:00", "00:00:00"},
	}

	for i, want := range expected {
		got := rotations[i]
		if got.ID != want.id {
			t.Errorf("rotation %d: got id %q, want %q", i, got.ID, want.id)
		}
		if got.RotationOrder != int64(i) {
			t.Errorf("rotation %s: got RotationOrder %d, want %d", want.id, got.RotationOrder, i)
		}
		if got.StartTime != want.virtualStart {
			t.Errorf("rotation %s: got StartTime %q, want %q", want.id, got.StartTime, want.virtualStart)
		}

		members, err := q.ListExtRotationMembers(ctx, want.id)
		if err != nil {
			t.Fatalf("error loading members for %s: %s", want.id, err)
		}
		if len(members) != len(want.members) {
			t.Fatalf("rotation %s: got %d members, want %d", want.id, len(members), len(want.members))
		}
		for j, m := range members {
			if m.UserID != want.members[j] {
				t.Errorf("rotation %s member %d: got %s, want %s", want.id, j, m.UserID, want.members[j])
			}
		}

		restrictions, err := q.ListExtRotationRestrictions(ctx, want.id)
		if err != nil {
			t.Fatalf("error loading restrictions for %s: %s", want.id, err)
		}
		if len(restrictions) != want.restrictionLen {
			t.Errorf("rotation %s: got %d restrictions, want %d", want.id, len(restrictions), want.restrictionLen)
		}
		if len(restrictions) > 0 {
			r := restrictions[0]
			if r.StartDay != want.firstStartDay {
				t.Errorf("rotation %s first restriction: got StartDay %q, want %q", want.id, r.StartDay, want.firstStartDay)
			}
			if r.EndDay != want.firstEndDay {
				t.Errorf("rotation %s first restriction: got EndDay %q, want %q", want.id, r.EndDay, want.firstEndDay)
			}
			if r.StartTime != want.firstStartTime {
				t.Errorf("rotation %s first restriction: got StartTime %q, want %q", want.id, r.StartTime, want.firstStartTime)
			}
			if r.EndTime != want.firstEndTime {
				t.Errorf("rotation %s first restriction: got EndTime %q, want %q", want.id, r.EndTime, want.firstEndTime)
			}
		}
	}

	markAllForImport(t, ctx)
	assertGoldenTF(t, ctx)
}

// =============================================================================
// Scenario 2: Custom cadence with a daily restriction that crosses midnight.
// =============================================================================
func TestPagerDutyCustomCadenceMidnightCrossing(t *testing.T) {
	ctx, pd := pdScenarioSetup(t)
	loadAll(t, ctx, pd, false)
	q := store.UseQueries(ctx)

	r, err := q.GetExtRotation(ctx, "L2CAD")
	if err != nil {
		t.Fatalf("error loading rotation L2CAD: %s", err)
	}
	if r.Strategy != "custom" {
		t.Errorf("Strategy: got %q, want %q", r.Strategy, "custom")
	}
	if r.ShiftDuration != "PT1209600S" {
		t.Errorf("ShiftDuration: got %q, want %q", r.ShiftDuration, "PT1209600S")
	}
	if r.StartTime != "2024-03-15T22:00:00-04:00" {
		t.Errorf("StartTime: got %q, want %q (virtual_start should pass through verbatim)", r.StartTime, "2024-03-15T22:00:00-04:00")
	}

	members, err := q.ListExtRotationMembers(ctx, "L2CAD")
	if err != nil {
		t.Fatalf("error loading rotation members: %s", err)
	}
	wantMembers := []string{"U2AVERY", "U2BLAKE", "U2CAM", "U2DREW"}
	if len(members) != len(wantMembers) {
		t.Fatalf("expected %d members, got %d", len(wantMembers), len(members))
	}
	for i, m := range members {
		if m.UserID != wantMembers[i] {
			t.Errorf("member %d: got %s, want %s", i, m.UserID, wantMembers[i])
		}
	}

	restrictions, err := q.ListExtRotationRestrictions(ctx, "L2CAD")
	if err != nil {
		t.Fatalf("error loading restrictions: %s", err)
	}
	if len(restrictions) != 7 {
		t.Fatalf("expected 7 daily-expanded restrictions, got %d", len(restrictions))
	}
	expectedDays := []struct{ start, end string }{
		{"sunday", "monday"},
		{"monday", "tuesday"},
		{"tuesday", "wednesday"},
		{"wednesday", "thursday"},
		{"thursday", "friday"},
		{"friday", "saturday"},
		{"saturday", "sunday"},
	}
	for i, r := range restrictions {
		if r.StartDay != expectedDays[i].start {
			t.Errorf("restriction %d: got StartDay %q, want %q", i, r.StartDay, expectedDays[i].start)
		}
		if r.EndDay != expectedDays[i].end {
			t.Errorf("restriction %d: got EndDay %q, want %q", i, r.EndDay, expectedDays[i].end)
		}
		if r.StartTime != "22:00:00" {
			t.Errorf("restriction %d: got StartTime %q, want 22:00:00", i, r.StartTime)
		}
		if r.EndTime != "07:00:00" {
			t.Errorf("restriction %d: got EndTime %q, want 07:00:00", i, r.EndTime)
		}
	}

	markAllForImport(t, ctx)
	assertGoldenTF(t, ctx)
}

// =============================================================================
// Scenario 3: 7-entry weekly restriction covering every day of the week. Pins
// the migrator's "+1 to map PD Sunday=0 → Go Sunday" offset at pagerduty.go:505.
// =============================================================================
func TestPagerDutyAllWeekdaysWeeklyRestriction(t *testing.T) {
	ctx, pd := pdScenarioSetup(t)
	loadAll(t, ctx, pd, false)
	q := store.UseQueries(ctx)

	r, err := q.GetExtRotation(ctx, "L3WK")
	if err != nil {
		t.Fatalf("error loading rotation L3WK: %s", err)
	}
	if r.Strategy != "weekly" {
		t.Errorf("Strategy: got %q, want weekly", r.Strategy)
	}
	if r.HandoffDay != "sunday" {
		t.Errorf("HandoffDay: got %q, want sunday", r.HandoffDay)
	}
	if r.HandoffTime != "09:00:00" {
		t.Errorf("HandoffTime: got %q, want 09:00:00", r.HandoffTime)
	}

	restrictions, err := q.ListExtRotationRestrictions(ctx, "L3WK")
	if err != nil {
		t.Fatalf("error loading restrictions: %s", err)
	}
	if len(restrictions) != 7 {
		t.Fatalf("expected 7 restrictions, got %d", len(restrictions))
	}
	expected := []struct{ start, end string }{
		{"sunday", "sunday"},
		{"monday", "monday"},
		{"tuesday", "tuesday"},
		{"wednesday", "wednesday"},
		{"thursday", "thursday"},
		{"friday", "friday"},
		{"saturday", "saturday"},
	}
	for i, want := range expected {
		got := restrictions[i]
		if got.StartDay != want.start {
			t.Errorf("restriction %d: got StartDay %q, want %q", i, got.StartDay, want.start)
		}
		if got.EndDay != want.end {
			t.Errorf("restriction %d: got EndDay %q, want %q", i, got.EndDay, want.end)
		}
		if got.StartTime != "09:00:00" {
			t.Errorf("restriction %d: got StartTime %q, want 09:00:00", i, got.StartTime)
		}
		if got.EndTime != "21:00:00" {
			t.Errorf("restriction %d: got EndTime %q, want 21:00:00", i, got.EndTime)
		}
	}

	markAllForImport(t, ctx)
	assertGoldenTF(t, ctx)
}

// =============================================================================
// Scenario 4: Three layers — one ended (filtered), one weekly, one daily that
// crosses midnight. Ended layer must produce zero artifacts in store + Terraform.
// =============================================================================
func TestPagerDutyMixedStrategiesWithEndedLayer(t *testing.T) {
	ctx, pd := pdScenarioSetup(t)
	loadAll(t, ctx, pd, false)
	q := store.UseQueries(ctx)

	rotations, err := q.ListExtRotationsByScheduleID(ctx, "S4MIX")
	if err != nil {
		t.Fatalf("error loading rotations: %s", err)
	}
	if len(rotations) != 2 {
		t.Fatalf("expected 2 rotations (ended layer filtered), got %d", len(rotations))
	}
	wantOrdered := []string{"L4PRIM", "L4SEC"}
	for i, want := range wantOrdered {
		if rotations[i].ID != want {
			t.Errorf("rotation %d: got %s, want %s", i, rotations[i].ID, want)
		}
	}

	if _, err := q.GetExtRotation(ctx, "L4END"); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("ended layer L4END should return ErrNoRows from GetExtRotation, got: %v", err)
	}

	// Layer B (L4PRIM): weekly, 5 restrictions Mon-Fri, 5 members in order.
	rb, err := q.GetExtRotation(ctx, "L4PRIM")
	if err != nil {
		t.Fatalf("error loading L4PRIM: %s", err)
	}
	if rb.Strategy != "weekly" {
		t.Errorf("L4PRIM Strategy: got %q, want weekly", rb.Strategy)
	}
	restB, err := q.ListExtRotationRestrictions(ctx, "L4PRIM")
	if err != nil {
		t.Fatalf("error loading L4PRIM restrictions: %s", err)
	}
	if len(restB) != 5 {
		t.Errorf("L4PRIM: expected 5 restrictions, got %d", len(restB))
	}
	membersB, err := q.ListExtRotationMembers(ctx, "L4PRIM")
	if err != nil {
		t.Fatalf("error loading L4PRIM members: %s", err)
	}
	wantB := []string{"U4AVERY", "U4BLAKE", "U4CAM", "U4DREW", "U4EVAN"}
	if len(membersB) != len(wantB) {
		t.Fatalf("L4PRIM: expected %d members, got %d", len(wantB), len(membersB))
	}
	for i, m := range membersB {
		if m.UserID != wantB[i] {
			t.Errorf("L4PRIM member %d: got %s, want %s", i, m.UserID, wantB[i])
		}
	}

	// Layer C (L4SEC): daily, daily_restriction crossing midnight expands to 7 entries.
	rc, err := q.GetExtRotation(ctx, "L4SEC")
	if err != nil {
		t.Fatalf("error loading L4SEC: %s", err)
	}
	if rc.Strategy != "daily" {
		t.Errorf("L4SEC Strategy: got %q, want daily", rc.Strategy)
	}
	restC, err := q.ListExtRotationRestrictions(ctx, "L4SEC")
	if err != nil {
		t.Fatalf("error loading L4SEC restrictions: %s", err)
	}
	if len(restC) != 7 {
		t.Fatalf("L4SEC: expected 7 daily-expanded restrictions, got %d", len(restC))
	}
	dayOrder := []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}
	for i, r := range restC {
		wantNext := dayOrder[(i+1)%7]
		if r.StartDay != dayOrder[i] {
			t.Errorf("L4SEC restriction %d: got StartDay %q, want %q", i, r.StartDay, dayOrder[i])
		}
		if r.EndDay != wantNext {
			t.Errorf("L4SEC restriction %d: got EndDay %q (should be day-after start due to midnight crossing), want %q", i, r.EndDay, wantNext)
		}
	}
	membersC, err := q.ListExtRotationMembers(ctx, "L4SEC")
	if err != nil {
		t.Fatalf("error loading L4SEC members: %s", err)
	}
	wantC := []string{"U4CAM", "U4DREW", "U4EVAN"}
	if len(membersC) != len(wantC) {
		t.Fatalf("L4SEC: expected %d members, got %d", len(wantC), len(membersC))
	}
	for i, m := range membersC {
		if m.UserID != wantC[i] {
			t.Errorf("L4SEC member %d: got %s, want %s", i, m.UserID, wantC[i])
		}
	}

	markAllForImport(t, ctx)
	assertGoldenTF(t, ctx)
}

// =============================================================================
// Scenario 5: One team, two schedules, an escalation policy that targets both
// schedules and a backstop user. Verifies dedup'd team membership + EP ordering.
// =============================================================================
func TestPagerDutyEscalationAcrossMultipleSchedules(t *testing.T) {
	ctx, pd := pdScenarioSetup(t)
	loadAll(t, ctx, pd, true)
	q := store.UseQueries(ctx)

	// Two rotations across two schedules.
	priRotations, err := q.ListExtRotationsByScheduleID(ctx, "S5PRI")
	if err != nil {
		t.Fatalf("error loading S5PRI rotations: %s", err)
	}
	if len(priRotations) != 1 || priRotations[0].ID != "L5PRI" {
		t.Fatalf("S5PRI: expected one rotation L5PRI, got %+v", priRotations)
	}
	if priRotations[0].Strategy != "weekly" {
		t.Errorf("L5PRI Strategy: got %q, want weekly", priRotations[0].Strategy)
	}

	mgrRotations, err := q.ListExtRotationsByScheduleID(ctx, "S5MGR")
	if err != nil {
		t.Fatalf("error loading S5MGR rotations: %s", err)
	}
	if len(mgrRotations) != 1 || mgrRotations[0].ID != "L5MGR" {
		t.Fatalf("S5MGR: expected one rotation L5MGR, got %+v", mgrRotations)
	}
	rmgr := mgrRotations[0]
	if rmgr.Strategy != "custom" {
		t.Errorf("L5MGR Strategy: got %q, want custom", rmgr.Strategy)
	}
	if rmgr.ShiftDuration != "PT2592000S" {
		t.Errorf("L5MGR ShiftDuration: got %q, want PT2592000S", rmgr.ShiftDuration)
	}
	restMgr, err := q.ListExtRotationRestrictions(ctx, "L5MGR")
	if err != nil {
		t.Fatalf("error loading L5MGR restrictions: %s", err)
	}
	if len(restMgr) != 0 {
		t.Errorf("L5MGR: expected 0 restrictions, got %d", len(restMgr))
	}

	// Team membership is dedup'd: each user appears exactly once for T5EP.
	memberships, err := q.ListExtTeamMemberships(ctx)
	if err != nil {
		t.Fatalf("error listing memberships: %s", err)
	}
	seen := map[string]int{}
	for _, m := range memberships {
		if m.ExtTeam.ID != "T5EP" {
			continue
		}
		seen[m.ExtUser.ID]++
	}
	wantUsers := []string{"U5AVERY", "U5BLAKE", "U5CAM", "U5DREW", "U5EVAN", "U5FREY", "U5GALE", "U5HARLAN"}
	for _, u := range wantUsers {
		if seen[u] != 1 {
			t.Errorf("user %s membership count: got %d, want 1", u, seen[u])
		}
	}

	// Escalation policy: three steps, in order, schedule/schedule/user.
	policies, err := q.ListExtEscalationPolicies(ctx)
	if err != nil {
		t.Fatalf("error loading escalation policies: %s", err)
	}
	if len(policies) != 1 || policies[0].ID != "EP5" {
		t.Fatalf("expected single policy EP5, got %+v", policies)
	}
	steps, err := q.ListExtEscalationPolicySteps(ctx, "EP5")
	if err != nil {
		t.Fatalf("error loading policy steps: %s", err)
	}
	if len(steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(steps))
	}
	wantSteps := []struct {
		id         string
		position   int64
		targetType string
		targetID   string
	}{
		{"EP5S1", 0, store.TARGET_TYPE_SCHEDULE, "S5PRI"},
		{"EP5S2", 1, store.TARGET_TYPE_SCHEDULE, "S5MGR"},
		{"EP5S3", 2, store.TARGET_TYPE_USER, "U5HARLAN"},
	}
	for i, want := range wantSteps {
		if steps[i].ID != want.id {
			t.Errorf("step %d: got id %s, want %s", i, steps[i].ID, want.id)
		}
		if steps[i].Position != want.position {
			t.Errorf("step %d: got Position %d, want %d", i, steps[i].Position, want.position)
		}
		targets, err := q.ListExtEscalationPolicyStepTargets(ctx, steps[i].ID)
		if err != nil {
			t.Fatalf("error loading targets for %s: %s", steps[i].ID, err)
		}
		if len(targets) != 1 {
			t.Fatalf("step %s: expected 1 target, got %d", steps[i].ID, len(targets))
		}
		if targets[0].TargetType != want.targetType {
			t.Errorf("step %s: got TargetType %q, want %q", steps[i].ID, targets[0].TargetType, want.targetType)
		}
		if targets[0].TargetID != want.targetID {
			t.Errorf("step %s: got TargetID %q, want %q", steps[i].ID, targets[0].TargetID, want.targetID)
		}
	}

	markAllForImport(t, ctx)
	// Mark the escalation policy for import so it shows up in Terraform.
	if err := q.MarkExtEscalationPolicyToImport(ctx, "EP5"); err != nil {
		t.Fatalf("error marking escalation policy to import: %s", err)
	}
	assertGoldenTF(t, ctx)
}
