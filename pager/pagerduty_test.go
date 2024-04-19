package pager_test

import (
	"context"
	"testing"

	"github.com/firehydrant/signals-migrator/pager"
	"github.com/firehydrant/signals-migrator/store"
)

func pagerDutyTestClient(t *testing.T, apiURL string) (context.Context, *pager.PagerDuty) {
	t.Helper()

	ctx := withTestDB(t)
	pd := pager.NewPagerDutyWithURL("api-key-very-secret", apiURL)
	return ctx, pd
}

func TestPagerDuty(t *testing.T) {
	ts := pagerProviderHttpServer(t)
	ctx, pd := pagerDutyTestClient(t, ts.URL)

	t.Run("ListUsers", func(t *testing.T) {
		u, err := pd.ListUsers(ctx)
		if err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		t.Logf("found %d users", len(u))
		assertJSON(t, u)
	})

	t.Run("LoadSchedules", func(t *testing.T) {
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
}
