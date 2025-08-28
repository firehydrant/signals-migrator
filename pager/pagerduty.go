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
	// PagerDuty uses a single schedule with multiple layers.
	//   Within PagerDuty, these layers are compressed into a single schedule, with the later layers overriding the earlier layers in case of an overlap
	// In FireHydrant, we will have a single schedule with multiple rotations,
	//   Each layer will be converted into a rotation, however, rotations do not override each other
	//   members of each respective rotation will all be on call

	// PagerDuty Team-Schedule Relationship:
	// According to PagerDuty documentation: https://support.pagerduty.com/main/docs/teams#team-association-behavior
	// - "Adding users to a schedule via API will add the users to any associated Teams"
	// - "Associating a Team with a schedule via API will add the schedule's users to the Team"
	// This means all teams associated with a schedule end up with the same users.
	// Therefore, we can safely use the first team as the schedule owner without losing user information.

	teamID := ""
	if len(schedule.Teams) > 0 {
		teamID = schedule.Teams[0].ID // Use first team as owner - all teams would have same users anyway
	} else {
		console.Warnf("No teams found for schedule %s, skipping...\n", schedule.ID)
		return nil
	}

	scheduleParams := store.InsertExtScheduleV2Params{
		ID:               schedule.ID,
		Name:             schedule.Name,
		Description:      schedule.Description,
		Timezone:         schedule.TimeZone,
		TeamID:           teamID,
		SourceSystem:     "pagerduty",
		SourceScheduleID: schedule.ID,
	}

	q := store.UseQueries(ctx)
	if err := q.InsertExtScheduleV2(ctx, scheduleParams); err != nil {
		return fmt.Errorf("saving schedule: %w", err)
	}

	// Create rotation records under the parent schedule (each layer becomes a rotation)
	for i, layer := range schedule.ScheduleLayers {
		if err := p.saveLayerToDB(ctx, schedule.ID, layer, i); err != nil {
			return fmt.Errorf("saving layer to db: %w", err)
		}
	}
	return nil
}

