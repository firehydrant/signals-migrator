package fetcher

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/firehydrant/signals-migrator/selector"
	"github.com/firehydrant/signals-migrator/types"
	"github.com/firehydrant/terraform-provider-firehydrant/firehydrant"
)

type FireHydrant struct {
	client firehydrant.Client
	teams  []*types.Team
	users  []*types.User
}

func NewFireHydrant(apiKey string, apiUrl string) (FireHydrant, error) {
	client, err := firehydrant.NewRestClient(apiKey, firehydrant.WithBaseURL(apiUrl))
	if err != nil {
		return FireHydrant{}, err
	}

	return FireHydrant{
		client: client,
	}, nil
}

// Takes a team provided by a third-party provider and allows for mapping it to an existing FireHydrant team
func (f *FireHydrant) ReconcileTeam(ctx context.Context, t *types.Team) (*types.Team, error) {
	teams, err := f.Teams(ctx)

	if err != nil {
		return nil, err
	}

	// TODO: find better way to pass slice of teams to selector
	list := make([]list.Item, len(teams))
	for i := range teams {
		list[i] = teams[i]
	}

	selTeam, err := selector.Selector(list, fmt.Sprintf("Select a FireHydrant team to map %s to", t.Name))
	if err != nil {
		return nil, fmt.Errorf("unable to select team for mapping: %w", err)
	}

	// I hate this, but bubbletea requires an array of items implementing list.Item, and the
	// type of the selected entry is *list.Item
	if tm, ok := (*selTeam).(*types.Team); ok {
		t.ID = tm.ID
		t.Name = tm.Name
		return t, nil
	}

	return nil, errors.New("could not use selected team")
}

// Takes a user provided by a third-party provider and allows for mapping it to an existing FireHydrant user
func (f *FireHydrant) ReconcileUser(ctx context.Context, u *types.User) (*types.User, error) {
	users, err := f.Users(ctx)

	if err != nil {
		return nil, err
	}

	list := make([]list.Item, len(users))
	for i := range users {
		list[i] = users[i]
	}

	selUser, err := selector.Selector(list, fmt.Sprintf("Select a FireHydrant user to map %s (%s) to", u.Name, u.Email))
	if err != nil {
		return nil, fmt.Errorf("unable to select user for mapping: %w", err)
	}

	if m, ok := (*selUser).(*types.User); ok {
		u.RemoteID = m.RemoteID
		u.Email = m.Email
		return u, nil
	}

	return nil, errors.New("could not use selected user")
}

func (f *FireHydrant) Users(ctx context.Context) ([]*types.User, error) {
	if len(f.users) > 0 {
		return f.users, nil
	}

	o := firehydrant.GetUserParams{}
	var users []*types.User

	// TODO: add pagination to GetUsers
	resp, err := f.client.GetUsers(ctx, o)
	if err != nil {
		return nil, err
	}

	for _, user := range resp.Users {
		users = append(users, &types.User{
			Email: user.Email,
			Resource: types.Resource{
				Name: user.Name,
				ID:   user.ID,
			},
		})
	}

	f.users = users
	return users, nil
}

func (f *FireHydrant) User(ctx context.Context, email string) (*types.User, error) {
	// TODO: add a get user by email method to the terraform FH api client
	resp, err := f.client.GetUsers(ctx, firehydrant.GetUserParams{Query: email})
	if err != nil {
		return nil, err
	}

	if len(resp.Users) == 0 {
		return nil, errors.New("user not found")
	}

	// TODO: memoize user fetch
	return &types.User{
		Email: resp.Users[0].Email,
		Resource: types.Resource{
			ID:   resp.Users[0].ID,
			Name: resp.Users[0].Name,
		},
	}, nil
}

func (f *FireHydrant) Team(ctx context.Context, name string) (*types.Team, error) {
	resp, err := f.client.Teams().Get(ctx, name)
	if err != nil {
		return nil, err
	}

	// TODO: memoize team fetch
	return &types.Team{
		Resource: types.Resource{
			Name: resp.Name,
			ID:   resp.ID,
		},
	}, nil
}

func (f *FireHydrant) Teams(ctx context.Context) ([]*types.Team, error) {
	if len(f.teams) > 0 {
		return f.teams, nil
	}

	resp, err := f.client.Teams().List(ctx, &firehydrant.TeamQuery{})
	if err != nil {
		return []*types.Team{}, err
	}

	var teams []*types.Team

	for _, team := range resp.Teams {
		teams = append(teams, &types.Team{
			Slug: team.Slug,
			Resource: types.Resource{
				ID:   team.ID,
				Name: team.Name,
			},
		})
	}

	f.teams = teams

	return teams, nil
}
