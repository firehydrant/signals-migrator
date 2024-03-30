package console

import (
	"fmt"

	"github.com/charmbracelet/huh/spinner"
)

func Spin(action func(), title string, args ...any) {
	spinner.New().
		Title(fmt.Sprintf(title, args...)).
		Action(action).Run()
}
