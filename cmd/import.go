package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/firehydrant/signals-migrator/console"
	"github.com/firehydrant/signals-migrator/internal/firehydrant"
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
	ctx, cancel := signal.NotifyContext(cliCtx.Context, os.Interrupt)
	defer cancel()

	providerName := cliCtx.String("provider")
	provider, err := pager.NewPager(
		ctx, providerName,
		cliCtx.String("provider-api-key"),
		cliCtx.String("provider-app-id"),
	)
	if err != nil {
		return fmt.Errorf("initializing pager provider: %w", err)
	}
	fh, err := firehydrant.NewClient(cliCtx.String("firehydrant-api-key"), cliCtx.String("firehydrant-api-endpoint"))
	if err != nil {
		return fmt.Errorf("initializing FireHydrant client: %w", err)
	}

	ctx = store.WithContext(ctx)
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
func importEscalationPolicies(ctx context.Context, provider pager.Pager, fh *firehydrant.Client) error {
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

func importTeams(ctx context.Context, provider pager.Pager, fh *firehydrant.Client) error {
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
	var fhTeams []store.FhTeam
	console.Spin(func() {
		fhTeams, err = fh.ListTeams(ctx)
	}, "Fetching all teams from FireHydrant...")
	if err != nil {
		return fmt.Errorf("unable to match users to FireHydrant: %w", err)
	}
	console.Successf("Found %d teams on FireHydrant.\n", len(fhTeams))

	if err := provider.LoadTeamMembers(ctx); err != nil {
		console.Errorf("unable to load team members: %s", err.Error())
		if y, yErr := console.YesNo("Continue without team members?"); yErr != nil || !y {
			return fmt.Errorf("unable to populate team members: %w", err)
		}
	}

	// First, we prompt users which teams to import to FireHydrant from the external provider.
	// We will mark the selected teams to import, then ask for user to match existing teams in FireHydrant (or create new).
	teams, err := provider.Teams(ctx)
	if err != nil {
		return fmt.Errorf("unable to list teams: %w", err)
	}
	console.Warnf("Please select which teams to migrate to FireHydrant.\n")
	_, toImport, err := console.MultiSelectf(teams, func(t store.ExtTeam) string {
		return fmt.Sprintf("%s %s", t.ID, t.Name)
	}, "Which teams should be migrated to FireHydrant?")
	if err != nil {
		return fmt.Errorf("selecting teams: %w", err)
	}
	for _, t := range toImport {
		if err := store.UseQueries(ctx).MarkExtTeamToImport(ctx, t.ID); err != nil {
			return fmt.Errorf("unable to mark team '%s' for import: %w", t.Name, err)
		}
	}

	// Now, we prompt users to match the teams that we are importing to FireHydrant.
	options := []store.FhTeam{{ID: "[+] CREATE NEW"}}
	options = append(options, fhTeams...)
	for _, t := range toImport {
		selected, fhTeam, err := console.Selectf(options, func(t store.FhTeam) string {
			return fmt.Sprintf("%s %s", t.ID, t.Name)
		}, fmt.Sprintf("Which FireHydrant team should '%s' be imported to?", t.Name))
		if err != nil {
			return fmt.Errorf("selecting FireHydrant team for '%s': %w", t.Name, err)
		}
		if selected == 0 {
			console.Infof("[+] Team '%s' will be created as new team in FireHydrant.\n", t.Name)
			continue
		}
		if err := store.UseQueries(ctx).LinkExtTeam(ctx, store.LinkExtTeamParams{
			ID:       t.ID,
			FhTeamID: sql.NullString{String: fhTeam.ID, Valid: true},
		}); err != nil {
			return fmt.Errorf("linking team '%s' to FireHydrant: %w", t.Name, err)
		}
	}
	return nil
}

func importUsers(ctx context.Context, provider pager.Pager, fh *firehydrant.Client) error {
	// Get all of the users registered from Pager Provider (e.g. PagerDuty)
	var err error
	console.Spin(func() {
		err = provider.LoadUsers(ctx)
	}, "Fetching all users from provider...")
	if err != nil {
		return fmt.Errorf("unable to fetch users from provider: %w", err)
	}
	console.Successf("Loaded all users from %s.\n", provider.Kind())

	// Find out which users do not already have a FireHydrant account
	console.Spin(func() {
		err = fh.MatchUsers(ctx)
	}, "Matching users with existing FireHydrant users by email...")
	if err != nil {
		return fmt.Errorf("unable to match users to FireHydrant: %w", err)
	}

	// Find out which users should be pre-created in FireHydrant via SCIM / matched to existing user.
	unmatched, err := store.UseQueries(ctx).ListUnmatchedExtUsers(ctx)
	if err != nil {
		return fmt.Errorf("unable to list unmatched users: %w", err)
	}
	if len(unmatched) == 0 {
		console.Successf("All users are already matched to FireHydrant.\n")
		return nil
	}

	// Get padding number to pretty print the information in table-like view.
	idPad := console.PadStrings(unmatched, func(u store.ExtUser) int { return len(u.ID) })
	emailPad := console.PadStrings(unmatched, func(u store.ExtUser) int { return len(u.Email) })

	importOpts := []store.ExtUser{{ID: "[+] IMPORT ALL"}, {ID: "[<] SKIP ALL"}}
	importOpts = append(importOpts, unmatched...)
	selected, toImport, err := console.MultiSelectf(importOpts, func(u store.ExtUser) string {
		return fmt.Sprintf("%*s  %-*s  %s", idPad, u.ID, emailPad, u.Email, u.Name)
	}, "These users do not have a FireHydrant account. Which users should be created / merged in FireHydrant?")
	if err != nil {
		return fmt.Errorf("selecting users to import: %w", err)
	}
	switch selected[0] {
	case 0:
		console.Successf("[+] All users will be created in FireHydrant.\n")
	case 1:
		console.Warnf("[<] No users will be created in FireHydrant.\n")
		if err := store.UseQueries(ctx).DeleteUnmatchedExtUsers(ctx); err != nil {
			return fmt.Errorf("unable to delete unmatched users: %w", err)
		}
		return nil
	default:
		console.Warnf("Selected %d users to be imported to FireHydrant.\n", len(toImport))
	}

	// We now ask if all of them are to be created as new users, or if they should be matched to existing FireHydrant users.
	console.Warnf("Please match the following users to a FireHydrant account.\n")
	fhUsers, err := fh.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("unable to list FireHydrant users: %w", err)
	}

	namePad := console.PadStrings(fhUsers, func(u store.FhUser) int { return len(u.Name) })

	matchOpts := []store.FhUser{{Name: "[+] CREATE THE REST OF USERS AS NEW"}, {Name: "[+] CREATE USER AS NEW"}}
	matchOpts = append(matchOpts, fhUsers...)
	for i, u := range toImport {
		selected, fhUser, err := console.Selectf(matchOpts, func(u store.FhUser) string {
			return fmt.Sprintf("%*s  %s", namePad, u.Name, u.Email)
		}, fmt.Sprintf("[%03d/%03d] Which FireHydrant user should '%s' be imported to?", i+1, len(toImport), u.Name))
		if err != nil {
			return fmt.Errorf("selecting FireHydrant user for '%s': %w", u.Name, err)
		}
		switch selected {
		case 0:
			console.Infof("[+] All users will be created in FireHydrant.\n")
			for _, u := range unmatched[i:] {
				fhUser, err := fh.CreateUser(ctx, &u)
				if err != nil {
					console.Warnf("unable to create user '%s': %s\n", u.Email, err.Error())
					continue
				}
				if err := store.UseQueries(ctx).LinkExtUser(ctx, store.LinkExtUserParams{
					ID:       u.ID,
					FhUserID: sql.NullString{String: fhUser.ID, Valid: true},
				}); err != nil {
					console.Warnf("unable to link user '%s': %s\n", u.Email, err.Error())
					continue
				}
			}
			return nil
		case 1:
			console.Infof("[+] User '%s (%s)' will be created in FireHydrant.\n", u.Name, u.Email)
			scimUser, err := fh.CreateUser(ctx, &u)
			if err != nil {
				return fmt.Errorf("creating user '%s': %w", u.Name, err)
			}
			fhUser = *scimUser
			fallthrough
		default:
			if err := store.UseQueries(ctx).LinkExtUser(ctx, store.LinkExtUserParams{
				ID:       u.ID,
				FhUserID: sql.NullString{String: fhUser.ID, Valid: true},
			}); err != nil {
				console.Warnf("unable to link user '%s': %s\n", u.Email, err.Error())
				continue
			}
			console.Successf("[=] User '%s' linked to FireHydrant user '%s'.\n", u.Email, fhUser.Email)
		}
	}
	return nil
}
