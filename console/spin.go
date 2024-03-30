package console

import (
	"fmt"

	"github.com/charmbracelet/huh/spinner"
)

func Spin(action func(), title string, args ...any) {
	if err := spinner.New().
		Title(fmt.Sprintf(title, args...)).
		Action(action).Run(); err != nil {
		panic(err)
	}
}
