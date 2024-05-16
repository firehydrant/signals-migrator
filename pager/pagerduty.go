package pager

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
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

var (
	pdTeamInterface string

	pdTeamInterfaces = []string{"team", "service"}
)

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
	return "PagerDuty"
}

// TeamInterfaces defines the available abstraction of a team from PagerDuty.
// When "team" is selected, the team is fetched from PagerDuty and imported as-is.
// When "service" is selected, a "service team" will be created as a proxy team, linked to regular PagerDuty teams,
// via ext_team_groups table. When populating user members, the service team will query all the linked teams for
// all their user members.
func (p *PagerDuty) TeamInterfaces() []string {
	return pdTeamInterfaces
}

func (p *PagerDuty) UseTeamInterface(interfaceName string) error {
	if slices.Contains(pdTeamInterfaces, interfaceName) {
		pdTeamInterface = interfaceName
		return nil
	}
	return fmt.Errorf("unknown team interface '%s'", interfaceName)
}

func (p *PagerDuty) Teams(ctx context.Context) ([]store.ExtTeam, error) {
	switch pdTeamInterface {
	case "team":
		return store.UseQueries(ctx).ListNonGroupExtTeams(ctx)
	case "service":
		return store.UseQueries(ctx).ListGroupExtTeams(ctx)
	case "":
		return nil, fmt.Errorf("team interface not set")
	default:
		return nil, fmt.Errorf("unknown team interface '%s'", pdTeamInterface)
	}
}

func (p *PagerDuty) LoadUsers(ctx context.Context) error {
	opts := pagerduty.ListUsersOptions{
		Offset: 0,
	}

	for {
		resp, err := p.client.ListUsersWithContext(ctx, opts)
		if err != nil {
			return fmt.Errorf("listing users: %w", err)
		}

		for _, user := range resp.Users {
			if err := store.UseQueries(ctx).InsertExtUser(ctx, store.InsertExtUserParams{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,

				Annotations: fmt.Sprintf("[PagerDuty] %s %s", user.Email, user.HTMLURL),
			}); err != nil {
				return fmt.Errorf("saving user to db: %w", err)
			}
		}

		// Results are paginated, so break if we're on the last page.
		if !resp.More {
			break
		}
		opts.Offset += uint(len(resp.Users))
	}

	return nil
}

func (p *PagerDuty) LoadTeams(ctx context.Context) error {
	switch pdTeamInterface {
	case "team":
		return p.loadTeams(ctx)
	case "service":
		return p.loadServices(ctx)
	case "":
		return fmt.Errorf("team interface not set")
	default:
		return fmt.Errorf("unknown team interface '%s'", pdTeamInterface)
	}
}

func (p *PagerDuty) loadTeams(ctx context.Context) error {
	opts := pagerduty.ListTeamOptions{
		Offset: 0,
	}

	for {
		resp, err := p.client.ListTeamsWithContext(ctx, opts)
		if err != nil {
			return fmt.Errorf("listing teams: %w", err)
		}

		for _, team := range resp.Teams {
			if err := store.UseQueries(ctx).InsertExtTeam(ctx, store.InsertExtTeamParams{
				ID:   team.ID,
				Name: team.Name,
				// PagerDuty does not expose slug, so we can safely generate one.
				Slug: slug.Make(team.Name),

				Annotations: fmt.Sprintf("[PagerDuty] %s %s", team.Name, team.HTMLURL),
			}); err != nil {
				return fmt.Errorf("saving team '%s (%s)' to db: %w", team.Name, team.ID, err)
			}
		}

		// Results are paginated, so break if we're on the last page.
		if !resp.More {
			break
		}
		opts.Offset += uint(len(resp.Teams))
	}

	return nil
}

func (p *PagerDuty) loadServices(ctx context.Context) error {
	opts := pagerduty.ListServiceOptions{
		Includes: []string{"teams"},
		Offset:   0,
	}

	q := store.UseQueries(ctx)

	for {
		resp, err := p.client.ListServicesWithContext(ctx, opts)
		if err != nil {
			return fmt.Errorf("listing services: %w", err)
		}

		for _, service := range resp.Services {
			if err := q.InsertExtTeam(ctx, store.InsertExtTeamParams{
				ID:   service.ID,
				Name: service.Name,
				// PagerDuty does not expose "Slug", so we can safely generate one.
				Slug:    slug.Make(service.Name),
				IsGroup: 1,

				Annotations: fmt.Sprintf("[PagerDuty] %s %s", service.Name, service.HTMLURL),
			}); err != nil {
				return fmt.Errorf("saving service '%s (%s)' as team to db: %w", service.Name, service.ID, err)
			}
			for _, team := range service.Teams {
				if err := q.InsertExtTeam(ctx, store.InsertExtTeamParams{
					ID:   team.ID,
					Name: team.Name,
					// PagerDuty does not expose "Slug", so we can safely generate one.
					Slug: slug.Make(team.Name),
				}); err != nil {
					if strings.Contains(err.Error(), "UNIQUE constraint") {
						// Assume that team was already imported from another service.
						console.Warnf("Team %s has been imported, skipping duplicate...\n", service.ID, team.ID)
					} else {
						return fmt.Errorf("saving team '%s (%s)' to db: %w", team.Name, team.ID, err)
					}
				}
				if err := q.InsertExtTeamGroup(ctx, store.InsertExtTeamGroupParams{
					GroupTeamID:  service.ID,
					MemberTeamID: team.ID,
				}); err != nil {
					if strings.Contains(err.Error(), "UNIQUE constraint") {
						// This should never happen, unless it's on a dirty database. Warn users anyway.
						console.Warnf("Service %s already has team %s, skipping duplicate...\n", service.ID, team.ID)
					} else {
						return fmt.Errorf("saving '%s (%s)' team as proxy for '%s (%s)' service: %w", team.Name, team.ID, service.Name, service.ID, err)
					}
				}
			}
		}
		// Results are paginated, so break if we're on the last page.
		if !resp.More {
			break
		}
		opts.Offset += uint(len(resp.Services))
	}

	return nil
}

