package tfrender_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/firehydrant/signals-migrator/store"
	"github.com/firehydrant/signals-migrator/tfrender"
	"gotest.tools/golden"
)

func tfrInit(t *testing.T) (context.Context, *tfrender.TFRender) {
	t.Helper()

	ctx := store.WithContext(context.Background())
	t.Cleanup(func() { store.FromContext(ctx).Close() })

	ctx = store.WithTx(ctx)
	t.Cleanup(func() { store.RollbackTx(ctx) })

	tfr, err := tfrender.New(t.TempDir(), t.Name()+".tf")
	if err != nil {
		t.Fatal(err)
	}
	return ctx, tfr
}

func goldenFile(name string) string {
	base := filepath.Base(name)
	ext := filepath.Ext(name)
	return fmt.Sprintf("%s.golden%s", base[:len(base)-len(ext)], ext)
}

func createUsers(t *testing.T, ctx context.Context, variant string) {
	t.Helper()
	id := fmt.Sprintf("id-for-user-%s", variant)
	extID := fmt.Sprintf("id-for-ext-user-%s", variant)
	email := fmt.Sprintf("user-%s@example.com", variant)
	name := fmt.Sprintf("User %s", variant)

	if err := store.UseQueries(ctx).InsertFhUser(ctx, store.InsertFhUserParams{
		ID:    id,
		Email: email,
		Name:  name,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.UseQueries(ctx).InsertExtUser(ctx, store.InsertExtUserParams{
		ID:       extID,
		Email:    email,
		Name:     name,
		FhUserID: sql.NullString{String: id, Valid: true},
	}); err != nil {
		t.Fatal(err)
	}
}

func createTeams(t *testing.T, ctx context.Context, variant string, withFhTeam bool) {
	t.Helper()
	id := fmt.Sprintf("id-for-team-%s", variant)
	slug := fmt.Sprintf("team-%s-slug", variant)
	extID := fmt.Sprintf("id-for-ext-team-%s", variant)
	name := fmt.Sprintf("Team %s", variant)

	if err := store.UseQueries(ctx).InsertFhTeam(ctx, store.InsertFhTeamParams{
		ID:   id,
		Name: name,
		Slug: slug,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.UseQueries(ctx).InsertExtTeam(ctx, store.InsertExtTeamParams{
		ID:       extID,
		Name:     name,
		Slug:     slug,
		FhTeamID: sql.NullString{String: id, Valid: withFhTeam},
	}); err != nil {
		t.Fatal(err)
	}
}

func TestRenderDataUser(t *testing.T) {
	ctx, tfr := tfrInit(t)
	for i := range 3 {
		createUsers(t, ctx, strconv.Itoa(i))
	}

	if err := tfr.Write(ctx); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(tfr.Filepath())
	if err != nil {
		t.Fatal(err)
	}

	// Expect output to have 3 user blocks.
	golden.Assert(t, string(content), goldenFile(tfr.Filename()))
}

func TestRenderTeamResource(t *testing.T) {
	ctx, tfr := tfrInit(t)
	for i := range 4 {
		createUsers(t, ctx, strconv.Itoa(i))
	}
	for i := range 4 {
		// team0 and team2 refers to existing FireHydrant teams,
		// so they will have import {} block associated to them.
		createTeams(t, ctx, strconv.Itoa(i), i%2 == 0)
	}

	if err := store.UseQueries(ctx).InsertExtMembership(ctx, store.InsertExtMembershipParams{
		UserID: "id-for-ext-user-0",
		TeamID: "id-for-ext-team-0",
	}); err != nil {
		t.Fatal(err)
	}

	if err := tfr.Write(ctx); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(tfr.Filepath())
	if err != nil {
		t.Fatal(err)
	}

	// Expect output to have:
	// - 4 user blocks
	// - 4 team blocks, where:
	//   - team 0 and team 2 have import {} block
	//   - team 0 has 1 user membership, linked with user 0
	golden.Assert(t, string(content), goldenFile(tfr.Filename()))
}
