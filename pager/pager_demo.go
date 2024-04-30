//go:build demo

package pager

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/firehydrant/signals-migrator/console"
	"github.com/gosimple/slug"
)

func NewPager(ctx context.Context, kind string, apiKey string, appId string) (Pager, error) {
	switch strings.ToLower(kind) {
	case "pagerduty":
		rs := recorderServer(ctx, "PagerDuty")
		return NewPagerDutyWithURL(apiKey, rs.URL), nil
	case "opsgenie":
		rs := recorderServer(ctx, "Opsgenie")
		return NewOpsgenieWithURL(apiKey, rs.URL), nil
	}

	return nil, fmt.Errorf("%w '%s'", ErrUnknownProvider, kind)
}

//go:embed testdata/*/apiserver
var testdata embed.FS

// recorderServer is a hack mostly taken from pager_test.go to serve recorded responses.
// Only use for demo purposes, see build tags.
func recorderServer(ctx context.Context, provider string) *httptest.Server {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		baseDir := fmt.Sprintf("Test%s", provider)
		urlPath := r.URL.Path
		if r.URL.RawQuery != "" {
			urlPath += "?" + r.URL.RawQuery
		}
		filename, err := url.JoinPath("testdata", baseDir, "apiserver", slug.Make(urlPath)+".json")
		if err != nil {
			console.Warnf("[httptest] error joining path for expected response: %s\n", err)
		}
		data, err := testdata.ReadFile(filename)
		if err != nil {
			console.Warnf("[httptest] error reading file '%s': %s\n", filename, err.Error())
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})
	s := httptest.NewServer(h)
	go func() {
		<-ctx.Done()
		s.Close()
	}()
	return s
}
