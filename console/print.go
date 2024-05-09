package console

import "github.com/fatih/color"

var (
	Successf = color.New(color.FgHiGreen).Add(color.Bold).PrintfFunc()
	Infof    = color.New(color.FgHiBlue).PrintfFunc()
	Errorf   = color.New(color.FgHiRed).Add(color.Underline).PrintfFunc()
	Warnf    = color.New(color.FgHiYellow).Add(color.Bold).PrintfFunc()
)

func PadStrings[T any](elems []T, intFn func(T) int) int {
	max := 0
	for _, elem := range elems {
		if l := intFn(elem); l > max {
			max = l
		}
	}
	return max
}
