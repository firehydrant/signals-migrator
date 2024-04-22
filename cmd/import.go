package cmd

import (
	"context"
	"fmt"
	"path/filepath"
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
	provider, err := pager.NewPager(
		providerName,
		cliCtx.String("provider-api-key"),
		cliCtx.String("provider-app-id"),
	)
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

	tfr, err := tfrender.New(filepath.Join(
		cliCtx.String("output-dir"),
		fmt.Sprintf("%s_to_fh_signals.tf", strings.ToLower(providerName)),
	))
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
	// Some providers made their users adopt an alternate concept of teams.
	//
	// For example, PagerDuty has "Teams" and "Services". In vacuum, they intuitively refer to
	// "people" and "computers", respectively. However, their implementation for on call schedule
	// is tied to "Services". As such, many users of PagerDuty never really defined "Teams" in
	// their instance and use "Services" in practice for grouping on call "Teams".
	//
	// Now, all of that is imported as "Teams" in FireHydrant. As such, we prompt user to select
	// their logical representation of "Teams" when a provider has multiple options.
	// There may be a case where users may want to import both to FireHydrant. It is not currently
	// supported, but can be a reasonable future enhancement.
	if choices := provider.TeamInterfaces(); len(choices) > 1 {
		_, ti, err := console.Selectf(choices, func(s string) string {
			return fmt.Sprintf("%s %s", provider.Kind(), s)
		}, "Let's fill out your teams in FireHydrant. Which team interface would you like to use?")
		if err != nil {
			return fmt.Errorf("selecting team interface: %w", err)
		}
		if err := provider.UseTeamInterface(ti); err != nil {
			return fmt.Errorf("setting team interface: %w", err)
		}
	}

	// Get all of the teams registered from Pager Provider (e.g. PagerDuty)
	var err error
	console.Spin(func() {
		err = provider.LoadTeams(ctx)
	}, "Fetching all teams from provider...")
	if err != nil {
		return fmt.Errorf("unable to fetch teams from provider: %w", err)
	}
	console.Successf("Loaded all teams from provider.\n")

	// List out all of the teams from FireHydrant.
	var fhTeams []*pager.Team
	console.Spin(func() {
		fhTeams, err = fh.ListTeams(ctx)
	}, "Fetching all teams from FireHydrant...")
	if err != nil {
		return fmt.Errorf("unable to match users to FireHydrant: %w", err)
	}
	console.Successf("Found %d teams on FireHydrant.\n", len(fhTeams))

	// TODO: implement matching.
	// Now, for every team we found in Pager provider, we prompt console for one of three choices:
	// 1. Create a new team in FireHydrant
	// 2. Match with an existing team in FireHydrant
	// 3. Skip / ignore the team entirely

	if err := provider.LoadTeamMembers(ctx); err != nil {
		return fmt.Errorf("unable to populate team members: %w", err)
	}
	return nil
}

func importUsers(ctx context.Context, provider pager.Pager, fh *pager.FireHydrant) error {
	// Get all of the users registered from Pager Provider (e.g. PagerDuty)
	var err error
	console.Spin(func() {
		err = provider.LoadUsers(ctx)
	}, "Fetching all users from provider...")
	if err != nil {
		return fmt.Errorf("unable to fetch users from provider: %w", err)
	}
	console.Successf("Loaded all users from provider.\n")

	// Find out which users do not already have a FireHydrant account
	console.Spin(func() {
		err = fh.MatchUsers(ctx)
	}, "Matching users with existing FireHydrant users...")
	if err != nil {
		return fmt.Errorf("unable to match users to FireHydrant: %w", err)
	}
	return nil
}
