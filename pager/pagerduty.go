package pager

import (
	"context"
	"fmt"
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
	s := store.InsertExtScheduleParams{
		ID:       schedule.ID + "-" + layer.ID,
		Name:     schedule.Name + "-" + layer.Name,
		Timezone: schedule.TimeZone,

		// Add fallback values and override them later if API provides valid information.
		Description: fmt.Sprintf("%s (%s)", schedule.Description, layer.Name),
		HandoffTime: "11:00",
		HandoffDay:  "wednesday",
		Strategy:    "weekly",
	}

	// TODO: support custom strategy. For now:
	// - any less than "weekly" is considered "daily"
	// - any more than "weekly" is considered "weekly" anyway
	if layer.RotationTurnLengthSeconds < 60*60*24*7 {
		s.Strategy = "daily"
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

	// TODO: add restrictions

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
