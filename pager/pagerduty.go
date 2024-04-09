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
		opts.Offset = resp.Offset
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
		console.Warnf("Found custom shift duration '%d seconds' for schedule '%s', rounding to daily value.\n", layer.RotationTurnLengthSeconds, s.Name)
		if layer.RotationTurnLengthSeconds < 60*60*24 {
			s.Strategy = "daily"
		} else {
			s.Strategy = "custom"
			s.ShiftDuration = fmt.Sprintf("P%dD", layer.RotationTurnLengthSeconds/60/60/24)
		}
	}
	virtualStart, err := time.Parse(time.RFC3339, layer.RotationVirtualStart)
	if err == nil {
		s.HandoffTime = fmt.Sprintf("%02d:%02d", virtualStart.Hour(), virtualStart.Minute())
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
				dayStr := strings.ToLower(time.Weekday(day).String())
				start, err := time.Parse(time.TimeOnly, restriction.StartTimeOfDay)
				if err != nil {
					return fmt.Errorf("parsing start time of day '%s': %w", restriction.StartTimeOfDay, err)
				}
				end := start.Add(time.Duration(restriction.DurationSeconds) * time.Second)

				r := store.InsertExtScheduleRestrictionParams{
					ScheduleID:       s.ID,
					RestrictionIndex: fmt.Sprintf("%d-%d", i, day),
					StartTime:        start.Format("15:04:05"),
					EndTime:          end.Format("15:04:05"),
					StartDay:         dayStr,
					EndDay:           dayStr,
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
			// restriction.StartDayOfWeek is 0, it is a Sunday.
			start = start.AddDate(0, 0, int(restriction.StartDayOfWeek)+1)
			end := start.Add(time.Duration(restriction.DurationSeconds) * time.Second)

			r := store.InsertExtScheduleRestrictionParams{
				ScheduleID:       s.ID,
				RestrictionIndex: strconv.Itoa(i),
				StartTime:        start.Format(time.TimeOnly),
				StartDay:         strings.ToLower(start.Weekday().String()),
				EndTime:          end.Format(time.TimeOnly),
				EndDay:           strings.ToLower(end.Weekday().String()),
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

func (p *PagerDuty) PopulateTeamMembers(ctx context.Context, team *Team) error {
	members := []*User{}
	opts := pagerduty.ListTeamMembersOptions{}

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
		opts.Offset = resp.Offset
	}
	team.Members = members
	return nil
}

func (p *PagerDuty) ListTeams(ctx context.Context) ([]*Team, error) {
	teams := []*Team{}
	opts := pagerduty.ListTeamOptions{}

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
		opts.Offset = resp.Offset
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
	opts := pagerduty.ListUsersOptions{}

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
		opts.Offset = resp.Offset
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