func (p *PagerDuty) saveLayerToDB(ctx context.Context, scheduleID string, layer pagerduty.ScheduleLayer, layerOrder int) error {
	schedule, err := store.UseQueries(ctx).GetExtScheduleV2(ctx, scheduleID)
	if err != nil {
		return fmt.Errorf("getting schedule info: %w", err)
	}

	rotationName := layer.Name
	if rotationName == "" {
		rotationName = fmt.Sprintf("%srotation%d", schedule.Name, layerOrder+1)
	}

	desc := fmt.Sprintf("%s (%s)", schedule.Description, rotationName)
	desc = strings.TrimSpace(desc)

	rotationParams := store.InsertExtRotationParams{
		ID:            layer.ID,
		ScheduleID:    scheduleID,
		Name:          rotationName,
		Description:   desc,
		Strategy:      "weekly",
		ShiftDuration: "",
		StartTime:     "",
		HandoffTime:   "11:00:00",
		HandoffDay:    "wednesday",
		RotationOrder: int64(layerOrder),
	}

	switch layer.RotationTurnLengthSeconds {
	case 60 * 60 * 24:
		rotationParams.Strategy = "daily"
	case 60 * 60 * 24 * 7:
		rotationParams.Strategy = "weekly"
	default:
		rotationParams.Strategy = "custom"
		rotationParams.ShiftDuration = fmt.Sprintf("PT%dS", layer.RotationTurnLengthSeconds)

		now := time.Now()
		loc, err := time.LoadLocation(schedule.Timezone)
		if err == nil {
			now = now.In(loc)
		} else {
			console.Warnf("unable to parse timezone '%v', using current machine's local time", schedule.Timezone)
		}
		rotationParams.StartTime = now.Format(time.RFC3339)
	}
	virtualStart, err := time.Parse(time.RFC3339, layer.RotationVirtualStart)
	if err == nil {
		rotationParams.HandoffTime = virtualStart.Format(time.TimeOnly)
		rotationParams.HandoffDay = strings.ToLower(virtualStart.Weekday().String())
	} else {
		console.Errorf("unable to parse virtual start time '%v', assuming default values", layer.RotationVirtualStart)
	}

	q := store.UseQueries(ctx)
	if err := q.InsertExtRotation(ctx, rotationParams); err != nil {
		return fmt.Errorf("saving rotation: %w", err)
	}

	// ExtRotationMembers
	// Add only the users specifically assigned to this schedule layer.
	// Team members who aren't in the layer will still be on the team but won't be in this rotation.
	for i, user := range layer.Users {
		// Store the order of members as they appear in the API response
		// This preserves the exact order from PagerDuty
		if err := q.InsertExtRotationMember(ctx, store.InsertExtRotationMemberParams{
			RotationID:  layer.ID,
			UserID:      user.User.ID,
			MemberOrder: int64(i),
		}); err != nil {
			if strings.Contains(err.Error(), "FOREIGN KEY constraint") {
				console.Warnf("User %s not found for rotation %s, skipping...\n", user.User.ID, layer.ID)
			} else if strings.Contains(err.Error(), "UNIQUE constraint") {
				console.Warnf("User %s already exists for rotation %s, skipping duplicate...\n", user.User.ID, layer.ID)
			} else {
				return fmt.Errorf("saving rotation user: %w", err)
			}
		}
	}

	// ExtRotationRestrictions
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
				r := store.InsertExtRotationRestrictionParams{
					RotationID:       layer.ID,
					RestrictionIndex: fmt.Sprintf("%d-%d", i, day),
					StartDay:         dayStr,
					EndDay:           dayStr,
					StartTime:        start.Format(time.TimeOnly),
					EndTime:          end.Format(time.TimeOnly),
				}
				if err := q.InsertExtRotationRestriction(ctx, r); err != nil {
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

			r := store.InsertExtRotationRestrictionParams{
				RotationID:       layer.ID,
				RestrictionIndex: strconv.Itoa(i),
				StartDay:         strings.ToLower(start.Weekday().String()),
				EndDay:           strings.ToLower(end.Weekday().String()),
				StartTime:        start.Format(time.TimeOnly),
				EndTime:          end.Format(time.TimeOnly),
			}
			if err := q.InsertExtRotationRestriction(ctx, r); err != nil {
				return fmt.Errorf("saving weekly restriction: %w", err)
			}
		default:
			console.Warnf("Unknown schedule restriction type '%s' for rotation '%s', skipping...\n", restriction.Type, layer.ID)
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
	annotations := fmt.Sprintf("[PagerDuty]\n  %s %s", policy.Name, policy.HTMLURL)
	if len(policy.Teams) > 0 {
		annotations += "\n[Teams]"
		for _, team := range policy.Teams {
			annotations += fmt.Sprintf("\n  - %s", team.ID)
		}
	}
	if len(policy.Services) > 0 {
		annotations += "\n[Services]"
		for _, service := range policy.Services {
			annotations += fmt.Sprintf("\n  - %s %s %s", service.ID, service.HTMLURL, service.Summary)
		}
	}
	ep := store.InsertExtEscalationPolicyParams{
		ID:          policy.ID,
		Name:        policy.Name,
		Description: policy.Description,
		RepeatLimit: int64(policy.NumLoops),
		Annotations: annotations,
	}

	if err := store.UseQueries(ctx).InsertExtEscalationPolicy(ctx, ep); err != nil {
		return fmt.Errorf("saving escalation policy %s (%s): %w", ep.Name, ep.ID, err)
	}

	// We do our best to match first one from the list of teams / services. Since the full list is annotated in comments,
	// users can duplicate as needed. While it's possible for us to fan out and replicate, it's likely not the desired behavior
	// as in such scenario, user should really be merging the teams instead of duplicating escalation policies across teams.
	if pdTeamInterface == "service" {
		for _, service := range policy.Services {
			if err := store.UseQueries(ctx).UpdateExtEscalationPolicyTeam(ctx, store.UpdateExtEscalationPolicyTeamParams{
				ID:     ep.ID,
				TeamID: sql.NullString{Valid: true, String: service.ID},
			}); err == nil {
				break
			}
		}
	} else {
		for _, team := range policy.Teams {
			if err := store.UseQueries(ctx).UpdateExtEscalationPolicyTeam(ctx, store.UpdateExtEscalationPolicyTeamParams{
				ID:     ep.ID,
				TeamID: sql.NullString{Valid: true, String: team.ID},
			}); err == nil {
				break
			}
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
