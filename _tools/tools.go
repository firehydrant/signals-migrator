//go:build tools

package tools

import (
	// Build
	_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
	_ "golang.org/x/tools/cmd/stringer"

	// Test / lint
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "gotest.tools/gotestsum"
)
