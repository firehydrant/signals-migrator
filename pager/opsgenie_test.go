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

	t.Run("LoadSchedules", func(t *testing.T) {
		ctx, og := setup(t)

		if err := og.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		s, err := store.UseQueries(ctx).ListExtSchedules(ctx)
		if err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}
		t.Logf("found %d schedules", len(s))
		assertJSON(t, s)
	})
}
