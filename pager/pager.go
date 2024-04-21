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
	Kind() string
	TeamInterfaces() []string
	UseTeamInterface(interfaceName string) error

	LoadUsers(ctx context.Context) error
	LoadTeams(ctx context.Context) error
	LoadTeamMembers(ctx context.Context) error
	LoadSchedules(ctx context.Context) error
	LoadEscalationPolicies(ctx context.Context) error

	// Deprecated: use LoadUsers instead
	ListUsers(ctx context.Context) ([]*User, error)
	// Deprecated: use LoadTeams instead
	ListTeams(ctx context.Context) ([]*Team, error)
	// Deprecated: use LoadTeamMembers instead
	PopulateTeamMembers(ctx context.Context, team *Team) error
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
