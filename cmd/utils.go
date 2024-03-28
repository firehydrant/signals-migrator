package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/firehydrant/signals-migrator/renderer"
)

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

func writeProviders() error {
	t, err := renderer.NewTerraform()
	if err != nil {
		return fmt.Errorf("initializing terraform renderer: %w", err)
	}
	t.Provider("datadog", "datadog/datadog", "~> 3.37.0")
	t.Provider("firehydrant", "firehydrant/firehydrant", "~> 0.5.0")

	hcl, err := t.Hcl()
	if err != nil {
		return err
	}

	return writeHclToFile(hcl, "providers.tf")
}

func writeHclToFile(hcl string, file string) error {
	outPath := filepath.Join(".", "output")
	err := os.MkdirAll(outPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create output path: %w", err)
	}

	err = os.WriteFile(filepath.Join(outPath, file), []byte(hcl), 0644)
	if err != nil {
		return fmt.Errorf("unable to write HCL to file: %w", err)
	}

	return nil
}
