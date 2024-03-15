package cmd

import (
	"fmt"

	"github.com/firehydrant/signals-migrator/fetcher"
	"github.com/firehydrant/signals-migrator/renderer"
	"github.com/firehydrant/signals-migrator/types"
	"github.com/urfave/cli/v2"
)

var datadogFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "datadog-api-key",
		Usage:    "Datadog API key",
		EnvVars:  []string{"DATADOG_API_KEY"},
		Required: true,
	},
	&cli.StringFlag{
		Name:     "datadog-app-key",
		Usage:    "Datadog application key",
		EnvVars:  []string{"DATADOG_APP_KEY"},
		Required: true,
	},
}

var DatadogCommand = &cli.Command{
	Name:   "datadog",
	Usage:  "Creates Signals webhooks for each team in your FireHydrant account",
	Action: datadogAction,
	Flags:  ConcatFlags([][]cli.Flag{datadogFlags, flags}),
}

func datadogAction(ctx *cli.Context) error {
	var fh fetcher.FireHydrant

	fh, err := fetcher.NewFireHydrant(ctx.String("firehydrant-api-key"), ctx.String("firehydrant-api-endpoint"))
	if err != nil {
		return fmt.Errorf("could not initialize FireHydrant client: %w", err)
	}

	org := types.NewOrganization()

	teams, err := fh.Teams(ctx.Context)
	if err != nil {
		return fmt.Errorf("could not fetch teams: %w", err)
	}
	org.Teams = teams

	ddApiKey := ctx.String("datadog-api-key")
	ddAppKey := ctx.String("datadog-app-key")

	err = writeProviders()
	if err != nil {
		return fmt.Errorf("could not write providers: %w", err)
	}

	ddr := renderer.NewDatadog(ddApiKey, ddAppKey, org)
	hcl, err := ddr.Hcl()

	if err != nil {
		return fmt.Errorf("unable to render HCL: %w", err)
	}

	writeHclToFile(hcl, "datadog.tf")

	return nil
}
