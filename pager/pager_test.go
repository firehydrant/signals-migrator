package pager_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/firehydrant/signals-migrator/store"
	"github.com/gosimple/slug"
	"gotest.tools/v3/golden"
)

func withTestDB(t *testing.T) context.Context {
	t.Helper()
	ctx := store.WithContext(context.Background())
	t.Cleanup(func() {
		if err := store.FromContext(ctx).Close(); err != nil {
			t.Fatalf("error closing db: %s", err)
		}
	})
	return ctx
}

func pagerProviderHttpServer(t *testing.T) *httptest.Server {
	t.Helper()
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		baseTestDir := t.Name()
		// Get the root-most test name as directory.
		// This should be "Test[ProviderName]", e.g. TestPagerDuty.
		for b := filepath.Dir(baseTestDir); b != "."; b = filepath.Dir(b) {
			baseTestDir = b
		}
		filename, err := url.JoinPath("testdata", baseTestDir, "apiserver", slug.Make(r.URL.Path)+".json")
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
	s := httptest.NewServer(h)
	t.Cleanup(s.Close)
	return s
}

// WARNING: `got` must be refer to struct or slice. Go's builtin map is not order-deterministic, thus it might produce inconsistent JSON.
func assertJSON(t *testing.T, got any) {
	t.Helper()

	b, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatalf("error marshalling to json: %s", err)
	}

	// Ensure the file ends with a newline, so when editors aggresively
	// auto-fix this, golden tests don't actually fail.
	if b[len(b)-1] != '\n' {
		b = append(b, '\n')
	}

	goldenFile := t.Name() + ".golden.json"
	golden.Assert(t, string(b), goldenFile)
}
