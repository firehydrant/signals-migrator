package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/firehydrant/signals-migrator/console"
	"github.com/firehydrant/signals-migrator/pager"
	"github.com/firehydrant/signals-migrator/store"
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
		Name:     "provider-app-id",
		Usage:    "Provider APP ID",
		EnvVars:  []string{"PROVIDER_APP_ID"},
		Required: false,
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

func importAction(cliCtx *cli.Context) error {
	providerName := cliCtx.String("provider")
	provider, err := pager.NewPager(providerName, cliCtx.String("provider-api-key"), cliCtx.String("provider-app-id"))
	if err != nil {
		return fmt.Errorf("initializing pager provider: %w", err)
	}
	fh, err := pager.NewFireHydrant(cliCtx.String("firehydrant-api-key"), cliCtx.String("firehydrant-api-endpoint"))
	if err != nil {
		return fmt.Errorf("initializing FireHydrant client: %w", err)
	}

	ctx := store.WithContext(cliCtx.Context)
	defer store.FromContext(ctx).Close()

	if err := importUsers(ctx, provider, fh); err != nil {
		return fmt.Errorf("importing users: %w", err)
	}
	console.Infof("Imported users from %s.\n", providerName)

	if err := importTeams(ctx, provider, fh); err != nil {
		return fmt.Errorf("importing teams: %w", err)
	}
	console.Infof("Imported teams from %s.\n", providerName)

	if err := provider.LoadSchedules(ctx); err != nil {
		return fmt.Errorf("importing schedules: %w", err)
	}
	console.Infof("Imported schedules from %s.\n", providerName)

	if err := importEscalationPolicies(ctx, provider, fh); err != nil {
		return fmt.Errorf("importing escalation policies: %w", err)
	}
	console.Infof("Imported escalation policies from %s.\n", providerName)

	tfr, err := tfrender.New(
		cliCtx.String("output-dir"),
		fmt.Sprintf("%s_to_fh_signals.tf", strings.ToLower(providerName)),
	)
	if err != nil {
		return fmt.Errorf("initializing Terraform render space: %w", err)
	}
	return tfr.Write(ctx)
}

// Because the amount of escalation policies being queried can be large for some organizations,
// we preemptively save everything from API response to database. Then, we prompt user to select
// which rows to import. We mark the selected rows from users in `to_import` field and delete
// the ones that we will not import to FireHydrant. This is done to simplify the state management
// between queries and filtering.
func importEscalationPolicies(ctx context.Context, provider pager.Pager, fh *pager.FireHydrant) error {
	if err := provider.LoadEscalationPolicies(ctx); err != nil {
		return fmt.Errorf("unable to load escalation policies: %w", err)
	}
	allEps, err := store.UseQueries(ctx).ListExtEscalationPolicies(ctx)
	if err != nil {
		return fmt.Errorf("unable to list escalation policies: %w", err)
	}
	options := []store.ExtEscalationPolicy{{ID: "[+] ADD ALL"}, {ID: "[<] SKIP ALL"}}
	options = append(options, allEps...)
	console.Warnf("Please select (out of %d) which escalation policies to migrate.\n", len(allEps))
	selected, toImport, err := console.MultiSelectf(options, func(ep store.ExtEscalationPolicy) string {
		return fmt.Sprintf("%s %s", ep.ID, ep.Name)
	}, "Which escalation policies should be migrated to FireHydrant?")
	if err != nil {
		return fmt.Errorf("selecting escalation policies: %w", err)
	}
	if len(selected) == 0 {
		console.Warnf("No escalation policies selected for migration. Assuming all...\n")
		selected = append(selected, 0)
	}

	switch selected[0] {
	case 0:
		console.Successf("[+] All escalation policies will be migrated to FireHydrant.\n")
		if err := store.UseQueries(ctx).MarkAllExtEscalationPolicyToImport(ctx); err != nil {
			return fmt.Errorf("unable to mark all escalation policies for import: %w", err)
		}
	case 1:
		console.Warnf("[<] No escalation policies will be migrated to FireHydrant.\n")
	default:
		for _, ep := range toImport {
			if ep.ID == "[+] ADD ALL" || ep.ID == "[<] SKIP ALL" {
				continue
			}
			if err := store.UseQueries(ctx).MarkExtEscalationPolicyToImport(ctx, ep.ID); err != nil {
				return fmt.Errorf("unable to mark escalation policy '%s' for import: %w", ep.Name, err)
			}
		}
	}

	if err := store.UseQueries(ctx).DeleteExtEscalationPolicyUnimported(ctx); err != nil {
		return fmt.Errorf("unable to delete unimported escalation policies: %w", err)
	}
	return nil
}

