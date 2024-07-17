package tfrender

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/firehydrant/signals-migrator/console"
	"github.com/firehydrant/signals-migrator/store"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type TFRender struct {
	f *hclwrite.File

	provider *hclwrite.Body
	root     *hclwrite.Body

	// Output file directory.
	dir string
	// Output file name.
	filename string
}

func fhProviderVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return ">= 0.8.0"
	}

	// Recommend version which is used in the migration tool
	for _, dep := range bi.Deps {
		if dep.Path == "github.com/firehydrant/terraform-provider-firehydrant" {
			versionStr := strings.Split(dep.Version, "-")[0]
			// Version tag may be using Go-module commit hash syntax.
			// This means the version we use is potentially unknown to Terraform provider registry.
			// Bail out and use the latest version.
			if versionStr != dep.Version {
				break
			}
			// Go versioning may include v-prefix, but Terraform doesn't like it.
			v := strings.TrimPrefix(dep.Version, "v")
			return fmt.Sprintf("~> %s", v)
		}
	}

	return ">= 0.8.0"
}

func New(name string) (*TFRender, error) {
	baseDir := filepath.Dir(name)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("preparing output directory: %w", err)
	}
	baseName := filepath.Base(name)

	f := hclwrite.NewEmptyFile()
	root := f.Body()
	provider := root.AppendNewBlock("terraform", nil).Body().AppendNewBlock("required_providers", nil).Body()
	provider.SetAttributeValue("firehydrant", cty.ObjectVal(map[string]cty.Value{
		"source":  cty.StringVal("firehydrant/firehydrant"),
		"version": cty.StringVal(fhProviderVersion()),
	}))

	return &TFRender{
		f:        f,
		provider: provider,
		root:     root,
		dir:      baseDir,
		filename: baseName,
	}, nil
}

func (r *TFRender) Filepath() string {
	return filepath.Join(r.dir, r.filename)
}

func (r *TFRender) Filename() string {
	return r.filename
}

func (r *TFRender) Write(ctx context.Context) error {
	f, err := os.Create(r.Filepath())
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	toWrite := []func(context.Context) error{
		r.DataFireHydrantUsers,
		r.ResourceFireHydrantTeams,
		r.ResourceFireHydrantOnCallSchedule,
		r.ResourceFireHydrantEscalationPolicy,
	}

	for _, w := range toWrite {
		if err := w(ctx); err != nil {
			return err
		}
	}

	if _, err := f.Write(r.f.Bytes()); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	console.Successf("Terraform file has been written to %s\n", r.Filepath())

	return nil
}

