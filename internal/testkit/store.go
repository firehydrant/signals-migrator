package testkit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/firehydrant/signals-migrator/store"
	"github.com/gosimple/slug"
)

func NewStore(t *testing.T, ctx context.Context) context.Context {
	t.Helper()

	f := filepath.Join(t.TempDir(), slug.Make(t.Name())+".db")
	_ = os.Remove(f)

	ctx = store.WithContextAndDSN(ctx, "file:"+f)
	t.Cleanup(func() {
		if err := store.FromContext(ctx).Close(); err != nil {
			t.Fatalf("error closing db: %s", err)
		}
	})
	return ctx
}
