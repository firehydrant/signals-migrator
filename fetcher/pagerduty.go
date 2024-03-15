package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/firehydrant/signals-migrator/types"
)

type PagerDuty struct {
	client pdClient
	rc     *pagerduty.Client
}

type pdClient interface {
	ListUsersWithContext(context.Context, pagerduty.ListUsersOptions) (*pagerduty.ListUsersResponse, error)
	ListTeamsWithContext(context.Context, pagerduty.ListTeamOptions) (*pagerduty.ListTeamResponse, error)
	ListTeamMembers(context.Context, string, pagerduty.ListTeamMembersOptions) (*pagerduty.ListTeamMembersResponse, error)
	GetUserWithContext(context.Context, string, pagerduty.GetUserOptions) (*pagerduty.User, error)
	GetTeamWithContext(context.Context, string) (*pagerduty.Team, error)
	ListSchedulesWithContext(context.Context, pagerduty.ListSchedulesOptions) (*pagerduty.ListSchedulesResponse, error)
	GetScheduleWithContext(context.Context, string, pagerduty.GetScheduleOptions) (*pagerduty.Schedule, error)
	Do(*http.Request, bool) (*http.Response, error)
}

func NewPagerDuty(apiKey string) *PagerDuty {
	return &PagerDuty{
		client: pagerduty.NewClient(apiKey),

		// for easily poking at the client interface
		rc: pagerduty.NewClient(apiKey),
	}
}

func (p *PagerDuty) Users(ctx context.Context) ([]*types.User, error) {
	var users []*types.User

	pdUsers, err := p.listUsers(ctx)
	if err != nil {
		return users, err
	}

	for _, u := range pdUsers {
		us := &types.User{
			Email: u.Email,
			Resource: types.Resource{
				RemoteID: u.ID,
				Name:     u.Name,
			},
		}

		users = append(users, us)
	}

	return users, nil
}

func (p *PagerDuty) User(ctx context.Context, id string) (*types.User, error) {
	u, err := p.getUser(ctx, id)
	if err != nil {
		return nil, err
	}

	t := &types.User{
		Email: u.Email,
		Resource: types.Resource{
			RemoteID: u.ID,
			Name:     u.Name,
		},
	}

	return t, nil
}

func (p *PagerDuty) Team(ctx context.Context, id string) (*types.Team, error) {
	u, err := p.getTeam(ctx, id)
	if err != nil {
		return nil, err
	}

	t := &types.Team{
		Resource: types.Resource{
			RemoteID: u.ID,
			Name:     u.Name,
		},
	}

	sch, err := p.TeamSchedules(ctx, id)
	if err != nil {
		return t, err
	}

	t.Schedules = sch

	var members []*types.User

	tm, err := p.getTeamMembers(ctx, id)
	if err != nil {
		return t, err
	}

	for _, m := range tm.Members {
		u, err := p.User(ctx, m.User.ID)
		if err != nil {
			return t, err
		}

		members = append(members, u)
	}

	t.Members = members

	return t, nil
}

