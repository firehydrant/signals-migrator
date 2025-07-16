package tfrender_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/firehydrant/signals-migrator/store"
	"github.com/firehydrant/signals-migrator/tfrender"
	"gotest.tools/v3/golden"
)

func tfrInit(t *testing.T) (context.Context, *tfrender.TFRender) {
	t.Helper()

	ctx := store.WithContext(context.Background())
	t.Cleanup(func() { store.FromContext(ctx).Close() })

	tfr, err := tfrender.New(filepath.Join(t.TempDir(), t.Name()+".tf"))
	if err != nil {
		t.Fatal(err)
	}
	return ctx, tfr
}

func goldenFile(name string) string {
	ext := filepath.Ext(name)
	return fmt.Sprintf("%s.golden%s", name[:len(name)-len(ext)], ext)
}

func assertRenderPager(t *testing.T) {
	seedFile := fmt.Sprintf("%s_seed.sql", t.Name())
	seed, err := os.ReadFile(filepath.Join("testdata", seedFile))
	if err != nil {
		t.Fatal(err)
	}

	ctx, tfr := tfrInit(t)
	sql := strings.TrimSpace(string(seed))
	if _, err := store.FromContext(ctx).ExecContext(ctx, sql); err != nil {
		t.Fatal(err)
	}
	if err := tfr.Write(ctx); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(tfr.Filepath())
	if err != nil {
		t.Fatal(err)
	}

	golden.Assert(t, string(content), filepath.Join(filepath.Dir(t.Name()), goldenFile(tfr.Filename())))
}

