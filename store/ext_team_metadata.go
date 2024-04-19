package store

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// ExtTeamMetadata is internal metadata for plumbing work. This is an "escape hatch" for any data which
// is needed for transformation of models in specific provider scenarios (e.g. PagerDuty) which would
// otherwise be unnecessary for the core functionality of how we think about teams and its relationships.
type ExtTeamMetadata struct {
	// PagerDutyProxyFor denotes the downstream teams which are proxied by this team. When we import a PagerDuty service
	// as a team, we still need information of which PagerDuty teams are contained within this service.
	//
	// For example, if a PagerDuty service "Kubernetes" is a proxy for PagerDuty teams "SRE" and "Platform",
	// then we will have the following entries in the store:
	// - ExtTeam{ID: "Kubernetes"}
	// - ExtTeam{ID: "SRE", Metadata: ExtTeamMetadata{ProxyFor: "Kubernetes"}}
	// - ExtTeam{ID: "Platform", Metadata: ExtTeamMetadata{ProxyFor: "Kubernetes"}}
	// As such, we have the references to consolidate the internal teams into the PagerDuty service in ExtTeam.
	//
	// Corollary to that, when PagerDuty responds with information of members for "SRE" and "Platform" teams,
	// we store the members as members of the "Kubernetes" service.
	// As a result, when we render Terraform for FireHydrant team, all the members of PagerDuty "SRE" and "Platform" teams
	// will be a member of the FireHydrant "Kubernetes" team.
	//
	// In the end, the "SRE" and "Platform" teams are not going to be imported to FireHydrant, but the members
	// will be merged into the "Kubernetes" team.
	PagerDutyProxyFor string `json:"pdProxyFor,omitempty"`
}

func (e *ExtTeamMetadata) Value() (driver.Value, error) {
	v, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

func (e *ExtTeamMetadata) Scan(src any) error {
	str, ok := src.(string)
	if ok {
		if str == "" {
			return nil
		}
		return json.Unmarshal([]byte(str), e)
	}
	return fmt.Errorf("could not scan as ExtTeamMetadata: %v", src)
}
