package pager

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrUnknownProvider = errors.New("unknown pager provider")
	ErrNoResults       = errors.New("no results found")
)

type Pager interface {
	ListUsers(ctx context.Context) ([]*User, error)
	ListTeams(ctx context.Context) ([]*Team, error)

	LoadSchedules(ctx context.Context) error

	PopulateTeamMembers(ctx context.Context, team *Team) error
	PopulateTeamSchedules(ctx context.Context, team *Team) error
}

func NewPager(kind string, apiKey string, appId string) (Pager, error) {
	switch strings.ToLower(kind) {
	case "pagerduty":
		return NewPagerDuty(apiKey), nil
	case "victorops":
		return NewVictorOps(apiKey, appId), nil
	case "opsgenie":
		return NewOpsgenie(apiKey), nil
	}

	return nil, fmt.Errorf("%w '%s'", ErrUnknownProvider, kind)
}
