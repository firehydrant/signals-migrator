package pager_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/firehydrant/signals-migrator/pager"
	"github.com/firehydrant/signals-migrator/store"
	"github.com/gosimple/slug"
)

func opsgenieHttpServer(t *testing.T) *httptest.Server {
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

func opsgenieTestClient(t *testing.T, apiURL string) (context.Context, *pager.Opsgenie) {
	ctx := withTestDB(t)
	og := pager.NewOpsgenieWithURL("api-key-very-secret", apiURL)
	return ctx, og
}

func TestOpsgenie(t *testing.T) {
	ts := opsgenieHttpServer(t)
	ctx, og := opsgenieTestClient(t, strings.TrimPrefix(ts.URL, "http://"))

	t.Run("ListUsers", func(t *testing.T) {
		u, err := og.ListUsers(ctx)
		if err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		t.Logf("found %d users", len(u))

		// Verify that the first user is as expected.
		mika := u[0]
		expected := &pager.User{
			Email: "john.doe@opsgenie.com",
			Resource: pager.Resource{
				ID:   "b5b92115-bfe7-43eb-8c2a-e467f2e5ddc4",
				Name: "john doe",
			},
		}
		assertDeepEqual(t, mika, expected)
	})

	t.Run("LoadSchedules", func(t *testing.T) {
		if err := og.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		s, err := store.UseQueries(ctx).ListExtSchedules(ctx)
		if err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}
		t.Logf("found %d schedules", len(s))

		// Verify that the first schedule is as expected.
		first := s[0]
		expected := store.ExtSchedule{
			ID:            "3fee43f2-02da-49be-ab50-c88ed13aecc3-b1b5f7f6-728b-47bd-af02-c3e1c33bf219",
			Name:          "Customer Success_schedule - Rot1",
			Description:   "(Rot1)",
			Timezone:      "America/Los_Angeles",
			Strategy:      "weekly",
			ShiftDuration: "",
			StartTime:     "",
			HandoffTime:   "04:45:32",
			HandoffDay:    "tuesday",
		}
		assertDeepEqual(t, first, expected)
	})

}
