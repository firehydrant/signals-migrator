package pager_test

import (
	"context"
	"testing"

	"github.com/firehydrant/signals-migrator/pager"
	"github.com/firehydrant/signals-migrator/store"
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
	})

	t.Run("LoadSchedules", func(t *testing.T) {
		t.Parallel()
		ctx, pd := setup(t)

		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		// At the moment, this will show "Team ... not found" warning in logs because
		// we didn't seed the database with that information. After we refactor the methods
		// ListTeams and ListUsers to use database, as LoadTeams and LoadUsers respectively,
		// we should expect the warning to go away.
		s, err := store.UseQueries(ctx).ListExtSchedules(ctx)
		if err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}
		t.Logf("found %d schedules", len(s))
		assertJSON(t, s)
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
