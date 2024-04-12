package pager_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/firehydrant/signals-migrator/pager"
	"github.com/firehydrant/signals-migrator/store"
	"github.com/gosimple/slug"
)

func pagerDutyHttpServer(t *testing.T) *httptest.Server {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filename, err := url.JoinPath("testdata", t.Name(), slug.Make(r.URL.Path)+".json")
		if err != nil {
			t.Fatalf("error joining path for expected response: %s", err)
		}
		if _, err := os.Stat(filename); err != nil {
			t.Logf("file not found: %s", filename)
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filename)
	})
	return httptest.NewServer(h)
}

func pagerDutyTestClient(t *testing.T, apiURL string) (context.Context, *pager.PagerDuty) {
	ctx := withTestDB(t)
	pd := pager.NewPagerDutyWithURL("api-key-very-secret", apiURL)
	return ctx, pd
}

func assertDeepEqual[T any](t *testing.T, got, want T) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("[FAIL]\n got: %+v\nwant: %+v", got, want)
	}
}

func TestPagerDuty(t *testing.T) {
	ts := pagerDutyHttpServer(t)
	ctx, pd := pagerDutyTestClient(t, ts.URL)

	t.Run("ListUsers", func(t *testing.T) {
		u, err := pd.ListUsers(ctx)
		if err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}
		t.Logf("found %d users", len(u))

		// Verify that the first user is as expected.
		mika := u[0]
		expected := &pager.User{
			Email: "mika+eng@example.com",
			Resource: pager.Resource{
				ID:   "P5A1XH2",
				Name: "Mika",
			},
		}
		assertDeepEqual(t, mika, expected)
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

		// Verify that the first schedule is as expected.
		first := s[0]
		expected := store.ExtSchedule{
			ID:            "P3D7DLW-PC1DX4O",
			Name:          "Jen - primary - Layer 2",
			Description:   "(Layer 2)",
			Timezone:      "America/Los_Angeles",
			Strategy:      "weekly",
			ShiftDuration: "",
			StartTime:     "",
			HandoffTime:   "16:00:00",
			HandoffDay:    "monday",
		}
		assertDeepEqual(t, first, expected)
	})
}
