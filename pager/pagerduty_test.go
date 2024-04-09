package pager_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/firehydrant/signals-migrator/pager"
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

func TestPagerDuty(t *testing.T) {
	ts := pagerDutyHttpServer(t)
	ctx, pd := pagerDutyTestClient(t, ts.URL)

	t.Run("ListUsers", func(t *testing.T) {
		u, err := pd.ListUsers(ctx)
		if err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}
		t.Logf("found %d users", len(u))
		// The only assertion so far from this test is that method will not error.
		// We need stronger checks for this test to be more useful.
		// TODO: verify that the users are accurate.
	})

	t.Run("LoadSchedules", func(t *testing.T) {
		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}
		// The only assertion so far from this test is that method will not error.
		// We need stronger checks for this test to be more useful.
		// TODO: verify that the schedules were saved to the database.
	})
}
