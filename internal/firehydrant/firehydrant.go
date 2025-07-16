package firehydrant

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	firehydrantgosdk "github.com/firehydrant/firehydrant-go-sdk"
	"github.com/firehydrant/firehydrant-go-sdk/models/components"
	"github.com/firehydrant/firehydrant-go-sdk/models/operations"
	"github.com/firehydrant/signals-migrator/console"
	"github.com/firehydrant/signals-migrator/store"
	"github.com/firehydrant/terraform-provider-firehydrant/firehydrant"
)

// firehyrant.Client is technically a kind of Pager, but it does not necessarily
// satisfy the Pager interface, since that's not what we're using it for.
type Client struct {
	legacyClient firehydrant.Client
	client       *firehydrantgosdk.FireHydrant

	apiKey string
	apiURL string
}

func NewClient(apiKey string, apiURL string) (*Client, error) {
	// Initialize legacy client for SCIM
	legacyClient, err := firehydrant.NewRestClient(
		apiKey,
		firehydrant.WithBaseURL(apiURL),
		firehydrant.WithUserAgentSuffix("signals-migrator"),
	)
	if err != nil {
		return nil, fmt.Errorf("initializing legacy FireHydrant client: %w", err)
	}

	// Initialize Go SDK client
	client := firehydrantgosdk.New(
		firehydrantgosdk.WithSecurity(components.Security{
			APIKey: apiKey,
		}),
	)

	return &Client{
		legacyClient: legacyClient,
		client:       client,
		apiKey:       apiKey,
		apiURL:       apiURL,
	}, nil
}

func (c *Client) ListTeams(ctx context.Context) ([]store.FhTeam, error) {
	q := store.UseQueries(ctx)
	stored, err := q.ListFhTeams(ctx)
	if err == nil && len(stored) > 0 {
		return stored, nil
	}

	teams := []store.FhTeam{}
	page := 1
	request := operations.ListTeamsRequest{
		Page: &page,
	}

	for {
		resp, err := c.client.Teams.ListTeams(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("fetching teams from FireHydrant: %w", err)
		}
		for _, t := range resp.Data {
			if t.ID == nil || t.Name == nil || t.Slug == nil {
				console.Warnf("skipping team with nil fields")
				continue
			}
			team := store.FhTeam{
				ID:   *t.ID,
				Name: *t.Name,
				Slug: *t.Slug,
			}

			if err := q.InsertFhTeam(ctx, store.InsertFhTeamParams(team)); err != nil {
				return nil, fmt.Errorf("storing teams to database: %w", err)
			}
			teams = append(teams, team)

		}

		if resp.Pagination.Next == nil || *resp.Pagination.Next == 0 {
			break
		}
		page = *resp.Pagination.Next

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
	page := 1
	for {

		resp, err := c.client.Users.ListUsers(ctx, &page, nil, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("fetching users from FireHydrant: %w", err)
		}
		for _, u := range resp.Data {
			users = append(users, store.FhUser{
				ID:    *u.ID,
				Name:  *u.Name,
				Email: *u.Email,
			})
		}

		next := resp.Pagination.Next
		if *next == 0 {
			break
		}
		page = *resp.Pagination.Next
	}

	for _, u := range users {
		if err := q.InsertFhUser(ctx, store.InsertFhUserParams(u)); err != nil {
			return nil, fmt.Errorf("storing users to database: %w", err)
		}
	}
	return users, nil
}

func (c *Client) fetchUser(ctx context.Context, email string) (*store.FhUser, error) {
	opts := firehydrant.GetUserParams{Query: email}
	resp, err := c.legacyClient.GetUsers(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("fetching users from FireHydrant: %w", err)
	}
	if len(resp.Users) == 0 {
		return nil, fmt.Errorf("fetching users from FireHydrant: no users found")
	}
	if len(resp.Users) > 1 {
		console.Warnf("multiple users found for email %s, selecting first one with ID: '%s'", email, resp.Users[0].ID)
	}
	user := store.FhUser{
		ID:    resp.Users[0].ID,
		Name:  resp.Users[0].Name,
		Email: resp.Users[0].Email,
	}
	if err := store.UseQueries(ctx).InsertFhUser(ctx, store.InsertFhUserParams(user)); err != nil {
		return nil, fmt.Errorf("storing user '%s' to database: %w", user.Email, err)
	}

	return &user, nil
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

type SCIMUser interface {
	Username() string
	FamilyName() string
	GivenName() string
	PrimaryEmail() string
}

var (
	_ SCIMUser = &store.ExtUser{}
)

// CreateUser provisions user via SCIM. Terraform client does not support this, therefore
// we are making the barebones request directly.
func (c *Client) CreateUser(ctx context.Context, u SCIMUser) (*store.FhUser, error) {
	payload := map[string]any{
		"userName": u.Username(),
		"name": map[string]any{
			"familyName": u.FamilyName(),
			"givenName":  u.GivenName(),
		},
		"emails": []map[string]any{
			{
				"value":   u.PrimaryEmail(),
				"primary": true,
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("converting user payload to JSON: %w", err)
	}
	buf := bytes.NewBuffer(body)
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/scim/v2/Users", c.apiURL), buf)
	if err != nil {
		return nil, fmt.Errorf("composing request to create user: %w", err)
	}
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		console.Errorf("unexpected status code %d: %s\n", resp.StatusCode, body)
		return nil, fmt.Errorf("creating user: unexpected status code %d", resp.StatusCode)
	}
	return c.fetchUser(ctx, u.PrimaryEmail())
}

func (c *Client) CreateUsers(ctx context.Context, users []SCIMUser) ([]store.FhUser, error) {
	created := []store.FhUser{}
	for _, u := range users {
		user, err := c.CreateUser(ctx, u)
		if err != nil {
			return nil, fmt.Errorf("creating user: %w", err)
		}
		created = append(created, *user)
	}
	return created, nil
}