func importTeams(ctx context.Context, provider pager.Pager, fh *pager.FireHydrant) error {
	// Get all of the teams registered from Pager Provider (e.g. PagerDuty)
	var err error
	var providerTeams []*pager.Team
	console.Spin(func() {
		providerTeams, err = provider.ListTeams(ctx)
	}, "Fetching all teams from provider...")
	if err != nil {
		return fmt.Errorf("unable to fetch teamsfrom provider: %w", err)
	}
	console.Successf("Found %d teams from provider.\n", len(providerTeams))

	// List out all of the teams from FireHydrant.
	var fhTeams []*pager.Team
	console.Spin(func() {
		fhTeams, err = fh.ListTeams(ctx)
	}, "Fetching all teams from FireHydrant...")
	if err != nil {
		return fmt.Errorf("unable to match users to FireHydrant: %w", err)
	}
	console.Successf("Found %d teams on FireHydrant.\n", len(fhTeams))

	// Now, for every team we found in Pager provider, we prompt console for one of three choices:
	// 1. Create a new team in FireHydrant
	// 2. Match with an existing team in FireHydrant
	// 3. Skip / ignore the team entirely
	options := append([]*pager.Team{
		&pager.Team{Slug: "[<] Skip"},
		&pager.Team{Slug: "[+] Create"},
	}, fhTeams...)
	for _, extTeam := range providerTeams {
		i, t, err := console.Selectf(options, func(u *pager.Team) string {
			return u.String()
		}, "For the team '%s' from provider:", extTeam.String())
		if err != nil {
			return fmt.Errorf("selecting match for '%s': %w", extTeam.String(), err)
		}
		switch i {
		case 0:
			console.Warnf("[< SKIPPED] '%s' will not be imported to FireHydrant.\n", extTeam.String())
			continue
		case 1:
			console.Successf("[+  CREATE] '%s' will be created in FireHydrant.\n", extTeam.String())
			if err := store.UseQueries(ctx).InsertExtTeam(ctx, store.InsertExtTeamParams{
				ID:       extTeam.ID,
				Name:     extTeam.Name,
				Slug:     extTeam.Slug,
				FhTeamID: sql.NullString{Valid: false},
			}); err != nil {
				return fmt.Errorf("unable to insert team '%s' into database: %w", extTeam.String(), err)
			}
			continue
		default:
			if err := store.UseQueries(ctx).InsertExtTeam(ctx, store.InsertExtTeamParams{
				ID:       extTeam.ID,
				Name:     extTeam.Name,
				Slug:     extTeam.Slug,
				FhTeamID: sql.NullString{Valid: true, String: t.ID},
			}); err != nil {
				return fmt.Errorf("unable to insert team '%s' into database: %w", extTeam.String(), err)
			} else {
				console.Infof("[= MATCHED]\n  '%s'\n  => '%s'.\n", extTeam.String(), t.String())
			}
		}
	}

	allTeams, err := store.UseQueries(ctx).ListTeams(ctx)
	if err != nil {
		return fmt.Errorf("unable to list all teams: %w", err)
	}
	for _, extTeam := range allTeams {
		t := &pager.Team{
			Resource: pager.Resource{
				ID:   extTeam.ID,
				Name: extTeam.Name,
			},
			Slug: extTeam.Slug,
		}
		if err := provider.PopulateTeamMembers(ctx, t); err != nil {
			return fmt.Errorf("unable to populate team members for '%s': %w", extTeam.Name, err)
		}

		for _, member := range t.Members {
			if err := store.UseQueries(ctx).InsertExtMembership(ctx, store.InsertExtMembershipParams{
				TeamID: extTeam.ID,
				UserID: member.ID,
			}); err != nil {
				return fmt.Errorf("unable to insert team member '%s' into database: %w", member.String(), err)
			}
		}
	}

	return nil
}

func importUsers(ctx context.Context, provider pager.Pager, fh *pager.FireHydrant) error {
	// Get all of the users registered from Pager Provider (e.g. PagerDuty)
	var err error
	var providerUsers []*pager.User
	console.Spin(func() {
		providerUsers, err = provider.ListUsers(ctx)
	}, "Fetching all users from provider...")
	if err != nil {
		return fmt.Errorf("unable to fetch users from provider: %w", err)
	}
	console.Successf("Found %d users from provider.\n", len(providerUsers))
	for _, user := range providerUsers {
		if err := store.UseQueries(ctx).InsertExtUser(ctx, store.InsertExtUserParams{
			ID:       user.ID,
			Name:     user.Name,
			Email:    user.Email,
			FhUserID: sql.NullString{Valid: false},
		}); err != nil {
			return fmt.Errorf("unable to insert user '%s' into database: %w", user.Email, err)
		}
	}

	// Find out which users do not already have a FireHydrant account
	var unmatchedUsers []*pager.User
	console.Spin(func() {
		unmatchedUsers, err = fh.MatchUsers(ctx, providerUsers)
		if err != nil {
			return
		}
	}, "Matching users with existing FireHydrant users...")
	if err != nil {
		return fmt.Errorf("unable to match users to FireHydrant: %w", err)
	}

	// Prompt console to match users manually if necessary.
	if len(unmatchedUsers) > 0 {
		console.Warnf("Found %d users which require manual mapping to FireHydrant.\n", len(unmatchedUsers))
		options, err := fh.ListUsers(ctx)
		if err != nil {
			return fmt.Errorf("unable to list users from FireHydrant: %w", err)
		}
		console.Warnf("Please select from %d FireHydrant users to match.\n", len(options))

		// Prepend options with a choice to skip
		options = append([]*pager.User{&pager.User{Email: "[<] Skip"}}, options...)
		for _, extUser := range unmatchedUsers {
			i, fhUser, err := console.Selectf(options, func(u *pager.User) string {
				return u.String()
			}, "Select a FireHydrant user to match with '%s'", extUser.String())
			if err != nil {
				return fmt.Errorf("selecting match for '%s': %w", extUser.String(), err)
			}
			if i == 0 {
				console.Warnf("[< SKIPPED] '%s' will not be imported to FireHydrant.\n", extUser.String())
				continue
			}
			if err := fh.PairUsers(ctx, fhUser.ID, extUser.ID); err != nil {
				return fmt.Errorf("pairing '%s' with '%s': %w", extUser.String(), fhUser.String(), err)
			} else {
				console.Successf("[= MATCHED]\n  '%s'\n  => '%s'.\n", extUser.String(), fhUser.String())
			}
		}
	}
	return nil
}
