//go:build tools

package tools

import (
	// Build
	_ "golang.org/x/tools/cmd/stringer"

	// Test / lint
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "gotest.tools/gotestsum"
)
