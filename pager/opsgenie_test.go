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
}
