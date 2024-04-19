package tfrender_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/firehydrant/signals-migrator/store"
	"gotest.tools/v3/golden"
)

func TestRenderPagerDuty(t *testing.T) {
	assert := func(t *testing.T) {
		seedFile := fmt.Sprintf("%s_seed.sql", t.Name())
		seed, err := os.ReadFile(filepath.Join("testdata", seedFile))
		if err != nil {
			t.Fatal(err)
		}
		sql := strings.TrimSpace(string(seed))

		ctx, tfr := tfrInit(t)
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

		golden.Assert(t, string(content), filepath.Join(filepath.Dir(t.Name()), goldenFile(tfr.Filename())))
	}

	t.Run("TeamWithSchedules", assert)
	t.Run("EscalationPolicy", assert)
}
