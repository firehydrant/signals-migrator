package console

import (
	"fmt"

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
		Options(opts...).
		Value(&value).
		WithHeight(15)

	if err := huh.NewForm(huh.NewGroup(s)).Run(); err != nil {
		return -1, options[0], fmt.Errorf("selecting options: %w", err)
	}

	return value, options[value], nil
}
