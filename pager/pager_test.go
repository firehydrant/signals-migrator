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

	f := filepath.Join(t.TempDir(), slug.Make(t.Name())+".db")
	_ = os.Remove(f)

	ctx := store.WithContextAndDSN(context.Background(), "file:"+f)
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
		urlPath := r.URL.Path
		if r.URL.RawQuery != "" {
			urlPath += "?" + r.URL.RawQuery
		}
		filename, err := url.JoinPath("testdata", baseTestDir, "apiserver", slug.Make(urlPath)+".json")
		if err != nil {
			t.Fatalf("error joining path for expected response: %s", err)
		}
		if _, err := os.Stat(filename); err != nil {
			t.Logf("file not found: %s, pre-creating file with empty JSON", filename)
			if f, err := os.OpenFile(filename, os.O_CREATE, 0644); err != nil {
				t.Fatalf("error creating file: %s", err)
			} else {
				_, _ = f.WriteString("{}")
				_ = f.Close()
			}
		}
		http.ServeFile(w, r, filename)
	})
	s := httptest.NewServer(h)
	t.Cleanup(s.Close)
	return s
}

// assertJSON compares the JSON representation of `got` with the golden file.
// The golden file is named conventionally after the full test name with `.golden.json` suffix.
//
// WARNING: `got` must not refer to data with non-deterministic order.
// For example, Go's builtin map is not order-deterministic, thus it might produce inconsistent JSON comparison in this method.
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
	t.Logf("using %s\n", goldenFile)
	golden.Assert(t, string(b), goldenFile)
}
