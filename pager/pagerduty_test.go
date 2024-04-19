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
		t.Parallel()

		ctx := withTestDB(t)
		ts := pagerProviderHttpServer(t)
		pd := pager.NewPagerDutyWithURL("api-key-very-secret", ts.URL)
		return ctx, pd
	}

	t.Run("ListUsers", func(t *testing.T) {
		ctx, pd := setup(t)

		u, err := pd.ListUsers(ctx)
		if err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		t.Logf("found %d users", len(u))
		assertJSON(t, u)
	})

	t.Run("LoadSchedules", func(t *testing.T) {
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
}
