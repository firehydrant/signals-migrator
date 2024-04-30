package pager

import (
	"context"
	"errors"

	"github.com/firehydrant/signals-migrator/store"
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

	Teams(context.Context) ([]store.ExtTeam, error)
}
