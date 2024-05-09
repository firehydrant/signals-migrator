package firehydrant_test

import (
	"context"
	"testing"

	"github.com/firehydrant/signals-migrator/internal/firehydrant"
	"github.com/firehydrant/signals-migrator/internal/testkit"
)

func TestFireHydrantClient(t *testing.T) {
	ctx := testkit.NewStore(t, context.Background())
	ts := testkit.NewHTTPServer(t)

	client, err := firehydrant.NewClient("testing-only", ts.URL)
	if err != nil {
		t.Fatalf("error creating FireHydrant client: %s", err)
	}

	t.Run("ListTeams", func(t *testing.T) {
		teams, err := client.ListTeams(ctx)
		if err != nil {
			t.Fatalf("error listing teams: %s", err)
		}
		testkit.GoldenJSON(t, teams)
	})
}
