package testkit

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/gosimple/slug"
)

func NewHTTPServer(t *testing.T) *httptest.Server {
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
			dir := filepath.Dir(filename)
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("error creating directory: %s", err)
			}
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