func (r *TFRender) ResourceFireHydrantEscalationPolicy(ctx context.Context) error {
	policies, err := store.UseQueries(ctx).ListExtEscalationPolicies(ctx)
	if err != nil {
		return fmt.Errorf("querying escalation policies: %w", err)
	}
	for _, p := range policies {
		r.root.AppendNewline()

		b := r.root.AppendNewBlock("resource", []string{
			"firehydrant_escalation_policy",
			p.TFSlug(),
		}).Body()
		b.SetAttributeValue("name", cty.StringVal(p.Name))
		if p.Description != "" {
			b.SetAttributeValue("description", cty.StringVal(p.Description))
		}

		if p.TeamID.Valid && p.TeamID.String != "" {
			t, err := store.UseQueries(ctx).GetTeamByExtID(ctx, p.TeamID.String)
			if err != nil {
				return fmt.Errorf("querying team '%s' for policy '%s': %w", p.TeamID.String, p.Name, err)
			}
			b.SetAttributeTraversal("team_id", hcl.Traversal{
				hcl.TraverseRoot{Name: "firehydrant_team"},
				hcl.TraverseAttr{Name: t.TFSlug()},
				hcl.TraverseAttr{Name: "id"},
			})
		}

		if p.Annotations != "" {
			b.AppendNewline()
			r.AppendComment(b, p.Annotations)
		}

		steps, err := store.UseQueries(ctx).ListExtEscalationPolicySteps(ctx, p.ID)
		if err != nil {
			return fmt.Errorf("querying steps for policy '%s': %w", p.Name, err)
		}

		for _, s := range steps {
			b.AppendNewline()
			step := b.AppendNewBlock("step", nil).Body()
			step.SetAttributeValue("timeout", cty.StringVal(s.Timeout))

			targets, err := store.UseQueries(ctx).ListExtEscalationPolicyStepTargets(ctx, s.ID)
			if err != nil {
				return fmt.Errorf("querying targets for step %d of %s: %w", s.Position, p.Name, err)
			}

			for _, t := range targets {
				idTraversals := []hcl.Traversal{}

				switch t.TargetType {
				case store.TARGET_TYPE_USER:
					u, err := store.UseQueries(ctx).GetUserByExtID(ctx, t.TargetID)
					if err != nil {
						console.Errorf("querying user '%s' for step %d of %s: %s\n", t.TargetID, s.Position, p.Name, err.Error())
						continue
					}
					idTraversals = append(idTraversals, hcl.Traversal{ //nolint:staticcheck // See "safeguard" below
						hcl.TraverseRoot{Name: "data"},
						hcl.TraverseAttr{Name: "firehydrant_user"},
						hcl.TraverseAttr{Name: u.TFSlug()},
						hcl.TraverseAttr{Name: "id"},
					})
				case store.TARGET_TYPE_SCHEDULE:
					schedules, err := store.UseQueries(ctx).ListExtSchedulesLikeID(ctx, fmt.Sprintf(`%s%%`, t.TargetID))
					if err != nil {
						console.Errorf("querying schedule '%s' for step %d of %s: %s\n", t.TargetID, s.Position, p.Name, err.Error())
						continue
					}
					for _, schedule := range schedules {
						teams, err := store.UseQueries(ctx).ListTeamsByExtScheduleID(ctx, schedule.ID)
						if err != nil {
							return fmt.Errorf("querying teams for schedule '%s': %w", schedule.Name, err)
						}
						for _, team := range teams {
							slug := fmt.Sprintf("%s_%s", team.TFSlug(), schedule.TFSlug())
							idTraversals = append(idTraversals, hcl.Traversal{ //nolint:staticcheck // See "safeguard" below
								hcl.TraverseRoot{Name: "firehydrant_on_call_schedule"},
								hcl.TraverseAttr{Name: slug},
								hcl.TraverseAttr{Name: "id"},
							})
						}
					}
				default:
					console.Errorf("unknown target type '%s' for step %d of %s\n", t.TargetType, s.Position, p.Name)
					continue
				}

				for _, targetTraversal := range idTraversals { //nolint:staticcheck // Safeguard in case the switch-case above is changed.
					step.AppendNewline()
					target := step.AppendNewBlock("targets", nil).Body()
					target.SetAttributeValue("type", cty.StringVal(t.TargetType))
					target.SetAttributeTraversal("id", targetTraversal)
				}
			}
		}

		b.AppendNewline()
		b.SetAttributeValue("repetitions", cty.NumberIntVal(p.RepeatLimit))

		// TODO: add policy handoff
	}
	return nil
}

