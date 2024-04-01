package pager

import (
	"context"

	"github.com/gosimple/slug"
	"github.com/victorops/go-victorops/victorops"
)

type VictorOps struct {
	client *victorops.Client
}

func NewVictorOps(apiKey string, appId string) *VictorOps {
	return &VictorOps{
		client: victorops.NewClient(appId, apiKey, "https://api.victorops.com"),
	}
}

func (v *VictorOps) Kind() string {
	return "victorops"
}

func (v *VictorOps) PopulateTeamMembers(ctx context.Context, team *Team) error {
	members := []*User{}

	vmembers, _, err := v.client.GetTeamMembers(team.ID)
	if err != nil {
		return err
	}

	for _, member := range vmembers.Members {
		members = append(members, &User{Resource: Resource{ID: member.Username}})
	}

	team.Members = members
	return nil
}

func (v *VictorOps) PopulateTeamSchedules(ctx context.Context, team *Team) error {
	// TODO: implement
	return nil
}

func (v *VictorOps) ListTeams(ctx context.Context) ([]*Team, error) {
	teams := []*Team{}

	vteams, _, err := v.client.GetAllTeams()
	if err != nil {
		return nil, err
	}

	for _, team := range *vteams {
		teams = append(teams, v.toTeam(team))
	}

	return teams, nil
}

func (v *VictorOps) toTeam(team victorops.Team) *Team {
	return &Team{
		// PagerDuty does not expose a slug, so generate one.
		Slug: slug.Make(team.Name),
		Resource: Resource{
			ID:   team.Slug,
			Name: team.Name,
		},
	}
}

func (v *VictorOps) ListUsers(ctx context.Context) ([]*User, error) {
	users := []*User{}

	vusers, _, err := v.client.GetAllUsers()
	if err != nil {
		return nil, err
	}

	for _, userSlice := range vusers.Users {
		for _, user := range userSlice {
			users = append(users, v.toUser(user))
		}
	}

	return users, nil
}

func (v *VictorOps) toUser(user victorops.User) *User {
	return &User{
		Email: user.Email,
		Resource: Resource{
			ID:   user.Username,
			Name: user.FirstName + " " + user.LastName,
		},
	}
}
