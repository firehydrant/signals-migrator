package cmd

import (
	"context"
	"fmt"

	"github.com/firehydrant/signals-migrator/console"
	"github.com/firehydrant/signals-migrator/pager"
	"github.com/firehydrant/signals-migrator/tfrender"
	"github.com/urfave/cli/v2"
)

var importFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "provider-api-key",
		Usage:    "Provider API key",
		EnvVars:  []string{"PROVIDER_API_KEY"},
		Required: true,
	},
	&cli.StringFlag{
		Name:     "provider",
		Usage:    "The alerting provider to generate from",
		EnvVars:  []string{"PROVIDER"},
		Required: true,
	},
	&cli.StringFlag{
		Name:    "output-dir",
		Usage:   "The directory to write the Terraform configuration to",
		EnvVars: []string{"OUTPUT_DIR"},
		Value:   "./output",
	},
}

var ImportCommand = &cli.Command{
	Name:   "import",
	Usage:  "Imports Signals resources from a legacy alerting provider",
	Action: importAction,
	Flags:  ConcatFlags([][]cli.Flag{importFlags, flags}),
}

func importAction(ctx *cli.Context) error {
	tfr, err := tfrender.New(ctx.String("output-dir"))
	if err != nil {
		return fmt.Errorf("initializing Terraform render space: %w", err)
	}

	providerName := ctx.String("provider")
	provider, err := pager.NewPager(providerName, ctx.String("provider-api-key"))
	if err != nil {
		return fmt.Errorf("initializing pager provider: %w", err)
	}
	fh, err := pager.NewFireHydrant(ctx.String("firehydrant-api-key"), ctx.String("firehydrant-api-endpoint"))
	if err != nil {
		return fmt.Errorf("initializing FireHydrant client: %w", err)
	}

	users, err := importUsers(ctx.Context, tfr, provider, fh)
	if err != nil {
		return fmt.Errorf("importing users: %w", err)
	}
	console.Infof("Imported users from %s to FireHydrant.\n", providerName)

	teams, err := importTeams(ctx.Context, tfr, users, provider, fh)
	if err != nil {
		return fmt.Errorf("importing teams: %w", err)
	}
	console.Infof("Imported %d teams from %s to FireHydrant.\n", len(teams), providerName)

	return fmt.Errorf("not implemented")
}

func importTeams(ctx context.Context, tfr *tfrender.TFRender, users map[string]*pager.User, provider pager.Pager, fh *pager.FireHydrant) (map[string]*pager.Team, error) {
	// Get all of the teams registered from Pager Provider (e.g. PagerDuty)
	var err error
	var providerTeams []*pager.Team
	console.Spin(func() {
		providerTeams, err = provider.ListTeams(ctx)
	}, "Fetching all teams from provider...")
	if err != nil {
		return nil, fmt.Errorf("unable to fetch teamsfrom provider: %w", err)
	}
	console.Successf("Found %d teams from provider.\n", len(providerTeams))

	// List out all of the teams from FireHydrant.
	var fhTeams []*pager.Team
	console.Spin(func() {
		fhTeams, err = fh.ListTeams(ctx)
	}, "Fetching all teams from FireHydrant...")
	if err != nil {
		return nil, fmt.Errorf("unable to match users to FireHydrant: %w", err)
	}
	console.Successf("Found %d teams on FireHydrant.\n", len(fhTeams))

	matchedTeams := map[string]*pager.Team{}
	toCreateTeams := []*pager.Team{}
	// Now, for every team we found in Pager provider, we prompt console for one of three choices:
	// 1. Create a new team in FireHydrant
	// 2. Match with an existing team in FireHydrant
	// 3. Skip / ignore the team entirely
	options := append([]*pager.Team{
		&pager.Team{Slug: "[<] Skip"},
		&pager.Team{Slug: "[+] Create"},
	}, fhTeams...)
	for _, team := range providerTeams {
		i, t, err := console.Selectf(options, func(u *pager.Team) string {
			return u.String()
		}, "For the team '%s' from provider:", team.String())
		if err != nil {
			return nil, fmt.Errorf("selecting match for '%s': %w", team.String(), err)
		}
		switch i {
		case 0:
			console.Warnf("[SKIPPED] '%s' will not be imported to FireHydrant.\n", team.String())
			continue
		case 1:
			console.Successf("[ CREATE] '%s' will be created in FireHydrant.\n", team.String())
			toCreateTeams = append(toCreateTeams, team)
			continue
		default:
			console.Infof("[MATCHED] '%s' => '%s'.\n", team.String(), t.String())
			matchedTeams[team.Resource.ID] = t
		}
	}
	console.Successf("Found %d teams with existing match and %d teams to be created\n", len(matchedTeams), len(toCreateTeams))

	for _, team := range toCreateTeams {
		if err := provider.PopulateTeamMembers(ctx, team); err != nil {
			return nil, fmt.Errorf("unable to populate members of '%s': %w", team.String(), err)
		}
		// TODO: populate schedules
	}
	for _, team := range matchedTeams {
		if err := provider.PopulateTeamMembers(ctx, team); err != nil {
			return nil, fmt.Errorf("unable to populate members of '%s': %w", team.String(), err)
		}
		// TODO: populate schedules
	}

	// TODO: write to Terraform
	return nil, nil
}

func importUsers(ctx context.Context, tfr *tfrender.TFRender, provider pager.Pager, fh *pager.FireHydrant) (map[string]*pager.User, error) {
	// Get all of the users registered from Pager Provider (e.g. PagerDuty)
	var err error
	var providerUsers []*pager.User
	console.Spin(func() {
		providerUsers, err = provider.ListUsers(ctx)
	}, "Fetching all users from provider...")
	if err != nil {
		return nil, fmt.Errorf("unable to fetch users from provider: %w", err)
	}
	console.Successf("Found %d users from provider.\n", len(providerUsers))

	// Find out which users do not already have a FireHydrant account
	var users map[string]*pager.User
	var unmatchedUsers []*pager.User
	console.Spin(func() {
		users, unmatchedUsers, err = fh.MatchUsers(ctx, providerUsers)
		if err != nil {
			return
		}
	}, "Matching users with existing FireHydrant users...")
	if err != nil {
		return nil, fmt.Errorf("unable to match users to FireHydrant: %w", err)
	}

	// Prompt console to match users manually if necessary.
	if len(unmatchedUsers) > 0 {
		console.Warnf("Found %d users which require manual mapping to FireHydrant.\n", len(unmatchedUsers))
		options, err := fh.ListUsers(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list users from FireHydrant: %w", err)
		}
		console.Warnf("Please select from %d FireHydrant users to match.\n", len(options))

		// Prepend options with a choice to skip
		options = append([]*pager.User{&pager.User{Email: "[<] Skip"}}, options...)
		for _, user := range unmatchedUsers {
			i, u, err := console.Selectf(options, func(u *pager.User) string {
				return u.String()
			}, "Select a FireHydrant user to match with '%s'", user.String())
			if err != nil {
				return nil, fmt.Errorf("selecting match for '%s': %w", user.String(), err)
			}
			if i == 0 {
				console.Warnf("[SKIPPED] '%s' will not be imported to FireHydrant.\n", user.String())
				continue
			}
			console.Infof("[MATCHED] '%s' => '%s'.\n", user.String(), u.String())
			users[user.Email] = u
		}
	}

	// Render user information to TFRender.
	if err := tfr.DataFireHydrantUsers(users); err != nil {
		return nil, fmt.Errorf("writing users to Terraform: %w", err)
	}
	return users, nil
}
