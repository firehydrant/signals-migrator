package tfrender_test

import (
	"os"
	"strings"
	"testing"

	"github.com/firehydrant/signals-migrator/store"
	"gotest.tools/golden"
)

func TestRenderOpsgenieSample(t *testing.T) {
	ctx, tfr := tfrInit(t)

	seed, err := os.ReadFile("testdata/opsgenie_seed.sql")
	if err != nil {
		t.Fatal(err)
	}

	sql := strings.TrimSpace(string(seed))

	if _, err := store.FromContext(ctx).ExecContext(ctx, sql); err != nil {
		t.Fatal(err)
	}
	if err := tfr.Write(ctx); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(tfr.Filepath())
	if err != nil {
		t.Fatal(err)
	}

	golden.Assert(t, string(content), goldenFile(tfr.Filename()))
}
