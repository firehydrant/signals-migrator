package cmd

import (
	"fmt"
	"log/slog"

	"github.com/firehydrant/signals-migrator/fetcher"
	"github.com/firehydrant/signals-migrator/renderer"
	"github.com/firehydrant/signals-migrator/types"
	"github.com/urfave/cli/v2"
)

var importFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "pagerduty-api-key",
		Usage:    "PagerDuty API key",
		EnvVars:  []string{"PAGERDUTY_API_KEY"},
		Required: true,
	},
	&cli.StringFlag{
		Name:  "provider",
		Usage: "The alerting provider to generate from",
		Value: "pagerduty",
	},
}

var ImportCommand = &cli.Command{
	Name:   "import",
	Usage:  "Imports Signals resources from a legacy alerting provider",
	Action: importAction,
	Flags:  ConcatFlags([][]cli.Flag{importFlags, flags}),
}

func importAction(ctx *cli.Context) error {
	var provider fetcher.Fetcher

	switch ctx.String("provider") {
	case "pagerduty":
		provider = fetcher.NewPagerDuty(ctx.String("pagerduty-api-key"))
	}

	fh, err := fetcher.NewFireHydrant(ctx.String("firehydrant-api-key"), ctx.String("firehydrant-api-endpoint"))
	if err != nil {
		return err
	}

	org := types.NewOrganization()

	users, err := provider.Users(ctx.Context)
	if err != nil {
		return fmt.Errorf("unable to fetch users from provider: %w", err)
	}
	org.Users = users

	for i, user := range org.Users {
		_, err := fh.User(ctx.Context, user.Email)

		if err != nil {
			slog.Warn("unable to find user", user.Email, err)
			org.Users[i], err = fh.ReconcileUser(ctx.Context, user)

			if err != nil {
				slog.Error("unable to reconcile user", user.Email, err)
			}

			continue
		}
	}

	teams, err := provider.Teams(ctx.Context)
	if err != nil {
		return fmt.Errorf("unable to fetch teams from provider: %w", err)
	}
	org.Teams = teams

	for i, team := range org.Teams {
		_, err := fh.Team(ctx.Context, team.Name)

		if err != nil {
			slog.Warn("unable to find team", team.Name, err)
			org.Teams[i], err = fh.ReconcileTeam(ctx.Context, team)
			if err != nil {
				slog.Error("unable to reconcile team", team.Name, err)
			}

			continue
		}
	}

	err = writeProviders()
	if err != nil {
		return fmt.Errorf("could not write providers: %w", err)
	}

	fhr := renderer.NewFireHydrant(org)
	hcl, err := fhr.Hcl()

	if err != nil {
		return fmt.Errorf("unable to render HCL: %w", err)
	}

	return writeHclToFile(hcl, "firehydrant.tf")
}
