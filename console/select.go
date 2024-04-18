package console

import (
	"fmt"
	"slices"

	"github.com/charmbracelet/huh"
)

func Selectf[T any](options []T, toString func(T) string, title string, args ...any) (int, T, error) {
	opts := make([]huh.Option[int], len(options))
	for i, option := range options {
		opts[i] = huh.NewOption(toString(option), i)
	}
	var value int

	s := huh.NewSelect[int]().
		Title(fmt.Sprintf(title, args...)).
		Description(fmt.Sprintf(title, args...)).
		Options(opts...).
		Value(&value).
		WithHeight(15)

	if err := huh.NewForm(huh.NewGroup(s)).Run(); err != nil {
		return -1, options[0], fmt.Errorf("selecting options: %w", err)
	}

	return value, options[value], nil
}

func MultiSelectf[T any](options []T, toString func(T) string, title string, args ...any) ([]int, []T, error) {
	opts := make([]huh.Option[int], len(options))
	for i, option := range options {
		opts[i] = huh.NewOption(toString(option), i)
	}
	var values []int

	s := huh.NewMultiSelect[int]().
		Title(fmt.Sprintf(title, args...)).
		Description("Select with <Space>, confirm with <Enter>").
		Options(opts...).
		Value(&values)

	for {
		if err := huh.NewForm(huh.NewGroup(s)).Run(); err != nil {
			return nil, nil, fmt.Errorf("selecting options: %w", err)
		}

		values = slices.Clip(values)
		if len(values) == 0 {
			Warnf("You have not selected any options.")
			continue
		}
		Warnf("You have selected: \n")
		for _, i := range values {
			Warnf("  %s\n", opts[i].Key)
		}

		response, _, err := Selectf([]string{"Yes", "No"}, func(s string) string { return s }, "Confirm selection?")
		if err != nil {
			return nil, nil, fmt.Errorf("confirming selection: %w", err)
		}
		if response == 0 {
			break
		}
	}

	selected := make([]T, len(values))
	for i, value := range values {
		selected[i] = options[value]
	}

	return values, selected, nil
}