func (p *PagerDuty) LoadTeamMembers(ctx context.Context) error {
	switch pdTeamInterface {
	case "team":
		return p.loadTeamMembers(ctx)
	case "service":
		return p.loadServiceTeamMembers(ctx)
	case "":
		return fmt.Errorf("team interface not set")
	default:
		return fmt.Errorf("unknown team interface '%s'", pdTeamInterface)
	}
}

func (p *PagerDuty) loadTeamMembers(ctx context.Context) error {
	teams, err := store.UseQueries(ctx).ListTeams(ctx)
	if err != nil {
		return fmt.Errorf("listing teams: %w", err)
	}

	for _, team := range teams {
		if err := p.loadMembers(ctx, team.ID); err != nil {
			return fmt.Errorf("loading team members: %w", err)
		}
	}
	return nil
}

func (p *PagerDuty) loadServiceTeamMembers(ctx context.Context) error {
	teams, err := store.UseQueries(ctx).ListNonGroupExtTeams(ctx)
	if err != nil {
		return fmt.Errorf("listing teams: %w", err)
	}

	for _, team := range teams {
		if err := p.loadMembers(ctx, team.ID); err != nil {
			return fmt.Errorf("loading service team members: %w", err)
		}
	}
	return nil
}

// loadMembers loads members of a team from PagerDuty API and saves them to the database.
// teamID is the team ID which will be used in HTTP query to PagerDuty API, while memberOfTeamID is the
// reference which will be used in database.
func (p *PagerDuty) loadMembers(ctx context.Context, teamID string) error {
	// PagerDuty REST API technically supports `includes[]=user` but it's not exposed in Go SDK.
	// As such, we currently assume the user is already present in the database and only save the relationship.
	opts := pagerduty.ListTeamMembersOptions{
		Offset: 0,
	}
	q := store.UseQueries(ctx)

	for {
		resp, err := p.client.ListTeamMembers(ctx, teamID, opts)
		if err != nil {
			return err
		}

		for _, member := range resp.Members {
			if err := q.InsertExtMembership(ctx, store.InsertExtMembershipParams{
				TeamID: teamID,
				UserID: member.User.ID,
			}); err != nil {
				if strings.Contains(err.Error(), "FOREIGN KEY constraint") {
					console.Warnf("User %s not found for team %s, skipping...\n", member.User.ID, teamID)
				} else if strings.Contains(err.Error(), "UNIQUE constraint") {
					console.Warnf("User %s already exists for team %s, skipping duplicate...\n", member.User.ID, teamID)
				} else {
					return fmt.Errorf("saving team member: %w", err)
				}
			}
		}

		// Results are paginated, so break if we're on the last page.
		if !resp.More {
			break
		}
		opts.Offset += uint(len(resp.Members))
	}
	return nil
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
	teamIDs := []string{}
	for _, team := range policy.Teams {
		teamIDs = append(teamIDs, team.ID)
	}
	ep := store.InsertExtEscalationPolicyParams{
		ID:          policy.ID,
		Name:        policy.Name,
		Description: policy.Description,
		RepeatLimit: int64(policy.NumLoops),
		Annotations: fmt.Sprintf("[PagerDuty]\n  %s %s\n  Teams: %v", policy.Name, policy.HTMLURL, teamIDs),
	}

	if err := store.UseQueries(ctx).InsertExtEscalationPolicy(ctx, ep); err != nil {
		return fmt.Errorf("saving escalation policy %s (%s): %w", ep.Name, ep.ID, err)
	}

	for _, team := range policy.Teams {
		if err := store.UseQueries(ctx).UpdateExtEscalationPolicyTeam(ctx, store.UpdateExtEscalationPolicyTeamParams{
			ID:     ep.ID,
			TeamID: sql.NullString{Valid: true, String: team.ID},
		}); err == nil {
			break
		}
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
