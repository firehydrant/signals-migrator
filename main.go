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
				fmt.Println("version: " + revision()) //nolint:forbidigo
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		color.HiRed("Error: %v\n", err.Error())
		os.Exit(1)
	}
}

func revision() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		panic("cannot retrieve buildinfo. is go go-ing?")
	}

	for _, i := range bi.Settings {
		if i.Key == "vcs.revision" {
			return i.Value
		}
	}
	return "dev"
}
