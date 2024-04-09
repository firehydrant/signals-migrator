package pager

import (
	"context"

	"github.com/gosimple/slug"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/opsgenie/opsgenie-go-sdk-v2/team"
	"github.com/opsgenie/opsgenie-go-sdk-v2/user"
)

type Opsgenie struct {
	userClient *user.Client
	teamClient *team.Client
}

func NewOpsgenie(apiKey string) *Opsgenie {

	// Create a new userClient
	var userClient, _ = user.NewClient(&client.Config{
		ApiKey: apiKey,
	})

	var teamClient, _ = team.NewClient(&client.Config{
		ApiKey: apiKey,
	})

	return &Opsgenie{
		userClient: userClient,
		teamClient: teamClient,
	}
}

func (p *Opsgenie) Kind() string {
	return "opsgenie"
}

func (o *Opsgenie) LoadSchedules(ctx context.Context) error {
	panic("not implemented") // TODO: Implement
}

func (p *Opsgenie) PopulateTeamMembers(ctx context.Context, t *Team) error {
	members := []*User{}

	resp, err := p.teamClient.Get(ctx, &team.GetTeamRequest{
		IdentifierType:  team.Name,
		IdentifierValue: t.Name,
	})

	if err != nil {
		return err
	}

	for _, member := range resp.Members {
		members = append(members, &User{Resource: Resource{ID: member.User.ID}})
	}

	t.Members = members

	return nil
}

func (p *Opsgenie) PopulateTeamSchedules(ctx context.Context, team *Team) error {
	// TODO: implement
	return nil
}

func (p *Opsgenie) ListTeams(ctx context.Context) ([]*Team, error) {
	teams := []*Team{}
	opts := team.ListTeamRequest{}

	resp, err := p.teamClient.List(ctx, &opts)
	if err != nil {
		return nil, err
	}

	for _, team := range resp.Teams {
		teams = append(teams, p.toTeam(team))
	}

	return teams, nil
}

func (p *Opsgenie) toTeam(team team.ListedTeams) *Team {
	return &Team{
		// Opsgenie does not expose a slug, so generate one.
		Slug: slug.Make(team.Name),
		Resource: Resource{
			ID:   team.Id,
			Name: team.Name,
		},
	}
}

func (p *Opsgenie) ListUsers(ctx context.Context) ([]*User, error) {
	users := []*User{}
	opts := user.ListRequest{}

	for {
		resp, err := p.userClient.List(ctx, &opts)
		if err != nil {
			return nil, err
		}

		for _, user := range resp.Users {
			users = append(users, p.toUser(user))
		}

		// Results are paginated, so break if we're on the last page.
		if resp.Paging.Next == "" {
			break
		}
		opts.Offset += len(resp.Users)
	}
	return users, nil
}

func (p *Opsgenie) toUser(user user.User) *User {
	return &User{
		Email: user.Username,
		Resource: Resource{
			ID:   user.Id,
			Name: user.FullName,
		},
	}
}
