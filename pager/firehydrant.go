package pager

import (
	"context"
	"fmt"
	"log"

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
	opts := &firehydrant.TeamQuery{}

	for {
		resp, err := f.client.Teams().List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("fetching teams from FireHydrant: %w", err)
		}
		log.Printf("%+v", resp.Pagination)
		for _, t := range resp.Teams {
			teams = append(teams, f.toTeam(t))
		}
		if resp.Pagination.Next == 0 || resp.Pagination.Page >= resp.Pagination.Last {
			break
		}
		opts.Page = resp.Pagination.Next
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

// TODO(wilsonehusin): cache / memoize responses
func (f *FireHydrant) FetchUsers(ctx context.Context) (map[string]*User, error) {
	users, err := f.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching users from FireHydrant: %w", err)
	}
	usersByEmail := map[string]*User{}
	for _, u := range users {
		usersByEmail[u.Email] = u
	}
	return usersByEmail, nil
}

// ListUsers retrieves all users from within a FireHydrant organization, based on
// the provided API key access.
func (f *FireHydrant) ListUsers(ctx context.Context) ([]*User, error) {
	users := []*User{}
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
// Returns:
// - mapping of user's email to their FireHydrant User counterpart.
// - a list of users which were not successfully matched.
func (f *FireHydrant) MatchUsers(ctx context.Context, users []*User) (map[string]*User, []*User, error) {
	fhUsers, err := f.FetchUsers(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching users from FireHydrant: %w", err)
	}

	matchedUsers := map[string]*User{}
	unmatchedUsers := []*User{}

	for _, user := range users {
		if u, ok := fhUsers[user.Email]; ok {
			matchedUsers[user.Email] = u
		} else {
			unmatchedUsers = append(unmatchedUsers, user)
		}
	}

	return matchedUsers, unmatchedUsers, nil
}
