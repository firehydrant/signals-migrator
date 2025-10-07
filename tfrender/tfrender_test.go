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

	if err := store.UseQueries(ctx).InsertExtScheduleV2(ctx, store.InsertExtScheduleV2Params{
		ID:               "id-for-ext-schedule-0",
		Name:             "Schedule 0",
		Description:      "Schedule 0 description",
		Timezone:         "UTC",
		TeamID:           "id-for-ext-team-1",
		SourceSystem:     "test",
		SourceScheduleID: "id-for-ext-schedule-0",
	}); err != nil {
		t.Fatal(err)
	}

	if err := store.UseQueries(ctx).InsertExtRotation(ctx, store.InsertExtRotationParams{
		ID:            "id-for-ext-rotation-0",
		ScheduleID:    "id-for-ext-schedule-0",
		Name:          "Daily Rotation",
		Strategy:      "daily",
		StartTime:     "11:00",
		HandoffTime:   "11:00",
		HandoffDay:    "wednesday",
		RotationOrder: 0,
	}); err != nil {
		t.Fatal(err)
	}

	if err := store.UseQueries(ctx).InsertExtRotationMember(ctx, store.InsertExtRotationMemberParams{
		RotationID:  "id-for-ext-rotation-0",
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

func TestRenderCustomStrategySchedule(t *testing.T) {
	ctx, tfr := tfrInit(t)

	// Create a team
	teamID := "team-1"
	teamExtID := "ext-team-1"
	if err := store.UseQueries(ctx).InsertFhTeam(ctx, store.InsertFhTeamParams{
		ID:   teamID,
		Name: "Test Team",
		Slug: "test-team",
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.UseQueries(ctx).InsertExtTeam(ctx, store.InsertExtTeamParams{
		ID:       teamExtID,
		Name:     "Test Team",
		Slug:     "test-team",
		FhTeamID: sql.NullString{String: teamID, Valid: true},
	}); err != nil {
		t.Fatal(err)
	}

	// Create a schedule with custom strategy rotation
	scheduleID := "schedule-1"
	if err := store.UseQueries(ctx).InsertExtScheduleV2(ctx, store.InsertExtScheduleV2Params{
		ID:               scheduleID,
		Name:             "Test Schedule",
		Description:      "Test Description",
		Timezone:         "America/Los_Angeles",
		TeamID:           teamExtID,
		SourceSystem:     "test",
		SourceScheduleID: scheduleID,
	}); err != nil {
		t.Fatal(err)
	}

	// Create a rotation with custom strategy
	rotationID := "rotation-1"
	if err := store.UseQueries(ctx).InsertExtRotation(ctx, store.InsertExtRotationParams{
		ID:            rotationID,
		ScheduleID:    scheduleID,
		Name:          "Custom Rotation",
		Description:   "Custom strategy rotation",
		Strategy:      "custom",
		ShiftDuration: "PT93600S",
		StartTime:     "2024-04-11T11:56:29-07:00",
		HandoffTime:   "",
		HandoffDay:    "",
		RotationOrder: 0,
	}); err != nil {
		t.Fatal(err)
	}

	// Generate Terraform
	if err := tfr.Write(ctx); err != nil {
		t.Fatal(err)
	}

	// Read the generated file
	content, err := os.ReadFile(tfr.Filepath())
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)

	// Verify that the rotation has start_time when using a custom strategy
	if !strings.Contains(contentStr, `start_time  = "2024-04-11T11:56:29-07:00"`) {
		t.Error("Rotation should have start_time for custom strategy")
	}

	// Verify that the rotation has custom strategy with shift_duration
	if !strings.Contains(contentStr, `type           = "custom"`) {
		t.Error("Rotation should have custom strategy type")
	}
	if !strings.Contains(contentStr, `shift_duration = "PT93600S"`) {
		t.Error("Rotation should have shift_duration for custom strategy")
	}

	// Verify that the rotation has start_time for custom strategy
	// Split the content to check rotation block specifically
	rotationBlockStart := strings.Index(contentStr, `resource "firehydrant_rotation"`)
	if rotationBlockStart != -1 {
		rotationBlockEnd := strings.Index(contentStr[rotationBlockStart:], "}")
		if rotationBlockEnd != -1 {
			rotationBlock := contentStr[rotationBlockStart : rotationBlockStart+rotationBlockEnd]
			if !strings.Contains(rotationBlock, `start_time  = "2024-04-11T11:56:29-07:00"`) {
				t.Error("Rotation should have start_time for custom strategy")
			}
		}
	}

	t.Logf("Generated Terraform:\n%s", contentStr)
}

func TestEscalationPolicyTargets(t *testing.T) {
	ctx, tfr := tfrInit(t)

	createUsers(t, ctx, "alice")
	createUsers(t, ctx, "bob")
	createTeams(t, ctx, "engineering", true)

	if err := store.UseQueries(ctx).InsertExtScheduleV2(ctx, store.InsertExtScheduleV2Params{
		ID:               "schedule-1",
		Name:             "Primary On-Call",
		Description:      "Primary on-call schedule",
		Timezone:         "America/New_York",
		TeamID:           "id-for-ext-team-engineering",
		SourceSystem:     "test",
		SourceScheduleID: "schedule-1",
	}); err != nil {
		t.Fatal(err)
	}

	if err := store.UseQueries(ctx).InsertExtEscalationPolicy(ctx, store.InsertExtEscalationPolicyParams{
		ID:          "ep-1",
		Name:        "Test Escalation Policy",
		Description: "Test escalation policy with targets",
		TeamID:      sql.NullString{String: "id-for-ext-team-engineering", Valid: true},
		RepeatLimit: 3,
		ToImport:    1, // Mark for import
	}); err != nil {
		t.Fatal(err)
	}

	if err := store.UseQueries(ctx).InsertExtEscalationPolicyStep(ctx, store.InsertExtEscalationPolicyStepParams{
		ID:                 "step-1",
		EscalationPolicyID: "ep-1",
		Position:           0,
		Timeout:            "PT15M",
	}); err != nil {
		t.Fatal(err)
	}

	targets := []struct {
		targetType string
		targetID   string
	}{
		{store.TARGET_TYPE_USER, "id-for-ext-user-alice"},
		{store.TARGET_TYPE_SCHEDULE, "schedule-1"},
	}

	for _, target := range targets {
		if err := store.UseQueries(ctx).InsertExtEscalationPolicyStepTarget(ctx, store.InsertExtEscalationPolicyStepTargetParams{
			EscalationPolicyStepID: "step-1",
			TargetType:             target.targetType,
			TargetID:               target.targetID,
		}); err != nil {
			t.Fatal(err)
		}
	}

	if err := tfr.Write(ctx); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(tfr.Filepath())
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, `targets {
      type = "User"
      id   = data.firehydrant_user.user_alice.id
    }`) {
		t.Error("Missing user target in escalation policy")
	}

	if !strings.Contains(contentStr, `targets {
      type = "OnCallSchedule"
      id   = firehydrant_on_call_schedule.team_engineering_slug_primary_on_call.id
    }`) {
		t.Error("Missing schedule target in escalation policy")
	}

	if !strings.Contains(contentStr, `resource "firehydrant_escalation_policy" "test_escalation_policy"`) {
		t.Error("Escalation policy resource not found")
	}

	if !strings.Contains(contentStr, `step {
    timeout = "PT15M"`) {
		t.Error("Escalation policy step not found")
	}
}
