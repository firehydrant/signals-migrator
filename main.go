package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/fatih/color"
	"github.com/firehydrant/signals-migrator/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "signals-migrator"
	app.Usage = "Easily migrate from legacy alerting providers to Signals by FireHydrant"
	app.Version = revision()

	app.Commands = []*cli.Command{
		cmd.ImportCommand,
		{
			Name:  "version",
			Usage: "Print the version",
			Action: func(c *cli.Context) error {
				fmt.Printf("signals-migrator version %s\n" + revision()) //nolint:forbidigo,govet
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		color.HiRed("Error: %v\n", err.Error())
		os.Exit(1)
	}
}

var version = "dev"

func revision() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return version
	}
	if v := bi.Main.Version; v != "(devel)" {
		// Installed via `go install`.
		version = v
	}

	rev := ""
	for _, i := range bi.Settings {
		if i.Key == "vcs.revision" {
			rev = i.Value
			break
		}
	}
	if rev != "" {
		return fmt.Sprintf("%s-%s", version, rev)
	}
	return version
}
