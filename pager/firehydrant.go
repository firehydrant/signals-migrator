package pager

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/firehydrant/signals-migrator/store"
	"github.com/firehydrant/terraform-provider-firehydrant/firehydrant"
)

// FireHydrant is technically a kind of Pager, but it does not necessarily
// satisfy the Pager interface, since that's not what we're using it for.
type FireHydrant struct {
	client firehydrant.Client
}

func NewFireHydrant(apiKey string, apiURL string) (*FireHydrant, error) {
	client, err := firehydrant.NewRestClient(
		apiKey,
		firehydrant.WithBaseURL(apiURL),
		firehydrant.WithUserAgentSuffix("signals-migrator"),
	)
	if err != nil {
		return nil, fmt.Errorf("initializing FireHydrant client: %w", err)
	}
	return &FireHydrant{client: client}, nil
}

func (f *FireHydrant) ListTeams(ctx context.Context) ([]*Team, error) {
	teams := []*Team{}
	stored, err := store.UseQueries(ctx).ListFhTeams(ctx)
	if err == nil && len(stored) > 0 {
		for _, t := range stored {
			teams = append(teams, &Team{
				Slug: t.Slug,
				Resource: Resource{
					ID:   t.ID,
					Name: t.Name,
				},
			})
		}
		return teams, nil
	}

	opts := &firehydrant.TeamQuery{}
	for {
		resp, err := f.client.Teams().List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("fetching teams from FireHydrant: %w", err)
		}
		for _, t := range resp.Teams {
			teams = append(teams, f.toTeam(t))
		}
		if resp.Pagination.Next == 0 || resp.Pagination.Page >= resp.Pagination.Last {
			break
		}
		opts.Page = resp.Pagination.Next
	}

	for _, t := range teams {
		if err := store.UseQueries(ctx).InsertFhTeam(ctx, store.InsertFhTeamParams{
			ID:   t.ID,
			Name: t.Name,
			Slug: t.Slug,
		}); err != nil {
			return nil, fmt.Errorf("storing teams to database: %w", err)
		}
	}
	return teams, nil
}

func (f *FireHydrant) toTeam(team firehydrant.TeamResponse) *Team {
	return &Team{
		Slug: team.Slug,
		Resource: Resource{
			ID:          team.ID,
			Name:        team.Name,
			Description: team.Description,
		},
	}
}

// ListUsers retrieves all users from within a FireHydrant organization, based on
// the provided API key access.
func (f *FireHydrant) ListUsers(ctx context.Context) ([]*User, error) {
	users := []*User{}
	stored, err := store.UseQueries(ctx).ListFhUsers(ctx)
	if err == nil && len(stored) > 0 {
		for _, u := range stored {
			users = append(users, &User{
				Email: u.Email,
				Resource: Resource{
					ID:   u.ID,
					Name: u.Name,
				},
			})
		}
		return users, nil
	}

	opts := firehydrant.GetUserParams{}
	for {
		resp, err := f.client.GetUsers(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("fetching users from FireHydrant: %w", err)
		}
		for _, u := range resp.Users {
			users = append(users, f.toUser(u))
		}
		if resp.Pagination.Next == 0 {
			break
		}
		opts.Page = resp.Pagination.Next
	}

	for _, u := range users {
		if err := store.UseQueries(ctx).InsertFhUser(ctx, store.InsertFhUserParams{
			ID:    u.ID,
			Email: u.Email,
			Name:  u.Name,
		}); err != nil {
			return nil, fmt.Errorf("storing users to database: %w", err)
		}
	}
	return users, nil
}

func (f *FireHydrant) toUser(user firehydrant.User) *User {
	return &User{
		Email: user.Email,
		Resource: Resource{
			ID:   user.ID,
			Name: user.Name,
		},
	}
}

// MatchUsers attempts to pair users in the parameter with its FireHydrant User counterpart.
// Returns: a list of users which were not successfully matched.
func (f *FireHydrant) MatchUsers(ctx context.Context) error {
	q := store.UseQueries(ctx)

	// Calling ListUsers just to make sure DB store exists.
	_, err := f.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("fetching FireHydrant users: %w", err)
	}

	users, err := q.ListUsersJoinByEmail(ctx)
	if err != nil {
		return fmt.Errorf("listing external users: %w", err)
	}

	for _, user := range users {
		if user.FhUser.ID != "" {
			if err := f.PairUsers(ctx, user.FhUser.ID, user.ExtUser.ID); err != nil {
				return fmt.Errorf("pairing users: %w", err)
			}
		}
	}

	return nil
}

func (f *FireHydrant) PairUsers(ctx context.Context, fhUserID string, extUserID string) error {
	return store.UseQueries(ctx).LinkExtUser(ctx, store.LinkExtUserParams{
		FhUserID: sql.NullString{Valid: true, String: fhUserID},
		ID:       extUserID,
	})
}
