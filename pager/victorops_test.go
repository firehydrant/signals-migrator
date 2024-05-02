package pager_test

import (
	"context"
	"testing"

	"github.com/firehydrant/signals-migrator/pager"
	"github.com/firehydrant/signals-migrator/store"
)

func TestVictorOps(t *testing.T) {
	// Avoid sharing setup code between tests to prevent test pollution in parallel execution.
	setup := func(t *testing.T) (context.Context, pager.Pager) {
		ctx := withTestDB(t)
		ts := pagerProviderHttpServer(t)
		vo := pager.NewVictorOpsWithURL("api-key-very-secret", "app-id-maybe-secret", ts.URL)
		return ctx, vo
	}

	t.Run("LoadUsers", func(t *testing.T) {
		t.Parallel()
		ctx, vo := setup(t)

		if err := vo.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		u, err := store.UseQueries(ctx).ListExtUsers(ctx)
		if err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		assertJSON(t, u)
	})

	t.Run("LoadTeams", func(t *testing.T) {
		t.Parallel()
		ctx, vo := setup(t)

		if err := vo.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		teams, err := store.UseQueries(ctx).ListExtTeams(ctx)
		if err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		assertJSON(t, teams)
	})

	t.Run("LoadTeamMembers", func(t *testing.T) {
		t.Parallel()
		ctx, vo := setup(t)

		if err := vo.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := vo.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		if err := vo.LoadTeamMembers(ctx); err != nil {
			t.Fatalf("error loading team members: %s", err)
		}
		members, err := store.UseQueries(ctx).ListExtTeamMemberships(ctx)
		if err != nil {
			t.Fatalf("error loading team members: %s", err)
		}
		assertJSON(t, members)
	})
}