func (p *PagerDuty) Teams(ctx context.Context) ([]*types.Team, error) {
	var teams []*types.Team

	pdTeams, err := p.listTeams(ctx)
	if err != nil {
		return teams, err
	}

	for _, t := range pdTeams {
		team, err := p.Team(ctx, t.ID)
		if err != nil {
			return teams, fmt.Errorf("unable to fetch team %s: %w", t.Name, err)
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func (p *PagerDuty) TeamSchedules(ctx context.Context, teamID string) ([]*types.Schedule, error) {
	var schedules []*types.Schedule

	pdSchedules, err := p.listSchedules(ctx, teamID)
	if err != nil {
		return schedules, err
	}

	for _, s := range pdSchedules {
		schedule, err := p.Schedule(ctx, s.ID)
		if err != nil {
			continue
		}

		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

func (p *PagerDuty) Schedule(ctx context.Context, id string) (*types.Schedule, error) {
	s, err := p.getSchedule(ctx, id)
	if err != nil {
		return nil, err
	}

	var strategy types.ScheduleStrategy
	if len(s.ScheduleLayers) == 0 {
		return nil, fmt.Errorf("no schedule layers found")
	}

	if s.ScheduleLayers[0].RotationTurnLengthSeconds == 86400 {
		strategy = types.Daily
	} else if s.ScheduleLayers[0].RotationTurnLengthSeconds == 604800 {
		strategy = types.Weekly
	}

	handoff, err := time.Parse(time.RFC3339, s.ScheduleLayers[0].RotationVirtualStart)
	if err != nil {
		return nil, err
	}

	sc := &types.Schedule{
		TimeZone:    s.TimeZone,
		Strategy:    strategy,
		HandoffTime: handoff,
		HandoffDay:  handoff.Weekday(),
		Source:      s,
		Resource: types.Resource{
			ID:          s.ID,
			Name:        s.Name,
			Description: s.Description,
			RemoteID:    s.ID,
		},
	}

	return sc, nil
}

func (p *PagerDuty) getTeam(ctx context.Context, id string) (*pagerduty.Team, error) {
	return p.client.GetTeamWithContext(ctx, id)
}

func (p *PagerDuty) getTeamMembers(ctx context.Context, teamID string) (*pagerduty.ListTeamMembersResponse, error) {
	return p.client.ListTeamMembers(ctx, teamID, pagerduty.ListTeamMembersOptions{})
}

func (p *PagerDuty) getUser(ctx context.Context, id string) (*pagerduty.User, error) {
	return p.client.GetUserWithContext(ctx, id, pagerduty.GetUserOptions{})
}

func (p *PagerDuty) listTeams(ctx context.Context) ([]pagerduty.Team, error) {
	teams := []pagerduty.Team{}

	o := pagerduty.ListTeamOptions{}

	for {
		resp, err := p.client.ListTeamsWithContext(ctx, o)

		if err != nil {
			return teams, err
		}

		teams = append(teams, resp.Teams...)

		if !resp.More {
			break
		}

		o.Offset = resp.Offset
	}

	return teams, nil
}

func (p *PagerDuty) listUsers(ctx context.Context) ([]pagerduty.User, error) {
	users := []pagerduty.User{}

	o := pagerduty.ListUsersOptions{}
	for {
		resp, err := p.client.ListUsersWithContext(ctx, o)

		if err != nil {
			return users, err
		}

		users = append(users, resp.Users...)

		if !resp.More {
			break
		}

		o.Offset = resp.Offset
	}

	return users, nil
}

func (p *PagerDuty) listSchedules(ctx context.Context, teamID string) ([]pagerduty.Schedule, error) {
	schedules := []pagerduty.Schedule{}

	var offset uint
	offset = 0

	for {
		r, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.pagerduty.com/schedules?team_ids[]=%s&offset=%v", teamID, offset), nil)
		if err != nil {
			return schedules, err
		}

		resp, err := p.client.Do(r, true)
		if err != nil {
			return schedules, err
		}

		var result pagerduty.ListSchedulesResponse
		if err = p.decodeJSON(resp, &result); err != nil {
			return schedules, err
		}

		schedules = append(schedules, result.Schedules...)

		if !result.More {
			break
		}

		offset = result.Offset
	}

	return schedules, nil
}

func (p *PagerDuty) getSchedule(ctx context.Context, id string) (*pagerduty.Schedule, error) {
	return p.client.GetScheduleWithContext(ctx, id, pagerduty.GetScheduleOptions{})
}

// copied from /go/pkg/mod/github.com/!pager!duty/go-pagerduty@v1.8.0/schedule.go
func (p *PagerDuty) decodeJSON(resp *http.Response, payload interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	return json.Unmarshal(body, payload)
}