func createUsers(t *testing.T, ctx context.Context, variant string) {
	t.Helper()
	id := fmt.Sprintf("id-for-user-%s", variant)
	extID := fmt.Sprintf("id-for-ext-user-%s", variant)
	email := fmt.Sprintf("user-%s@example.com", variant)
	name := fmt.Sprintf("User %s", variant)

	if err := store.UseQueries(ctx).InsertFhUser(ctx, store.InsertFhUserParams{
		ID:    id,
		Email: email,
		Name:  name,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.UseQueries(ctx).InsertExtUser(ctx, store.InsertExtUserParams{
		ID:       extID,
		Email:    email,
		Name:     name,
		FhUserID: sql.NullString{String: id, Valid: true},
	}); err != nil {
		t.Fatal(err)
	}
}

func createTeams(t *testing.T, ctx context.Context, variant string, withFhTeam bool) {
	t.Helper()
	id := fmt.Sprintf("id-for-team-%s", variant)
	slug := fmt.Sprintf("team-%s-slug", variant)
	extID := fmt.Sprintf("id-for-ext-team-%s", variant)
	name := fmt.Sprintf("Team %s", variant)

	if err := store.UseQueries(ctx).InsertFhTeam(ctx, store.InsertFhTeamParams{
		ID:   id,
		Name: name,
		Slug: slug,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.UseQueries(ctx).InsertExtTeam(ctx, store.InsertExtTeamParams{
		ID:       extID,
		Name:     name,
		Slug:     slug,
		FhTeamID: sql.NullString{String: id, Valid: withFhTeam},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.UseQueries(ctx).MarkExtTeamToImport(ctx, extID); err != nil {
		t.Fatal(err)
	}
}

func TestRenderDataUser(t *testing.T) {
	ctx, tfr := tfrInit(t)
	for i := range 3 {
		createUsers(t, ctx, strconv.Itoa(i))
	}

	if err := tfr.Write(ctx); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(tfr.Filepath())
	if err != nil {
		t.Fatal(err)
	}

	// Expect output to have 3 user blocks.
	golden.Assert(t, string(content), goldenFile(tfr.Filename()))
}

func TestRenderTeamResource(t *testing.T) {
	ctx, tfr := tfrInit(t)
	for i := range 4 {
		createUsers(t, ctx, strconv.Itoa(i))
	}
	for i := range 4 {
		// team0 and team2 refers to existing FireHydrant teams,
		// so they will have import {} block associated to them.
		createTeams(t, ctx, strconv.Itoa(i), i%2 == 0)
	}

	if err := store.UseQueries(ctx).InsertExtMembership(ctx, store.InsertExtMembershipParams{
		UserID: "id-for-ext-user-0",
		TeamID: "id-for-ext-team-0",
	}); err != nil {
		t.Fatal(err)
	}

	if err := tfr.Write(ctx); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(tfr.Filepath())
	if err != nil {
		t.Fatal(err)
	}

	// Expect output to have:
	// - 4 user blocks
	// - 4 team blocks, where:
	//   - team 0 and team 2 have import {} block
	//   - team 0 has 1 user membership, linked with user 0
	// Updated expectation: Teams now use memberships_input with proper HCL traversals
	golden.Assert(t, string(content), goldenFile(tfr.Filename()))
}

func TestRenderOnCallScheduleResource(t *testing.T) {
	ctx, tfr := tfrInit(t)
	for i := range 4 {
		createUsers(t, ctx, strconv.Itoa(i))
	}
	for i := range 4 {
		// team0 and team2 refers to existing FireHydrant teams,
		// so they will have import {} block associated to them.
		createTeams(t, ctx, strconv.Itoa(i), i%2 == 0)
	}

	if err := store.UseQueries(ctx).InsertExtSchedule(ctx, store.InsertExtScheduleParams{
		ID:          "id-for-ext-schedule-0",
		Name:        "Schedule 0",
		Description: "Schedule 0 description",
		HandoffTime: "11:00",
		HandoffDay:  "wednesday",
		Strategy:    "daily",
		Timezone:    "UTC",
	}); err != nil {
		t.Fatal(err)
	}

	if err := store.UseQueries(ctx).InsertExtScheduleTeam(ctx, store.InsertExtScheduleTeamParams{
		ScheduleID: "id-for-ext-schedule-0",
		TeamID:     "id-for-ext-team-1",
	}); err != nil {
		t.Fatal(err)
	}

	if err := store.UseQueries(ctx).InsertExtScheduleMember(ctx, store.InsertExtScheduleMemberParams{
		ScheduleID:  "id-for-ext-schedule-0",
		UserID:      "id-for-ext-user-0",
		MemberOrder: 0,
	}); err != nil {
		t.Fatal(err)
	}

	if err := tfr.Write(ctx); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(tfr.Filepath())
	if err != nil {
		t.Fatal(err)
	}

	// Expect output to have:
	// - 4 user blocks
	// - 4 team blocks, where:
	//   - team 0 and team 2 have import {} block
	//   - team 0 has 1 user membership, linked with user 0
	// - 1 on-call schedule block, linked with team 1
	// Updated expectation: Schedule now uses firehydrant_signals_api_on_call_schedule
	// and uses members_input with proper HCL traversals, strategy_input as object
	golden.Assert(t, string(content), goldenFile(tfr.Filename()))
}

func TestRenderOnCallScheduleResourceWithRestrictions(t *testing.T) {
	ctx, tfr := tfrInit(t)
	for i := range 2 {
		createUsers(t, ctx, strconv.Itoa(i))
	}
	for i := range 2 {
		createTeams(t, ctx, strconv.Itoa(i), false)
	}

	if err := store.UseQueries(ctx).InsertExtSchedule(ctx, store.InsertExtScheduleParams{
		ID:          "id-for-ext-schedule-0",
		Name:        "Schedule 0",
		Description: "Schedule 0 description",
		HandoffTime: "09:00",
		HandoffDay:  "monday",
		Strategy:    "weekly",
		Timezone:    "America/New_York",
	}); err != nil {
		t.Fatal(err)
	}

	if err := store.UseQueries(ctx).InsertExtScheduleTeam(ctx, store.InsertExtScheduleTeamParams{
		ScheduleID: "id-for-ext-schedule-0",
		TeamID:     "id-for-ext-team-0",
	}); err != nil {
		t.Fatal(err)
	}

	if err := store.UseQueries(ctx).InsertExtScheduleMember(ctx, store.InsertExtScheduleMemberParams{
		ScheduleID:  "id-for-ext-schedule-0",
		UserID:      "id-for-ext-user-0",
		MemberOrder: 0,
	}); err != nil {
		t.Fatal(err)
	}

	if err := store.UseQueries(ctx).InsertExtScheduleMember(ctx, store.InsertExtScheduleMemberParams{
		ScheduleID:  "id-for-ext-schedule-0",
		UserID:      "id-for-ext-user-1",
		MemberOrder: 1,
	}); err != nil {
		t.Fatal(err)
	}

	// Add some restrictions
	if err := store.UseQueries(ctx).InsertExtScheduleRestriction(ctx, store.InsertExtScheduleRestrictionParams{
		ScheduleID:       "id-for-ext-schedule-0",
		RestrictionIndex: "0", // Add index to avoid constraint violation
		StartDay:         "monday",
		StartTime:        "09:00",
		EndDay:           "friday",
		EndTime:          "17:00",
	}); err != nil {
		t.Fatal(err)
	}

	if err := store.UseQueries(ctx).InsertExtScheduleRestriction(ctx, store.InsertExtScheduleRestrictionParams{
		ScheduleID:       "id-for-ext-schedule-0",
		RestrictionIndex: "1", // Add index to avoid constraint violation
		StartDay:         "saturday",
		StartTime:        "10:00",
		EndDay:           "sunday",
		EndTime:          "16:00",
	}); err != nil {
		t.Fatal(err)
	}

	if err := tfr.Write(ctx); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(tfr.Filepath())
	if err != nil {
		t.Fatal(err)
	}

	// Expect output to have:
	// - 2 user blocks
	// - 2 team blocks
	// - 1 on-call schedule block with:
	//   - members_input with 2 users
	//   - strategy_input with weekly strategy
	//   - restrictions_input with 2 restrictions
	golden.Assert(t, string(content), goldenFile(tfr.Filename()))
}

func TestRenderEscalationPolicyResource(t *testing.T) {
	ctx, tfr := tfrInit(t)
	for i := range 2 {
		createUsers(t, ctx, strconv.Itoa(i))
	}
	for i := range 2 {
		createTeams(t, ctx, strconv.Itoa(i), false)
	}

	// Create a schedule first
	if err := store.UseQueries(ctx).InsertExtSchedule(ctx, store.InsertExtScheduleParams{
		ID:          "id-for-ext-schedule-0",
		Name:        "Schedule 0",
		Description: "Schedule 0 description",
		HandoffTime: "09:00",
		HandoffDay:  "monday",
		Strategy:    "weekly",
		Timezone:    "UTC",
	}); err != nil {
		t.Fatal(err)
	}

	if err := store.UseQueries(ctx).InsertExtScheduleTeam(ctx, store.InsertExtScheduleTeamParams{
		ScheduleID: "id-for-ext-schedule-0",
		TeamID:     "id-for-ext-team-0",
	}); err != nil {
		t.Fatal(err)
	}

	// Create escalation policy
	if err := store.UseQueries(ctx).InsertExtEscalationPolicy(ctx, store.InsertExtEscalationPolicyParams{
		ID:          "id-for-ext-policy-0",
		Name:        "Policy 0",
		Description: "Policy 0 description",
		TeamID:      sql.NullString{String: "id-for-ext-team-0", Valid: true},
		RepeatLimit: 3,
	}); err != nil {
		t.Fatal(err)
	}

	// Create escalation policy step
	if err := store.UseQueries(ctx).InsertExtEscalationPolicyStep(ctx, store.InsertExtEscalationPolicyStepParams{
		ID:                 "id-for-ext-step-0",
		EscalationPolicyID: "id-for-ext-policy-0",
		Position:           1,
		Timeout:            "PT5M",
	}); err != nil {
		t.Fatal(err)
	}

	// Add user target
	if err := store.UseQueries(ctx).InsertExtEscalationPolicyStepTarget(ctx, store.InsertExtEscalationPolicyStepTargetParams{
		EscalationPolicyStepID: "id-for-ext-step-0",
		TargetID:               "id-for-ext-user-0",
		TargetType:             "user",
	}); err != nil {
		t.Fatal(err)
	}

	// Add schedule target
	if err := store.UseQueries(ctx).InsertExtEscalationPolicyStepTarget(ctx, store.InsertExtEscalationPolicyStepTargetParams{
		EscalationPolicyStepID: "id-for-ext-step-0",
		TargetID:               "id-for-ext-schedule-0",
		TargetType:             "schedule",
	}); err != nil {
		t.Fatal(err)
	}

	if err := tfr.Write(ctx); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(tfr.Filepath())
	if err != nil {
		t.Fatal(err)
	}

	// Expect output to have:
	// - 2 user blocks
	// - 2 team blocks
	// - 1 on-call schedule block (firehydrant_signals_api_on_call_schedule)
	// - 1 escalation policy block (firehydrant_signals_api_escalation_policy)
	//   with references to user and schedule targets
	golden.Assert(t, string(content), goldenFile(tfr.Filename()))
}

func TestRenderComplexScenario(t *testing.T) {
	ctx, tfr := tfrInit(t)

	// Create multiple users
	for i := range 5 {
		createUsers(t, ctx, strconv.Itoa(i))
	}

	// Create multiple teams
	for i := range 3 {
		createTeams(t, ctx, strconv.Itoa(i), i == 0) // only team 0 has existing FH team
	}

	// Add team memberships
	for i := range 3 {
		if err := store.UseQueries(ctx).InsertExtMembership(ctx, store.InsertExtMembershipParams{
			UserID: fmt.Sprintf("id-for-ext-user-%d", i),
			TeamID: fmt.Sprintf("id-for-ext-team-%d", i%2), // distribute across teams 0 and 1
		}); err != nil {
			t.Fatal(err)
		}
	}

	// Create schedules
	for i := range 2 {
		if err := store.UseQueries(ctx).InsertExtSchedule(ctx, store.InsertExtScheduleParams{
			ID:          fmt.Sprintf("id-for-ext-schedule-%d", i),
			Name:        fmt.Sprintf("Schedule %d", i),
			Description: fmt.Sprintf("Schedule %d description", i),
			HandoffTime: "09:00",
			HandoffDay:  "monday",
			Strategy:    "weekly",
			Timezone:    "UTC",
		}); err != nil {
			t.Fatal(err)
		}

		if err := store.UseQueries(ctx).InsertExtScheduleTeam(ctx, store.InsertExtScheduleTeamParams{
			ScheduleID: fmt.Sprintf("id-for-ext-schedule-%d", i),
			TeamID:     fmt.Sprintf("id-for-ext-team-%d", i),
		}); err != nil {
			t.Fatal(err)
		}

		// Add members to schedules
		for j := range 2 {
			if err := store.UseQueries(ctx).InsertExtScheduleMember(ctx, store.InsertExtScheduleMemberParams{
				ScheduleID:  fmt.Sprintf("id-for-ext-schedule-%d", i),
				UserID:      fmt.Sprintf("id-for-ext-user-%d", j+(i*2)),
				MemberOrder: int64(j),
			}); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err := tfr.Write(ctx); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(tfr.Filepath())
	if err != nil {
		t.Fatal(err)
	}

	// Expect output to have:
	// - 5 user blocks
	// - 3 team blocks with team 0 having import block
	// - 2 on-call schedule blocks using V2 schema with proper HCL traversals
	// - Team memberships using memberships_input with HCL traversals
	// - Schedule members using members_input with HCL traversals
	// - Schedule strategy using strategy_input as object
	golden.Assert(t, string(content), goldenFile(tfr.Filename()))
}
