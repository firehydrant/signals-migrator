//go:build !demo

package pager

import (
	"context"
	"fmt"
	"strings"
)

func NewPager(_ context.Context, kind string, apiKey string, appId string) (Pager, error) {
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
