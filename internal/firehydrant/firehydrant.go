package firehydrant

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/firehydrant/signals-migrator/store"
	"github.com/firehydrant/terraform-provider-firehydrant/firehydrant"
)

// firehyrant.Client is technically a kind of Pager, but it does not necessarily
// satisfy the Pager interface, since that's not what we're using it for.
type Client struct {
	client firehydrant.Client
}

func NewClient(apiKey string, apiURL string) (*Client, error) {
	client, err := firehydrant.NewRestClient(
		apiKey,
		firehydrant.WithBaseURL(apiURL),
		firehydrant.WithUserAgentSuffix("signals-migrator"),
	)
	if err != nil {
		return nil, fmt.Errorf("initializing FireHydrant client: %w", err)
	}
	return &Client{client: client}, nil
}

func (c *Client) ListTeams(ctx context.Context) ([]store.FhTeam, error) {
	q := store.UseQueries(ctx)
	stored, err := q.ListFhTeams(ctx)
	if err == nil && len(stored) > 0 {
		return stored, nil
	}

	teams := []store.FhTeam{}
	opts := &firehydrant.TeamQuery{}
	for {
		resp, err := c.client.Teams().List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("fetching teams from FireHydrant: %w", err)
		}
		for _, t := range resp.Teams {
			teams = append(teams, store.FhTeam{
				ID:   t.ID,
				Name: t.Name,
				Slug: t.Slug,
			})
		}
		if resp.Pagination.Next == 0 || resp.Pagination.Page >= resp.Pagination.Last {
			break
		}
		opts.Page = resp.Pagination.Next
	}

	for _, t := range teams {
		if err := q.InsertFhTeam(ctx, store.InsertFhTeamParams(t)); err != nil {
			return nil, fmt.Errorf("storing teams to database: %w", err)
		}
	}
	return teams, nil
}

// ListUsers retrieves all users from within a FireHydrant organization, based on
// the provided API key access.
func (c *Client) ListUsers(ctx context.Context) ([]store.FhUser, error) {
	q := store.UseQueries(ctx)
	stored, err := q.ListFhUsers(ctx)
	if err == nil && len(stored) > 0 {
		return stored, nil
	}

	users := []store.FhUser{}
	opts := firehydrant.GetUserParams{}
	for {
		resp, err := c.client.GetUsers(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("fetching users from FireHydrant: %w", err)
		}
		for _, u := range resp.Users {
			users = append(users, store.FhUser{
				ID:    u.ID,
				Name:  u.Name,
				Email: u.Email,
			})
		}
		if resp.Pagination.Next == 0 {
			break
		}
		opts.Page = resp.Pagination.Next
	}

	for _, u := range users {
		if err := q.InsertFhUser(ctx, store.InsertFhUserParams(u)); err != nil {
			return nil, fmt.Errorf("storing users to database: %w", err)
		}
	}
	return users, nil
}

// MatchUsers attempts to pair users in the parameter with its FireHydrant User counterpart.
// Returns: a list of users which were not successfully matched.
func (c *Client) MatchUsers(ctx context.Context) error {
	q := store.UseQueries(ctx)

	// Calling ListUsers just to make sure DB store exists.
	_, err := c.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("fetching FireHydrant users: %w", err)
	}

	users, err := q.ListUsersJoinByEmail(ctx)
	if err != nil {
		return fmt.Errorf("listing external users: %w", err)
	}

	for _, user := range users {
		if user.FhUser.ID != "" {
			if err := c.PairUsers(ctx, user.FhUser.ID, user.ExtUser.ID); err != nil {
				return fmt.Errorf("pairing users: %w", err)
			}
		}
	}

	return nil
}

func (c *Client) PairUsers(ctx context.Context, fhUserID string, extUserID string) error {
	return store.UseQueries(ctx).LinkExtUser(ctx, store.LinkExtUserParams{
		FhUserID: sql.NullString{Valid: true, String: fhUserID},
		ID:       extUserID,
	})
}
