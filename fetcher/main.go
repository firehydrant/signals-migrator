package fetcher

import (
	"context"

	"github.com/firehydrant/signals-migrator/types"
)

// Ideally only the generic types in types/ should be returned by alerting provider
// fetchers, leaving the specific types (pagerduty.Team, etc) buried in the fetcher.
type Fetcher interface {
	Teams(context.Context) ([]*types.Team, error)
	Users(context.Context) ([]*types.User, error)
	User(context.Context, string) (*types.User, error)
	Team(context.Context, string) (*types.Team, error)
	TeamSchedules(context.Context, string) ([]*types.Schedule, error)
}
