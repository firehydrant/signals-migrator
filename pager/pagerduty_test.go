package pager_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/firehydrant/signals-migrator/pager"
	"github.com/firehydrant/signals-migrator/store"
	"github.com/firehydrant/signals-migrator/tfrender"
)

func TestPagerDuty(t *testing.T) {
	// What we're testing here is whether we manage to produce expected data in SQLite for the given API responses.
	// In other words, given responses within ./testdata/TestPagerDuty/apiserver, we expect the data transformed by
	// our implementation to be stored in the database to be consistent.
	// Asserting the content of database is a little tricky. As such, we encode the data into JSON and compare it
	// with the expected JSON in ./testdata/TestPagerDuty/[TestName].golden.json for each test case.

	// Avoid sharing setup code between tests to prevent test pollution in parallel execution.
	setup := func(t *testing.T) (context.Context, pager.Pager) {
		ctx := withTestDB(t)
		ts := pagerProviderHttpServer(t)
		pd := pager.NewPagerDutyWithURL("api-key-very-secret", ts.URL)
		return ctx, pd
	}

	t.Run("LoadUsers", func(t *testing.T) {
		t.Parallel()
		ctx, pd := setup(t)

		if err := pd.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}

		u, err := store.UseQueries(ctx).ListExtUsers(ctx)
		if err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		assertJSON(t, u)
	})

	// LoadTeams has 2 variants: one for literal teams and another for importing services as teams.
	// The "state" is maintained globally and as such should not be run in parallel.
	t.Run("LoadTeams", func(t *testing.T) {
		t.Run("loadTeams", func(t *testing.T) {
			ctx, pd := setup(t)

			if err := pd.UseTeamInterface("team"); err != nil {
				t.Fatalf("error setting team interface: %s", err)
			}
			if err := pd.LoadTeams(ctx); err != nil {
				t.Fatalf("error loading teams: %s", err)
			}
			teams, err := store.UseQueries(ctx).ListExtTeams(ctx)
			if err != nil {
				t.Fatalf("error loading teams: %s", err)
			}
			t.Logf("found %d teams", len(teams))
			assertJSON(t, teams)
		})

		t.Run("loadServices", func(t *testing.T) {
			ctx, pd := setup(t)

			if err := pd.UseTeamInterface("service"); err != nil {
				t.Fatalf("error setting team interface: %s", err)
			}
			if err := pd.LoadTeams(ctx); err != nil {
				t.Fatalf("error loading teams: %s", err)
			}
			teams, err := store.UseQueries(ctx).ListExtTeams(ctx)
			if err != nil {
				t.Fatalf("error loading teams: %s", err)
			}
			t.Logf("found %d teams, including services", len(teams))

			// We have 2 services:
			// - Endeavour, which has a team: Page Responder Team
			// - Server under Jack's desk, which has a team: Jack's team
			// We expect the membership association to link the users with the services, not their immediate team.
			assertJSON(t, teams)
		})
	})

	t.Run("LoadTeamMembers", func(t *testing.T) {
		t.Run("loadTeamMembers", func(t *testing.T) {
			ctx, pd := setup(t)

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
			members, err := store.UseQueries(ctx).ListExtTeamMemberships(ctx)
			if err != nil {
				t.Fatalf("error loading team members: %s", err)
			}
			t.Logf("found %d team members", len(members))
			assertJSON(t, members)
		})

		t.Run("loadTeamMembersPreservesOrder", func(t *testing.T) {
			ctx, pd := setup(t)

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
			members, err := store.UseQueries(ctx).ListExtTeamMemberships(ctx)
			if err != nil {
				t.Fatalf("error loading team members: %s", err)
			}

			// Verify that the order matches the PagerDuty API responses
			// For team PT54U20 (Jen): PXI6XNI, P2C9LBA
			// For team PV9JOXL (Service Catalog Team): PRXEEQ8, PXI6XNI

			// Find members for team PT54U20
			var pt54u20Members []string
			for _, member := range members {
				if member.ExtTeam.ID == "PT54U20" {
					pt54u20Members = append(pt54u20Members, member.ExtUser.ID)
				}
			}

			expectedPT54U20Order := []string{"PXI6XNI", "P2C9LBA"}
			if len(pt54u20Members) != len(expectedPT54U20Order) {
				t.Fatalf("team PT54U20: expected %d members, got %d", len(expectedPT54U20Order), len(pt54u20Members))
			}
			for i, userID := range pt54u20Members {
				if userID != expectedPT54U20Order[i] {
					t.Errorf("team PT54U20 member at position %d: expected %s, got %s", i, expectedPT54U20Order[i], userID)
				}
			}

			// Find members for team PV9JOXL
			var pv9joxlMembers []string
			for _, member := range members {
				if member.ExtTeam.ID == "PV9JOXL" {
					pv9joxlMembers = append(pv9joxlMembers, member.ExtUser.ID)
				}
			}

			expectedPV9JOXLOrder := []string{"PRXEEQ8", "PXI6XNI"}
			if len(pv9joxlMembers) != len(expectedPV9JOXLOrder) {
				t.Fatalf("team PV9JOXL: expected %d members, got %d", len(expectedPV9JOXLOrder), len(pv9joxlMembers))
			}
			for i, userID := range pv9joxlMembers {
				if userID != expectedPV9JOXLOrder[i] {
					t.Errorf("team PV9JOXL member at position %d: expected %s, got %s", i, expectedPV9JOXLOrder[i], userID)
				}
			}
		})

		t.Run("loadServiceMembers", func(t *testing.T) {
			ctx, pd := setup(t)

			if err := pd.UseTeamInterface("service"); err != nil {
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
			members, err := store.UseQueries(ctx).ListGroupExtTeamMemberships(ctx)
			if err != nil {
				t.Fatalf("error loading team members: %s", err)
			}
			t.Logf("found %d team members, including services", len(members))
			assertJSON(t, members)
		})

		t.Run("loadServiceMembersPreservesOrder", func(t *testing.T) {
			ctx, pd := setup(t)

			if err := pd.UseTeamInterface("service"); err != nil {
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
			members, err := store.UseQueries(ctx).ListGroupExtTeamMemberships(ctx)
			if err != nil {
				t.Fatalf("error loading team members: %s", err)
			}

			// Verify that the order matches the PagerDuty API responses
			// Services are loaded in order: P4XMRL3 (Endeavour), P3IIAF1 (Server under Jack's desk)

			// Extract service IDs in order
			var serviceIDs []string
			seen := make(map[string]bool)
			for _, member := range members {
				if !seen[member.ExtTeam.ID] {
					serviceIDs = append(serviceIDs, member.ExtTeam.ID)
					seen[member.ExtTeam.ID] = true
				}
			}

			expectedServiceOrder := []string{"P4XMRL3", "P3IIAF1"}
			if len(serviceIDs) != len(expectedServiceOrder) {
				t.Fatalf("expected %d services, got %d", len(expectedServiceOrder), len(serviceIDs))
			}
			for i, serviceID := range serviceIDs {
				if serviceID != expectedServiceOrder[i] {
					t.Errorf("service at position %d: expected %s, got %s", i, expectedServiceOrder[i], serviceID)
				}
			}
		})
	})

	t.Run("LoadSchedules", func(t *testing.T) {
		ctx, pd := setup(t)

		if err := pd.UseTeamInterface("team"); err != nil {
			t.Fatalf("error setting team interface: %s", err)
		}
		if err := pd.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := pd.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}
		schedules, err := store.UseQueries(ctx).ListExtSchedules(ctx)
		if err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}
		t.Logf("found %d schedules", len(schedules))
		assertJSON(t, schedules)
	})

	t.Run("LoadSchedulesPreservesMemberOrder", func(t *testing.T) {
		ctx, pd := setup(t)

		if err := pd.UseTeamInterface("team"); err != nil {
			t.Fatalf("error setting team interface: %s", err)
		}
		if err := pd.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := pd.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		// Check that schedule members are ordered correctly
		// For schedule P3D7DLW-PSQ0VRL, we expect users in order: PXI6XNI, P2C9LBA
		members, err := store.UseQueries(ctx).ListExtScheduleMembers(ctx, "P3D7DLW-PSQ0VRL")
		if err != nil {
			t.Fatalf("error loading schedule members: %s", err)
		}

		expectedOrder := []string{"PXI6XNI", "P2C9LBA"}
		if len(members) != len(expectedOrder) {
			t.Fatalf("schedule P3D7DLW-PSQ0VRL: expected %d members, got %d", len(expectedOrder), len(members))
		}
		for i, member := range members {
			if member.UserID != expectedOrder[i] {
				t.Errorf("schedule P3D7DLW-PSQ0VRL member at position %d: expected %s, got %s", i, expectedOrder[i], member.UserID)
			}
			if member.MemberOrder != int64(i) {
				t.Errorf("schedule P3D7DLW-PSQ0VRL member at position %d: expected order %d, got %d", i, i, member.MemberOrder)
			}
		}
	})

	t.Run("TerraformOutputMatchesScheduleOrder", func(t *testing.T) {
		ctx, pd := setup(t)

		if err := pd.UseTeamInterface("team"); err != nil {
			t.Fatalf("error setting team interface: %s", err)
		}
		if err := pd.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := pd.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		// First, verify the expected order from the database
		// For schedule P3D7DLW-PSQ0VRL, we expect users in order: PXI6XNI, P2C9LBA
		members, err := store.UseQueries(ctx).ListExtScheduleMembers(ctx, "P3D7DLW-PSQ0VRL")
		if err != nil {
			t.Fatalf("error loading schedule members: %s", err)
		}

		expectedOrder := []string{"PXI6XNI", "P2C9LBA"}
		if len(members) != len(expectedOrder) {
			t.Fatalf("schedule P3D7DLW-PSQ0VRL: expected %d members, got %d", len(expectedOrder), len(members))
		}

		// Verify database order
		for i, member := range members {
			if member.UserID != expectedOrder[i] {
				t.Errorf("schedule P3D7DLW-PSQ0VRL member at position %d: expected %s, got %s", i, expectedOrder[i], member.UserID)
			}
			if member.MemberOrder != int64(i) {
				t.Errorf("schedule P3D7DLW-PSQ0VRL member at position %d: expected order %d, got %d", i, i, member.MemberOrder)
			}
		}

		// Mark all teams for import to ensure they appear in Terraform
		teams, err := store.UseQueries(ctx).ListExtTeams(ctx)
		if err != nil {
			t.Fatalf("error listing teams: %s", err)
		}
		for _, team := range teams {
			if err := store.UseQueries(ctx).MarkExtTeamToImport(ctx, team.ID); err != nil {
				t.Fatalf("error marking team for import: %s", err)
			}
		}

		// Create FireHydrant teams for linking
		for _, team := range teams {
			fhTeamID := fmt.Sprintf("fh-team-%s", team.ID)
			if err := store.UseQueries(ctx).InsertFhTeam(ctx, store.InsertFhTeamParams{
				ID:   fhTeamID,
				Name: team.Name,
				Slug: fmt.Sprintf("fh-team-%s", team.ID),
			}); err != nil {
				t.Fatalf("error creating FireHydrant team: %s", err)
			}
			if err := store.UseQueries(ctx).LinkExtTeam(ctx, store.LinkExtTeamParams{
				ID:       team.ID,
				FhTeamID: sql.NullString{String: fhTeamID, Valid: true},
			}); err != nil {
				t.Fatalf("error linking team: %s", err)
			}
		}

		// Mark all users for import to ensure they appear in Terraform
		users, err := store.UseQueries(ctx).ListExtUsers(ctx)
		if err != nil {
			t.Fatalf("error listing users: %s", err)
		}
		for _, user := range users {
			fhUserID := fmt.Sprintf("fh-user-%s", user.ID)
			if err := store.UseQueries(ctx).InsertFhUser(ctx, store.InsertFhUserParams{
				ID:    fhUserID,
				Email: user.Email,
				Name:  user.Name,
			}); err != nil {
				t.Fatalf("error creating FireHydrant user: %s", err)
			}
			if err := store.UseQueries(ctx).LinkExtUser(ctx, store.LinkExtUserParams{
				ID:       user.ID,
				FhUserID: sql.NullString{String: fhUserID, Valid: true},
			}); err != nil {
				t.Fatalf("error linking user: %s", err)
			}
		}

		// Generate Terraform file
		tfFile := filepath.Join(t.TempDir(), "schedule_order_verification.tf")
		tfr, err := tfrender.New(tfFile)
		if err != nil {
			t.Fatalf("error creating Terraform renderer: %s", err)
		}

		if err := tfr.Write(ctx); err != nil {
			t.Fatalf("error writing Terraform file: %s", err)
		}

		// Read the generated Terraform file
		tfContent, err := os.ReadFile(tfFile)
		if err != nil {
			t.Fatalf("error reading Terraform file: %s", err)
		}

		tfContentStr := string(tfContent)
		t.Logf("Generated Terraform file:\n%s", tfContentStr)

		// Extract the on-call schedule resource for P3D7DLW-PSQ0VRL
		// Look for the schedule resource that contains the expected member order
		scheduleRegex := regexp.MustCompile(`resource "firehydrant_on_call_schedule" "[^"]*" \{[\s\S]*?\}`)
		scheduleMatches := scheduleRegex.FindAllString(tfContentStr, -1)

		if len(scheduleMatches) == 0 {
			t.Fatal("No firehydrant_on_call_schedule resources found in Terraform file")
		}

		// Find the schedule that contains our expected users
		// Looking for the schedule that contains both acme_success_eng (P2C9LBA) and acme_eng (PXI6XNI)
		var targetSchedule string
		for _, schedule := range scheduleMatches {
			if strings.Contains(schedule, "acme_success_eng") && strings.Contains(schedule, "acme_eng") {
				targetSchedule = schedule
				break
			}
		}

		if targetSchedule == "" {
			t.Fatal("No schedule found containing both acme_success_eng and acme_eng users")
		}

		t.Logf("Found target schedule:\n%s", targetSchedule)

		// Extract the member_ids array from the Terraform resource
		memberIDsRegex := regexp.MustCompile(`member_ids = \[([^\]]*)\]`)
		memberIDsMatch := memberIDsRegex.FindStringSubmatch(targetSchedule)

		if len(memberIDsMatch) < 2 {
			t.Fatal("No member_ids found in schedule resource")
		}

		memberIDsStr := memberIDsMatch[1]
		t.Logf("Member IDs string: %s", memberIDsStr)

		// Extract individual user references from the member_ids array
		userRefRegex := regexp.MustCompile(`data\.firehydrant_user\.([^\.]+)\.id`)
		userRefMatches := userRefRegex.FindAllStringSubmatch(memberIDsStr, -1)

		if len(userRefMatches) != len(expectedOrder) {
			t.Errorf("Expected %d user references in Terraform, got %d", len(expectedOrder), len(userRefMatches))
		}

		// Convert the Terraform user slugs back to the original PagerDuty user IDs
		var terraformUserOrder []string
		for _, userRefMatch := range userRefMatches {
			terraformUserSlug := userRefMatch[1]
			// Map the Terraform slugs to PagerDuty user IDs
			switch terraformUserSlug {
			case "acme_eng":
				terraformUserOrder = append(terraformUserOrder, "PXI6XNI")
			case "acme_success_eng":
				terraformUserOrder = append(terraformUserOrder, "P2C9LBA")
			default:
				t.Errorf("Unknown user slug in Terraform: %s", terraformUserSlug)
			}
		}

		t.Logf("Expected order from PagerDuty API: %v", expectedOrder)
		t.Logf("Actual order in Terraform file: %v", terraformUserOrder)

		// Verify that the order matches
		if len(terraformUserOrder) != len(expectedOrder) {
			t.Errorf("Order length mismatch: expected %d, got %d", len(expectedOrder), len(terraformUserOrder))
		} else {
			for i, userID := range terraformUserOrder {
				if userID != expectedOrder[i] {
					t.Errorf("Order mismatch at position %d: expected %s, got %s", i, expectedOrder[i], userID)
				}
			}
		}

		// Additional verification: Check that the schedule resource contains the correct structure
		if !strings.Contains(targetSchedule, "firehydrant_on_call_schedule") {
			t.Error("Schedule resource should contain firehydrant_on_call_schedule")
		}

		if !strings.Contains(targetSchedule, "member_ids") {
			t.Error("Schedule resource should contain member_ids")
		}

		t.Logf("✅ Terraform output order verification completed successfully")
	})

	t.Run("LoadEscalationPolicies", func(t *testing.T) {
		t.Parallel()
		ctx, pd := setup(t)
		data := map[string]any{}

		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		if err := pd.LoadEscalationPolicies(ctx); err != nil {
			t.Fatalf("error loading escalation policies: %s", err)
		}

		ep, err := store.UseQueries(ctx).ListExtEscalationPolicies(ctx)
		if err != nil {
			t.Fatalf("error loading escalation policies: %s", err)
		}
		data["escalation_policies"] = ep

		data["escalation_policy_steps"] = []any{}
		data["escalation_policy_step_targets"] = []any{}
		for _, p := range ep {
			steps, err := store.UseQueries(ctx).ListExtEscalationPolicySteps(ctx, p.ID)
			if err != nil {
				t.Fatalf("error loading escalation policy steps: %s", err)
			}
			for _, s := range steps {
				data["escalation_policy_steps"] = append(data["escalation_policy_steps"].([]any), s)
				targets, err := store.UseQueries(ctx).ListExtEscalationPolicyStepTargets(ctx, s.ID)
				if err != nil {
					t.Fatalf("error loading escalation policy step targets for '%s': %s", s.ID, err)
				}
				for _, t := range targets {
					data["escalation_policy_step_targets"] = append(data["escalation_policy_step_targets"].([]any), t)
				}
			}
		}

		assertJSON(t, data)
	})
}
