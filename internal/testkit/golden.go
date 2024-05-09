package testkit

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/golden"
)

func IsUpdate() bool {
	return golden.FlagUpdate()
}

func GoldenJSON(t *testing.T, received any) {
	t.Helper()

	b, err := json.MarshalIndent(received, "", "  ")
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
