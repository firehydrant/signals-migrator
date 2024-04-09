package tfrender

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/firehydrant/signals-migrator/store"
	"github.com/hashicorp/hcl/v2"
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
		return ">= 0.7.1"
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
			return fmt.Sprintf("~> %s", dep.Version)
		}
	}

	return ">= 0.7.1"
}

func New(dir string, name string) (*TFRender, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("preparing output directory: %w", err)
	}

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
		dir:      dir,
		filename: name,
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
	}

	for _, w := range toWrite {
		if err := w(ctx); err != nil {
			return err
		}
	}

	if _, err := f.Write(r.f.Bytes()); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

func (r *TFRender) ResourceFireHydrantOnCallSchedule(ctx context.Context) error {
	schedules, err := store.UseQueries(ctx).ListExtSchedules(ctx)
	if err != nil {
		return fmt.Errorf("querying schedules: %w", err)
	}

	for _, s := range schedules {
		teams, err := store.UseQueries(ctx).ListFhTeamsByExtScheduleID(ctx, s.ID)
		if err != nil {
			return fmt.Errorf("querying teams for schedule '%s': %w", s.Name, err)
		}
		for _, t := range teams {
			r.root.AppendNewline()

			b := r.root.AppendNewBlock("resource", []string{"firehydrant_on_call_schedule", s.TFSlug()}).Body()
			b.SetAttributeValue("name", cty.StringVal(s.Name))
			if s.Description != "" {
				b.SetAttributeValue("description", cty.StringVal(s.Description))
			}
			b.SetAttributeTraversal("team_id", hcl.Traversal{
				hcl.TraverseRoot{Name: "resource"},
				hcl.TraverseAttr{Name: "firehydrant_team"},
				hcl.TraverseAttr{Name: t.TFSlug()},
				hcl.TraverseAttr{Name: "id"},
			})
			b.SetAttributeValue("time_zone", cty.StringVal(s.Timezone))

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
			strategy.SetAttributeValue("handoff_time", cty.StringVal(s.HandoffTime))
			strategy.SetAttributeValue("handoff_day", cty.StringVal(s.HandoffDay))
		}
	}
	return nil
}

func (r *TFRender) ResourceFireHydrantTeams(ctx context.Context) error {
	extTeams, err := store.UseQueries(ctx).ListExtTeams(ctx)
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

		members, err := store.UseQueries(ctx).ListFhMembersByExtTeamID(ctx, t.ExtTeam().ID)
		if err != nil {
			return fmt.Errorf("querying team members: %w", err)
		}
		for _, m := range members {
			if importedMembership[tfSlug+m.TFSlug()] {
				continue
			}

			b := fhTeamBlocks[name]
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

		// If there is an existing FireHydrant team already, declare import to prevent duplication.
		if t.FhTeamID.Valid && t.FhTeamID.String != "" && !importedTeams[t.FhTeamID.String] {
			r.root.AppendNewline()
			importBody := r.root.AppendNewBlock("import", []string{}).Body()
			importBody.SetAttributeValue("id", cty.StringVal(t.FhTeamID.String))
			importBody.SetAttributeTraversal("to", hcl.Traversal{
				hcl.TraverseRoot{Name: "resource"},
				hcl.TraverseAttr{Name: "firehydrant_team"},
				hcl.TraverseAttr{Name: tfSlug},
				hcl.TraverseAttr{Name: "id"},
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
	}
	return nil
}