func (r *TFRender) ResourceFireHydrantOnCallSchedule(ctx context.Context) error {
	schedules, err := store.UseQueries(ctx).ListExtSchedules(ctx)
	if err != nil {
		return fmt.Errorf("querying schedules: %w", err)
	}

	for _, s := range schedules {
		teams, err := store.UseQueries(ctx).ListTeamsByExtScheduleID(ctx, s.ID)
		if err != nil {
			return fmt.Errorf("querying teams for schedule '%s': %w", s.Name, err)
		}
		for _, t := range teams {
			r.root.AppendNewline()

			b := r.root.AppendNewBlock("resource", []string{
				"firehydrant_on_call_schedule",
				fmt.Sprintf("%s_%s", t.TFSlug(), s.TFSlug()),
			}).Body()
			b.SetAttributeValue("name", cty.StringVal(s.Name))
			if s.Description != "" {
				b.SetAttributeValue("description", cty.StringVal(s.Description))
			}
			b.SetAttributeTraversal("team_id", hcl.Traversal{
				hcl.TraverseRoot{Name: "firehydrant_team"},
				hcl.TraverseAttr{Name: t.TFSlug()},
				hcl.TraverseAttr{Name: "id"},
			})
			b.SetAttributeValue("time_zone", cty.StringVal(s.Timezone))
			if s.Strategy == "custom" && s.StartTime != "" {
				b.SetAttributeValue("start_time", cty.StringVal(s.StartTime))
			}

			if t.Annotations != "" {
				b.AppendNewline()
				r.AppendComment(b, t.Annotations)
			}

			members, err := store.UseQueries(ctx).ListFhMembersByExtScheduleID(ctx, s.ID)
			if err != nil {
				return fmt.Errorf("querying members for schedule '%s': %w", s.Name, err)
			}

			memberList := []hclwrite.Tokens{}
			for _, m := range members {
				member := hcl.Traversal{
					hcl.TraverseRoot{Name: "data"},
					hcl.TraverseAttr{Name: "firehydrant_user"},
					hcl.TraverseAttr{Name: m.TFSlug()},
					hcl.TraverseAttr{Name: "id"},
				}
				memberList = append(memberList, hclwrite.TokensForTraversal(member))
			}

			b.AppendNewline()
			b.SetAttributeRaw("member_ids", hclwrite.TokensForTuple(memberList))

			b.AppendNewline()
			strategy := b.AppendNewBlock("strategy", []string{}).Body()
			strategy.SetAttributeValue("type", cty.StringVal(s.Strategy))
			if s.Strategy == "weekly" {
				strategy.SetAttributeValue("handoff_day", cty.StringVal(s.HandoffDay))
			}
			if s.Strategy == "custom" {
				strategy.SetAttributeValue("shift_duration", cty.StringVal(s.ShiftDuration))
			} else {
				strategy.SetAttributeValue("handoff_time", cty.StringVal(s.HandoffTime))
			}

			restrictions, err := store.UseQueries(ctx).ListExtScheduleRestrictionsByExtScheduleID(ctx, s.ID)
			if err != nil {
				return fmt.Errorf("querying restrictions for schedule '%s': %w", s.Name, err)
			}
			for _, r := range restrictions {
				b.AppendNewline()
				restriction := b.AppendNewBlock("restrictions", []string{}).Body()
				restriction.SetAttributeValue("start_day", cty.StringVal(r.StartDay))
				restriction.SetAttributeValue("start_time", cty.StringVal(r.StartTime))
				restriction.SetAttributeValue("end_day", cty.StringVal(r.EndDay))
				restriction.SetAttributeValue("end_time", cty.StringVal(r.EndTime))
			}
		}
	}
	return nil
}

