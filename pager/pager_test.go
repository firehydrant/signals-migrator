package pager_test

import (
	"context"
	"testing"

	"github.com/firehydrant/signals-migrator/store"
)

func withTestDB(t *testing.T) context.Context {
	ctx := store.WithContext(context.Background())
	t.Cleanup(func() {
		if err := store.FromContext(ctx).Close(); err != nil {
			t.Fatalf("error closing db: %s", err)
		}
	})
	return ctx
}
