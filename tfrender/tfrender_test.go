package tfrender_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/firehydrant/signals-migrator/store"
	"github.com/firehydrant/signals-migrator/tfrender"
	"gotest.tools/golden"
)

func tfrInit(t *testing.T) (context.Context, *tfrender.TFRender) {
	t.Helper()
	tfr, err := tfrender.New(t.TempDir(), t.Name()+".tf")
	if err != nil {
		t.Fatal(err)
	}
	return context.Background(), tfr
}

func createUsers(t *testing.T, ctx context.Context, count int) {
	t.Helper()
	for i := range count {
		if err := store.Query.InsertFhUser(ctx, store.InsertFhUserParams{
			ID:    fmt.Sprintf("id-for-user%d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
			Name:  fmt.Sprintf("User %d", i),
		}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRenderDataUser(t *testing.T) {
	ctx, tfr := tfrInit(t)
	createUsers(t, ctx, 3)

	if err := tfr.Write(ctx); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(tfr.Filepath())
	if err != nil {
		t.Fatal(err)
	}

	golden.Assert(t, string(content), tfr.Filename()+".golden")
}
