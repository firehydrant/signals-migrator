package pager

import (
	"context"

	"github.com/PagerDuty/go-pagerduty"
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

func (p *PagerDuty) PopulateTeamSchedules(ctx context.Context, team *Team) error {
	// TODO: implement
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
		resp, err := p.client.ListUsers(opts)
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
