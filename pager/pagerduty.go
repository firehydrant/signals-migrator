package pager

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/firehydrant/signals-migrator/console"
	"github.com/firehydrant/signals-migrator/store"
	"github.com/gosimple/slug"
)

type PagerDuty struct {
	client *pagerduty.Client
}

func NewPagerDuty(apiKey string) *PagerDuty {
	return &PagerDuty{
		client: pagerduty.NewClient(apiKey),
	}
}

func NewPagerDutyWithURL(apiKey, url string) *PagerDuty {
	return &PagerDuty{
		client: pagerduty.NewClient(apiKey, pagerduty.WithAPIEndpoint(url)),
	}
}

func (p *PagerDuty) Kind() string {
	return "pagerduty"
}

func (p *PagerDuty) LoadSchedules(ctx context.Context) error {
	opts := pagerduty.ListSchedulesOptions{Includes: []string{"schedule_layers"}}
	for {
		resp, err := p.client.ListSchedulesWithContext(ctx, opts)
		if err != nil {
			return err
		}

		for _, schedule := range resp.Schedules {
			if err := p.saveScheduleToDB(ctx, schedule); err != nil {
				return fmt.Errorf("saving schedule to db: %w", err)
			}
		}

		// Results are paginated, so break if we're on the last page.
		if !resp.More {
			break
		}
		opts.Offset += uint(len(resp.Schedules))
	}

	return nil
}

func (p *PagerDuty) saveScheduleToDB(ctx context.Context, schedule pagerduty.Schedule) error {
	for _, layer := range schedule.ScheduleLayers {
		if err := p.saveLayerToDB(ctx, schedule, layer); err != nil {
			return fmt.Errorf("saving layer to db: %w", err)
		}
	}
	return nil
}

