package pager

import (
	"context"
	"fmt"

	"github.com/firehydrant/signals-migrator/console"
	"github.com/firehydrant/signals-migrator/store"
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
	vusers, _, err := v.client.GetAllUserV2()
	if err != nil {
		return fmt.Errorf("querying victorops: %w", err)
	}

	for _, user := range vusers.Users {
		if err := store.UseQueries(ctx).InsertExtUser(ctx, store.InsertExtUserParams{
			ID:    user.Username,
			Email: user.Email,
			Name:  user.FirstName + " " + user.LastName,
		}); err != nil {
			return fmt.Errorf("saving user to db: %w", err)
		}
	}

	return nil
}

func (v *VictorOps) LoadTeams(ctx context.Context) error {
	vteams, _, err := v.client.GetAllTeams()
	if err != nil {
		return fmt.Errorf("querying victorops: %w", err)
	}

	for _, team := range *vteams {
		tSlug := slug.Make(team.Name)
		if err := store.UseQueries(ctx).InsertExtTeam(ctx, store.InsertExtTeamParams{
			ID:   tSlug,
			Slug: tSlug,
			Name: team.Name,
		}); err != nil {
			return fmt.Errorf("saving team to db: %w", err)
		}
	}
	return nil
}

func (v *VictorOps) LoadTeamMembers(ctx context.Context) error {
	console.Warnf("victorops.LoadTeamMembers is not currently supported.")
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
