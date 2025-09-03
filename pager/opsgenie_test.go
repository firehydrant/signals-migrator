package pager_test

import (
	"context"
	"strings"
	"testing"

	"github.com/firehydrant/signals-migrator/pager"
	"github.com/firehydrant/signals-migrator/store"
)

func TestOpsgenie(t *testing.T) {
	// What we're testing here is whether we manage to produce expected data in SQLite for the given API responses.
	// In other words, given responses within ./testdata/TestOpsgenie/apiserver, we expect the data transformed by
	// our implementation to be stored in the database to be consistent.
	// Asserting the content of database is a little tricky. As such, we encode the data into JSON and compare it
	// with the expected JSON in ./testdata/TestOpsgenie/[TestName].golden.json for each test case.

	// Avoid sharing setup code between tests to prevent test pollution in parallel execution.
	setup := func(t *testing.T) (context.Context, pager.Pager) {
		t.Parallel()

		ctx := withTestDB(t)
		ts := pagerProviderHttpServer(t)
		og := pager.NewOpsgenieWithURL("api-key-very-secret", strings.TrimPrefix(ts.URL, "http://"))
		return ctx, og
	}

	t.Run("LoadUsers", func(t *testing.T) {
		ctx, og := setup(t)

		if err := og.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}

		u, err := store.UseQueries(ctx).ListExtUsers(ctx)
		if err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		assertJSON(t, u)
	})

	t.Run("LoadTeams", func(t *testing.T) {
		ctx, og := setup(t)

		if err := og.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}

		u, err := store.UseQueries(ctx).ListExtTeams(ctx)
		if err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		assertJSON(t, u)
	})

	t.Run("LoadTeamMembers", func(t *testing.T) {
		ctx, og := setup(t)

		if err := og.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := og.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		if err := og.LoadTeamMembers(ctx); err != nil {
			t.Fatalf("error loading team members: %s", err)
		}

		u, err := store.UseQueries(ctx).ListExtTeamMemberships(ctx)
		if err != nil {
			t.Fatalf("error loading team memberships: %s", err)
		}
		assertJSON(t, u)
	})

	t.Run("LoadSchedules", func(t *testing.T) {
		ctx, og := setup(t)

		// Load teams first since schedules reference teams
		if err := og.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}

		if err := og.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		s, err := store.UseQueries(ctx).ListExtSchedulesV2(ctx)
		if err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}
		t.Logf("found %d schedules", len(s))
		assertJSON(t, s)
	})

	t.Run("LoadRotations", func(t *testing.T) {
		ctx, og := setup(t)

		// Load teams first since schedules reference teams
		if err := og.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}

		if err := og.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		// Get rotations for the schedule
		scheduleID := "3fee43f2-02da-49be-ab50-c88ed13aecc3"
		rotations, err := store.UseQueries(ctx).ListExtRotationsByScheduleID(ctx, scheduleID)
		if err != nil {
			t.Fatalf("error loading rotations: %s", err)
		}
		t.Logf("found %d rotations", len(rotations))
		assertJSON(t, rotations)
	})

	t.Run("LoadSchedulesPreservesMemberOrder", func(t *testing.T) {
		ctx, og := setup(t)

		if err := og.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := og.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		if err := og.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		// Check that rotation members are ordered correctly
		// For rotation b1b5f7f6-728b-47bd-af02-c3e1c33bf219
		// we expect users in order based on the API response
		rotationID := "b1b5f7f6-728b-47bd-af02-c3e1c33bf219"
		members, err := store.UseQueries(ctx).ListExtRotationMembers(ctx, rotationID)
		if err != nil {
			t.Fatalf("error loading rotation members: %s", err)
		}

		// Verify that members are ordered by member_order
		for i, member := range members {
			if member.MemberOrder != int64(i) {
				t.Errorf("rotation %s member at position %d: expected order %d, got %d", rotationID, i, i, member.MemberOrder)
			}
		}

		// Verify that we have at least one member and the order is sequential
		if len(members) > 0 {
			for i := 1; i < len(members); i++ {
				if members[i].MemberOrder <= members[i-1].MemberOrder {
					t.Errorf("rotation %s: member order is not sequential at position %d", rotationID, i)
				}
			}
		}
	})

	t.Run("LoadEscalationPolicies", func(t *testing.T) {
		ctx, og := setup(t)

		if err := og.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := og.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		if err := og.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		if err := og.LoadEscalationPolicies(ctx); err != nil {
			t.Fatalf("error loading escalation policies: %s", err)
		}

		s, err := store.UseQueries(ctx).ListExtEscalationPolicies(ctx)
		if err != nil {
			t.Fatalf("error loading escalation policies: %s", err)
		}
		t.Logf("found %d escalation policies", len(s))
		assertJSON(t, s)
	})

	t.Run("LoadEscalationPoliciesWithTargets", func(t *testing.T) {
		ctx, og := setup(t)

		if err := og.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := og.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		if err := og.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		if err := og.LoadEscalationPolicies(ctx); err != nil {
			t.Fatalf("error loading escalation policies: %s", err)
		}

		policies, err := store.UseQueries(ctx).ListExtEscalationPolicies(ctx)
		if err != nil {
			t.Fatalf("error loading escalation policies: %s", err)
		}

		// Policy should have both user and schedule targets
		var targetPolicy *store.ExtEscalationPolicy
		for _, policy := range policies {
			if policy.Name == "Customer Success_escalation" {
				targetPolicy = &policy
				break
			}
		}

		if targetPolicy == nil {
			t.Fatal("Could not find 'Customer Success_escalation' policy")
		}

		steps, err := store.UseQueries(ctx).ListExtEscalationPolicySteps(ctx, targetPolicy.ID)
		if err != nil {
			t.Fatalf("error loading escalation policy steps: %s", err)
		}

		if len(steps) != 2 {
			t.Fatalf("Expected 2 steps, got %d", len(steps))
		}

		for i, step := range steps {
			targets, err := store.UseQueries(ctx).ListExtEscalationPolicyStepTargets(ctx, step.ID)
			if err != nil {
				t.Fatalf("error loading targets for step %d: %s", i, err)
			}

			if len(targets) != 1 {
				t.Fatalf("Expected 1 target for step %d, got %d", i, len(targets))
			}

			target := targets[0]
			switch i {
			case 0:
				if target.TargetType != store.TARGET_TYPE_USER {
					t.Errorf("Step 0: Expected target type USER, got %s", target.TargetType)
				}
				if target.TargetID != "b5b92115-bfe7-43eb-8c2a-e467f2e5ddc4" {
					t.Errorf("Step 0: Expected user target ID 'b5b92115-bfe7-43eb-8c2a-e467f2e5ddc4', got %s", target.TargetID)
				}
			case 1:
				if target.TargetType != store.TARGET_TYPE_SCHEDULE {
					t.Errorf("Step 1: Expected target type SCHEDULE, got %s", target.TargetType)
				}
				if target.TargetID != "3fee43f2-02da-49be-ab50-c88ed13aecc3" {
					t.Errorf("Step 1: Expected schedule target ID '3fee43f2-02da-49be-ab50-c88ed13aecc3', got %s", target.TargetID)
				}

				schedule, err := store.UseQueries(ctx).GetExtScheduleV2(ctx, target.TargetID)
				if err != nil {
					t.Errorf("Step 1: Could not resolve schedule target '%s': %s", target.TargetID, err)
				} else {
					t.Logf("Step 1: Successfully resolved schedule target to '%s'", schedule.Name)
				}
			}
		}

		t.Logf("✅ Escalation policy targeting verification completed successfully")
	})
}
