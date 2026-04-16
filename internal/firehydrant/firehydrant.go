package firehydrant

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	fhsdk "github.com/firehydrant/firehydrant-go-sdk"
	"github.com/firehydrant/firehydrant-go-sdk/models/components"
	"github.com/firehydrant/firehydrant-go-sdk/models/operations"
	"github.com/firehydrant/signals-migrator/console"
	"github.com/firehydrant/signals-migrator/store"
	"github.com/firehydrant/terraform-provider-firehydrant/firehydrant"
)

// firehyrant.Client is technically a kind of Pager, but it does not necessarily
// satisfy the Pager interface, since that's not what we're using it for.
type Client struct {
	client *firehydrant.APIClient
	sdk    *fhsdk.FireHydrant

	apiKey string
	apiURL string
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
	sdk := fhsdk.New(
		fhsdk.WithServerURL(strings.TrimSuffix(apiURL, "v1/")),
		fhsdk.WithSecurity(components.Security{
			APIKey: apiKey,
		}),
	)
	return &Client{
		client: client,
		sdk:    sdk,
		apiKey: apiKey,
		apiURL: apiURL,
	}, nil
}

func (c *Client) ListTeams(ctx context.Context) ([]store.FhTeam, error) {
	q := store.UseQueries(ctx)
	stored, err := q.ListFhTeams(ctx)
	if err == nil && len(stored) > 0 {
		return stored, nil
	}

	// Dedupe by ID across pages: laddertruck's TeamReader sorts by name only,
	// which is unstable when teams share a name and can return the same row on
	// multiple pages.
	teams := []store.FhTeam{}
	seen := map[string]bool{}
	page := 1
	for {
		resp, err := c.sdk.Teams.ListTeams(ctx, operations.ListTeamsRequest{Page: &page})
		if err != nil {
			return nil, fmt.Errorf("fetching teams from FireHydrant: %w", err)
		}
		for _, t := range resp.GetData() {
			id := *t.GetID()
			if seen[id] {
				continue
			}
			seen[id] = true
			team := store.FhTeam{
				ID:   id,
				Name: *t.GetName(),
				Slug: *t.GetSlug(),
			}

			if err := q.InsertFhTeam(ctx, store.InsertFhTeamParams(team)); err != nil {
				return nil, fmt.Errorf("storing teams to database: %w", err)
			}
			teams = append(teams, team)
		}
		pg := resp.GetPagination()
		if pg == nil || pg.GetNext() == nil || *pg.GetNext() == 0 {
			break
		}
		page = *pg.GetNext()
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

	// Dedupe by ID across pages: laddertruck's UserReader sorts by name only,
	// which is unstable when users share a name and can return the same row on
	// multiple pages.
	users := []store.FhUser{}
	seen := map[string]bool{}
	opts := firehydrant.GetUserParams{}
	for {
		resp, err := c.client.GetUsers(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("fetching users from FireHydrant: %w", err)
		}
		for _, u := range resp.Users {
			if seen[u.ID] {
				continue
			}
			seen[u.ID] = true
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

func (c *Client) fetchUser(ctx context.Context, email string) (*store.FhUser, error) {
	opts := firehydrant.GetUserParams{Query: email}
	resp, err := c.client.GetUsers(ctx, opts)
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
