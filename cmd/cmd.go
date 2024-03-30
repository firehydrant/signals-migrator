package cmd

import (
	"github.com/urfave/cli/v2"
)

var flags = []cli.Flag{
	&cli.StringFlag{
		Name:     "firehydrant-api-key",
		Usage:    "FireHydrant API key - can be created at https://app.firehydrant.io/settings/api_keys",
		EnvVars:  []string{"FIREHYDRANT_API_KEY"},
		Required: true,
	},
	&cli.StringFlag{
		Name:    "firehydrant-api-endpoint",
		Usage:   "FireHydrant API endpoint",
		EnvVars: []string{"FIREHYDRANT_API_ENDPOINT"},
		Value:   "https://api.firehydrant.io/v1/",
	},
}

func ConcatFlags[T any](slices [][]T) []T {
	var totalLen int

	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]T, totalLen)

	var i int

	for _, s := range slices {
		i += copy(result[i:], s)
	}

	return result
}
