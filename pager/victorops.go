package pager

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/firehydrant/signals-migrator/console"
	"github.com/firehydrant/signals-migrator/store"
	"github.com/victorops/go-victorops/victorops"
)

type VictorOps struct {
	client *victorops.Client
}

func NewVictorOps(apiKey string, appId string) *VictorOps {
	return NewVictorOpsWithURL(apiKey, appId, "https://api.victorops.com")
}

func NewVictorOpsWithURL(apiKey string, appId string, url string) *VictorOps {
	return &VictorOps{
		client: victorops.NewClient(appId, apiKey, url),
	}
}

func (v *VictorOps) Kind() string {
	return "VictorOps"
}

func (v *VictorOps) TeamInterfaces() []string {
	return []string{"team"}
}

func (v *VictorOps) UseTeamInterface(string) error {
	return nil
}

func (v *VictorOps) Teams(ctx context.Context) ([]store.ExtTeam, error) {
	return store.UseQueries(ctx).ListExtTeams(ctx)
}

func (v *VictorOps) LoadUsers(ctx context.Context) error {
	type User struct {
		FirstName string `json:"firstName,omitempty"`
		LastName  string `json:"lastName,omitempty"`
		Username  string `json:"username,omitempty"`
		Email     string `json:"email,omitempty"`
		Admin     bool   `json:"admin,omitempty"`
		SelfURL   string `json:"_selfUrl,omitempty"`
	}
	type Users struct {
		Users []User `json:"users,omitempty"`
	}

	_, details, err := v.client.GetAllUserV2()
	if err != nil {
		return fmt.Errorf("querying victorops: %w", err)
	}
	var users Users
	if err := json.Unmarshal([]byte(details.ResponseBody), &users); err != nil {
		return fmt.Errorf("unmarshalling victorops users: %w", err)
	}

	for _, user := range users.Users {
		annotations := user.SelfURL
		if user.Admin {
			annotations += " [Admin]"
		}
		if err := store.UseQueries(ctx).InsertExtUser(ctx, store.InsertExtUserParams{
			ID:    user.Username,
			Email: user.Email,
			Name:  user.FirstName + " " + user.LastName,

			Annotations: annotations,
		}); err != nil {
			return fmt.Errorf("saving user to db: %w", err)
		}
	}

	return nil
}

func (v *VictorOps) LoadTeams(ctx context.Context) error {
	type Team struct {
		Name          string `json:"name,omitempty"`
		Slug          string `json:"slug,omitempty"`
		IsDefaultTeam bool   `json:"isDefaultTeam,omitempty"`
		SelfURL       string `json:"_selfUrl,omitempty"`
	}

	_, details, err := v.client.GetAllTeams()
	if err != nil {
		return fmt.Errorf("querying victorops: %w", err)
	}
	var vteams []Team
	if err := json.Unmarshal([]byte(details.ResponseBody), &vteams); err != nil {
		return fmt.Errorf("unmarshalling victorops teams: %w", err)
	}

	for _, team := range vteams {
		annotations := team.SelfURL
		if team.IsDefaultTeam {
			annotations += " [Default Team]"
		}
		if err := store.UseQueries(ctx).InsertExtTeam(ctx, store.InsertExtTeamParams{
			ID:   team.Slug,
			Slug: team.Slug,
			Name: team.Name,

			Annotations: annotations,
		}); err != nil {
			return fmt.Errorf("saving team to db: %w", err)
		}
	}
	return nil
}

func (v *VictorOps) LoadTeamMembers(ctx context.Context) error {
	q := store.UseQueries(ctx)
	teams, err := q.ListTeams(ctx)
	if err != nil {
		return fmt.Errorf("loading teams: %w", err)
	}
	for _, team := range teams {
		vteam, _, err := v.client.GetTeamMembers(team.ID)
		if err != nil {
			return fmt.Errorf("querying victorops: %w", err)
		}
		for _, member := range vteam.Members {
			if err := q.InsertExtMembership(ctx, store.InsertExtMembershipParams{
				TeamID: team.ID,
				UserID: member.Username,
			}); err != nil {
				return fmt.Errorf("saving team member to db: %w", err)
			}
		}
	}
	return nil
}

func (v *VictorOps) LoadSchedules(ctx context.Context) error {
	// TODO: implement
	console.Warnf("victorops.LoadSchedules is not currently supported.")
	return nil
}

func (v *VictorOps) LoadEscalationPolicies(ctx context.Context) error {
	// TODO: implement
	console.Warnf("victorops.LoadEscalationPolicies is not currently supported.")
	return nil
}