func (p *PagerDuty) saveLayerToDB(ctx context.Context, schedule pagerduty.Schedule, layer pagerduty.ScheduleLayer) error {
	desc := fmt.Sprintf("%s (%s)", schedule.Description, layer.Name)
	desc = strings.TrimSpace(desc)

	s := store.InsertExtScheduleParams{
		ID:       schedule.ID + "-" + layer.ID,
		Name:     schedule.Name + " - " + layer.Name,
		Timezone: schedule.TimeZone,

		// Add fallback values and override them later if API provides valid information.
		Description:   desc,
		HandoffTime:   "11:00:00",
		HandoffDay:    "wednesday",
		Strategy:      "weekly",
		ShiftDuration: "",
	}

	switch layer.RotationTurnLengthSeconds {
	case 60 * 60 * 24:
		s.Strategy = "daily"
	case 60 * 60 * 24 * 7:
		s.Strategy = "weekly"
	default:
		s.Strategy = "custom"
		s.ShiftDuration = fmt.Sprintf("PT%dS", layer.RotationTurnLengthSeconds)

		now := time.Now()
		loc, err := time.LoadLocation(schedule.TimeZone)
		if err == nil {
			now = now.In(loc)
		} else {
			console.Warnf("unable to parse timezone '%v', using current machine's local time", schedule.TimeZone)
		}
		s.StartTime = now.Format(time.RFC3339)
	}
	virtualStart, err := time.Parse(time.RFC3339, layer.RotationVirtualStart)
	if err == nil {
		s.HandoffTime = virtualStart.Format(time.TimeOnly)
		s.HandoffDay = strings.ToLower(virtualStart.Weekday().String())
	} else {
		console.Errorf("unable to parse virtual start time '%v', assuming default values", layer.RotationVirtualStart)
	}

	q := store.UseQueries(ctx)
	if err := q.InsertExtSchedule(ctx, s); err != nil {
		return fmt.Errorf("saving schedule: %w", err)
	}

	for _, team := range schedule.Teams {
		if err := q.InsertExtScheduleTeam(ctx, store.InsertExtScheduleTeamParams{
			ScheduleID: s.ID,
			TeamID:     team.ID,
		}); err != nil {
			if strings.Contains(err.Error(), "FOREIGN KEY constraint") {
				console.Warnf("Team %s not found for schedule %s, skipping...\n", team.ID, s.ID)
			} else if strings.Contains(err.Error(), "UNIQUE constraint") {
				console.Warnf("Team %s already exists for schedule %s, skipping duplicate...\n", team.ID, s.ID)
			} else {
				return fmt.Errorf("saving schedule team: %w", err)
			}
		}
	}

	for _, user := range layer.Users {
		if err := q.InsertExtScheduleMember(ctx, store.InsertExtScheduleMemberParams{
			ScheduleID: s.ID,
			UserID:     user.User.ID,
		}); err != nil {
			if strings.Contains(err.Error(), "FOREIGN KEY constraint") {
				console.Warnf("User %s not found for schedule %s, skipping...\n", user.User.ID, s.ID)
			} else if strings.Contains(err.Error(), "UNIQUE constraint") {
				console.Warnf("User %s already exists for schedule %s, skipping duplicate...\n", user.User.ID, s.ID)
			} else {
				return fmt.Errorf("saving schedule user: %w", err)
			}
		}
	}

	for i, restriction := range layer.Restrictions {
		switch restriction.Type {
		case "daily_restriction":
			for day := range 7 {
				start, err := time.Parse(time.TimeOnly, restriction.StartTimeOfDay)
				if err != nil {
					return fmt.Errorf("parsing start time of day '%s': %w", restriction.StartTimeOfDay, err)
				}
				end := start.Add(time.Duration(restriction.DurationSeconds) * time.Second)

				dayStr := strings.ToLower(time.Weekday(day).String())
				r := store.InsertExtScheduleRestrictionParams{
					ScheduleID:       s.ID,
					RestrictionIndex: fmt.Sprintf("%d-%d", i, day),
					StartDay:         dayStr,
					EndDay:           dayStr,
					StartTime:        start.Format(time.TimeOnly),
					EndTime:          end.Format(time.TimeOnly),
				}
				if err := q.InsertExtScheduleRestriction(ctx, r); err != nil {
					return fmt.Errorf("saving daily restriction: %w", err)
				}
			}
		case "weekly_restriction":
			start, err := time.Parse(time.TimeOnly, restriction.StartTimeOfDay)
			if err != nil {
				return fmt.Errorf("parsing start time of day '%s': %w", restriction.StartTimeOfDay, err)
			}
			// 0000-01-01 is a Saturday, so we need to adjust +1 such that when
			// restriction.StartDayOfWeek is 0, it yields Sunday.
			start = start.AddDate(0, 0, int(restriction.StartDayOfWeek+1))
			end := start.Add(time.Duration(restriction.DurationSeconds) * time.Second)

			r := store.InsertExtScheduleRestrictionParams{
				ScheduleID:       s.ID,
				RestrictionIndex: strconv.Itoa(i),
				StartDay:         strings.ToLower(start.Weekday().String()),
				EndDay:           strings.ToLower(end.Weekday().String()),
				StartTime:        start.Format(time.TimeOnly),
				EndTime:          end.Format(time.TimeOnly),
			}
			if err := q.InsertExtScheduleRestriction(ctx, r); err != nil {
				return fmt.Errorf("saving weekly restriction: %w", err)
			}
		default:
			console.Warnf("Unknown schedule restriction type '%s' for schedule '%s', skipping...\n", restriction.Type, s.ID)
		}
	}

	return nil
}

func (p *PagerDuty) LoadEscalationPolicies(ctx context.Context) error {
	opts := pagerduty.ListEscalationPoliciesOptions{
		Includes: []string{"targets"},
	}

	for {
		resp, err := p.client.ListEscalationPoliciesWithContext(ctx, opts)
		if err != nil {
			return err
		}

		for _, policy := range resp.EscalationPolicies {
			if err := p.saveEscalationPolicyToDB(ctx, policy); err != nil {
				return fmt.Errorf("saving escalation policy to db: %w", err)
			}
		}

		if !resp.More {
			break
		}
		opts.Offset += uint(len(resp.EscalationPolicies))
	}

	return nil
}

func (p *PagerDuty) saveEscalationPolicyToDB(ctx context.Context, policy pagerduty.EscalationPolicy) error {
	ep := store.InsertExtEscalationPolicyParams{
		ID:          policy.ID,
		Name:        policy.Name,
		Description: policy.Description,
		RepeatLimit: int64(policy.NumLoops),
	}
	if err := store.UseQueries(ctx).InsertExtEscalationPolicy(ctx, ep); err != nil {
		return fmt.Errorf("saving escalation policy %s (%s): %w", ep.Name, ep.ID, err)
	}

	// PagerDuty's Escalation Rule is equivalent to FireHydrant Escalation Policy Step.
	for i, rule := range policy.EscalationRules {
		if err := p.saveEscalationPolicyStepToDB(ctx, ep.ID, rule, int64(i)); err != nil {
			return fmt.Errorf("saving escalation rule to db: %w", err)
		}
	}
	return nil
}

