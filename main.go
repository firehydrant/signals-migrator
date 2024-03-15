package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/firehydrant/signals-migrator/cmd"
	"github.com/urfave/cli/v2"
)

var Revision string

func main() {
	app := cli.NewApp()
	app.Name = "signals-migrator"
	app.Usage = "Easily migrate from legacy alerting providers to Signals by FireHydrant"
	app.Version = Revision

	app.Commands = []*cli.Command{
		cmd.ImportCommand,
		cmd.DatadogCommand,
		{
			Name:  "version",
			Usage: "Print the version",
			Action: func(c *cli.Context) error {
				fmt.Println("version: " + Revision) //nolint:forbidigo
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("could not run application", err)
	}
}
