package renderer

import (
	"github.com/firehydrant/signals-migrator/types"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type firehydrant struct {
	organization types.Organization
	file         *hclwrite.File
}

func NewFireHydrant(o types.Organization) firehydrant {
	return firehydrant{organization: o, file: hclwrite.NewEmptyFile()}
}

func (r *firehydrant) Hcl() (string, error) {
	err := r.Build(r.file)
	if err != nil {
		return "", err
	}

	return string(r.file.Bytes()), nil
}

func (r *firehydrant) Build(f *hclwrite.File) error {
	rootBody := f.Body()

	_ = rootBody.AppendNewBlock("provider", []string{"firehydrant"}).Body()

	for _, u := range r.organization.Users {
		r.user(rootBody, u)
	}

	for _, t := range r.organization.Teams {
		r.team(rootBody, t)
		for _, s := range t.Schedules {
			r.schedule(rootBody, t, s)
		}

		for _, s := range t.Schedules {
			r.ep(rootBody, t, s)
		}
	}

	return nil
}

func (r *firehydrant) team(f *hclwrite.Body, team *types.Team) {
	t := f.AppendNewBlock("data", []string{"firehydrant_team", team.ToResource()}).Body()
	t.SetAttributeValue("id", cty.StringVal(team.ID))
}

func (r *firehydrant) ep(f *hclwrite.Body, team *types.Team, schedule *types.Schedule) {
	e := f.AppendNewBlock("resource", []string{"firehydrant_escalation_policy", schedule.ToResource()}).Body()
	e.SetAttributeValue("name", cty.StringVal("default"))
	e.SetAttributeValue("description", cty.StringVal("Default escalation policy"))
	e.SetAttributeTraversal("team_id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "data",
		},
		hcl.TraverseAttr{
			Name: "firehydrant_team",
		},
		hcl.TraverseAttr{
			Name: team.ToResource(),
		},
		hcl.TraverseAttr{
			Name: "id",
		},
	})

	step := e.AppendNewBlock("step", nil).Body()
	step.SetAttributeValue("timeout", cty.StringVal("PT5M"))
	target := step.AppendNewBlock("targets", nil).Body()
	target.SetAttributeValue("type", cty.StringVal("OnCallSchedule"))
	target.SetAttributeTraversal("id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "resource",
		},
		hcl.TraverseAttr{
			Name: "firehydrant_on_call_schedule",
		},
		hcl.TraverseAttr{
			Name: schedule.ToResource(),
		},
		hcl.TraverseAttr{
			Name: "id",
		},
	})

	e.SetAttributeValue("repetitions", cty.NumberIntVal(2))
}

func (r *firehydrant) schedule(f *hclwrite.Body, team *types.Team, schedule *types.Schedule) {
	s := f.AppendNewBlock("resource", []string{"firehydrant_on_call_schedule", schedule.ToResource()}).Body()
	s.SetAttributeValue("name", cty.StringVal(schedule.Name))
	s.SetAttributeValue("description", cty.StringVal(schedule.Description))
	s.SetAttributeTraversal("team_id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "data",
		},
		hcl.TraverseAttr{
			Name: "firehydrant_team",
		},
		hcl.TraverseAttr{
			Name: team.ToResource(),
		},
		hcl.TraverseAttr{
			Name: "id",
		},
	})
	s.SetAttributeValue("time_zone", cty.StringVal(schedule.TimeZone))
	st := s.AppendNewBlock("strategy", nil).Body()
	st.SetAttributeValue("type", cty.StringVal(schedule.Strategy.String()))
	st.SetAttributeValue("handoff_time", cty.StringVal(schedule.HandoffTime.Format("15:04:05")))
	st.SetAttributeValue("handoff_day", cty.StringVal(schedule.HandoffDay.String()))

	if len(team.Members) > 0 {
		var mids []hclwrite.Tokens
		for _, user := range team.Members {
			mids = append(mids, hclwrite.TokensForTraversal(hcl.Traversal{
				hcl.TraverseRoot{
					Name: "data",
				},
				hcl.TraverseAttr{
					Name: "firehydrant_user",
				},
				hcl.TraverseAttr{
					Name: user.ToResource(),
				},
				hcl.TraverseAttr{
					Name: "id",
				},
			}))
		}

		s.SetAttributeRaw("member_ids", hclwrite.TokensForTuple(mids))
	}
}

func (r *firehydrant) user(f *hclwrite.Body, user *types.User) {
	t := f.AppendNewBlock("data", []string{"firehydrant_user", user.ToResource()}).Body()
	t.SetAttributeValue("email", cty.StringVal(user.Email))
}