func (p *PagerDuty) saveEscalationPolicyStepToDB(
	ctx context.Context,
	policyID string,
	rule pagerduty.EscalationRule,
	position int64,
) error {
	step := store.InsertExtEscalationPolicyStepParams{
		ID:                 rule.ID,
		EscalationPolicyID: policyID,
		Position:           position,
	}
	if rule.Delay > 0 {
		step.Timeout = fmt.Sprintf("PT%dM", rule.Delay)
	}
	if err := store.UseQueries(ctx).InsertExtEscalationPolicyStep(ctx, step); err != nil {
		return fmt.Errorf("saving escalation policy step: %w", err)
	}

	for i, target := range rule.Targets {
		if err := p.saveEscalationPolicyStepTargetToDB(ctx, step.ID, target, i); err != nil {
			return fmt.Errorf("saving escalation policy step target: %w", err)
		}
	}
	return nil
}

func (p *PagerDuty) saveEscalationPolicyStepTargetToDB(
	ctx context.Context,
	stepID string,
	pdTarget pagerduty.APIObject,
	position int,
) error {
	t := store.InsertExtEscalationPolicyStepTargetParams{EscalationPolicyStepID: stepID}
	switch pdTarget.Type {
	case "user", "user_reference":
		t.TargetType = store.TARGET_TYPE_USER
		t.TargetID = pdTarget.ID
	case "schedule", "schedule_reference":
		t.TargetType = store.TARGET_TYPE_SCHEDULE
		t.TargetID = pdTarget.ID
	default:
		console.Warnf("Unknown escalation policy step target type '%s' for step '%s', skipping...\n", pdTarget.Type, stepID)
		return nil
	}
	if err := store.UseQueries(ctx).InsertExtEscalationPolicyStepTarget(ctx, t); err != nil {
		return fmt.Errorf("saving escalation policy step target: %w", err)
	}
	return nil
}

func (p *PagerDuty) PopulateTeamMembers(ctx context.Context, team *Team) error {
	members := []*User{}
	opts := pagerduty.ListTeamMembersOptions{
		Offset: 0,
	}

	for {
		resp, err := p.client.ListTeamMembers(ctx, team.ID, opts)
		if err != nil {
			return err
		}

		for _, member := range resp.Members {
			members = append(members, &User{Resource: Resource{ID: member.User.ID}})
		}

		// Results are paginated, so break if we're on the last page.
		if !resp.More {
			break
		}
		opts.Offset += uint(len(resp.Members))
	}
	team.Members = members
	return nil
}

func (p *PagerDuty) ListTeams(ctx context.Context) ([]*Team, error) {
	teams := []*Team{}
	opts := pagerduty.ListTeamOptions{
		Offset: 0,
	}

	for {
		resp, err := p.client.ListTeamsWithContext(ctx, opts)
		if err != nil {
			return nil, err
		}

		for _, team := range resp.Teams {
			teams = append(teams, p.toTeam(team))
		}

		// Results are paginated, so break if we're on the last page.
		if !resp.More {
			break
		}
		opts.Offset += uint(len(resp.Teams))
	}
	return teams, nil
}

func (p *PagerDuty) toTeam(team pagerduty.Team) *Team {
	return &Team{
		// PagerDuty does not expose a slug, so generate one.
		Slug: slug.Make(team.Name),
		Resource: Resource{
			ID:   team.ID,
			Name: team.Name,
		},
	}
}

func (p *PagerDuty) ListUsers(ctx context.Context) ([]*User, error) {
	users := []*User{}
	opts := pagerduty.ListUsersOptions{
		Offset: 0,
	}

	for {
		resp, err := p.client.ListUsersWithContext(ctx, opts)
		if err != nil {
			return nil, err
		}

		for _, user := range resp.Users {
			users = append(users, p.toUser(user))
		}

		// Results are paginated, so break if we're on the last page.
		if !resp.More {
			break
		}
		opts.Offset += uint(len(resp.Users))
	}
	return users, nil
}

func (p *PagerDuty) toUser(user pagerduty.User) *User {
	return &User{
		Email: user.Email,
		Resource: Resource{
			ID:   user.ID,
			Name: user.Name,
		},
	}
}