func (r *TFRender) ResourceFireHydrantTeams(ctx context.Context) error {
	q := store.UseQueries(ctx)
	extTeams, err := q.ListTeamsToImport(ctx)
	if err != nil {
		return fmt.Errorf("querying teams: %w", err)
	}

	// Use hashmap to deduplicate import and membership.
	// There is probably a smarter way to do it in SQL, this just so happen to be easy and convenient.
	importedTeams := map[string]bool{}
	importedMembership := map[string]bool{}

	fhTeamBlocks := map[string]*hclwrite.Body{}
	for _, t := range extTeams {
		name := t.ValidName()
		tfSlug := t.TFSlug()

		if _, ok := fhTeamBlocks[name]; !ok {
			r.root.AppendNewline()
			fhTeamBlocks[name] = r.root.AppendNewBlock("resource", []string{"firehydrant_team", tfSlug}).Body()
			fhTeamBlocks[name].SetAttributeValue("name", cty.StringVal(name))
		}

		// For a given t in extTeams, they may be a "group team" which contains "member team" entities.
		// In those cases, the "member team" is merged into the "group team" in FireHydrant.
		// As such, the user members will be consolidated to the "group team".
		memberTeams, err := q.ListMemberExtTeams(ctx, t.ID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("querying member teams: %w", err)
			}
		}
		if memberTeams == nil {
			memberTeams = []store.ExtTeam{}
		}
		teamIDs := []string{t.ID}
		for _, mt := range memberTeams {
			teamIDs = append(teamIDs, mt.ID)
		}

		for _, teamID := range teamIDs {
			b := fhTeamBlocks[name]
			annotation, err := q.GetExtTeamAnnotation(ctx, teamID)
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					return fmt.Errorf("querying annotation for team '%s': %w", teamID, err)
				} else {
					annotation = ""
				}
			}
			if annotation != "" {
				b.AppendNewline()
				r.AppendComment(b, annotation)
			}

			members, err := q.ListFhMembersByExtTeamID(ctx, teamID)
			if err != nil {
				return fmt.Errorf("querying team members: %w", err)
			}
			for _, m := range members {
				if importedMembership[tfSlug+m.TFSlug()] {
					continue
				}

				b.AppendNewline()
				b.AppendNewBlock("memberships", []string{}).Body().
					SetAttributeTraversal("user_id", hcl.Traversal{
						hcl.TraverseRoot{Name: "data"},
						hcl.TraverseAttr{Name: "firehydrant_user"},
						hcl.TraverseAttr{Name: m.TFSlug()},
						hcl.TraverseAttr{Name: "id"},
					})
				importedMembership[tfSlug+m.TFSlug()] = true
			}
		}

		// If there is an existing FireHydrant team already, declare import to prevent duplication.
		if t.FhTeamID.Valid && t.FhTeamID.String != "" && !importedTeams[t.FhTeamID.String] {
			r.root.AppendNewline()
			importBody := r.root.AppendNewBlock("import", []string{}).Body()
			importBody.SetAttributeValue("id", cty.StringVal(t.FhTeamID.String))
			importBody.SetAttributeTraversal("to", hcl.Traversal{
				hcl.TraverseRoot{Name: "firehydrant_team"},
				hcl.TraverseAttr{Name: tfSlug},
			})
			importedTeams[t.FhTeamID.String] = true
		}
	}
	return nil
}

func (r *TFRender) DataFireHydrantUsers(ctx context.Context) error {
	users, err := store.UseQueries(ctx).ListFhUsers(ctx)
	if err != nil {
		return fmt.Errorf("querying users: %w", err)
	}
	for _, u := range users {
		r.root.AppendNewline()
		b := r.root.AppendNewBlock("data", []string{"firehydrant_user", u.TFSlug()}).Body()
		b.SetAttributeValue("email", cty.StringVal(u.Email))

		annotations, err := store.UseQueries(ctx).ListFhUserAnnotations(ctx, sql.NullString{String: u.ID, Valid: true})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			} else {
				return fmt.Errorf("querying annotations for user '%s': %w", u.Email, err)
			}
		}
		for _, a := range annotations {
			r.AppendComment(b, a)
		}
	}
	return nil
}

func (r *TFRender) AppendComment(b *hclwrite.Body, comment string) {
	str := strings.TrimSpace(comment)
	if str == "" {
		return
	}

	lines := strings.Split(str, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		comment := fmt.Sprintf("# %s\n", line)
		b.AppendUnstructuredTokens(
			hclwrite.Tokens{
				&hclwrite.Token{
					Type:  hclsyntax.TokenComment,
					Bytes: []byte(comment),

					SpacesBefore: 2,
				},
			},
		)
	}
}
