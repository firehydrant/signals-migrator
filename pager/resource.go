package pager

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
)

// Resource refers to identification properties within Terraform.
// It is meant to be embedded in other types.
type Resource struct {
	ID          string
	Name        string
	Description string
}

// Organization refers to a collection of teams and users within a Pager service.
// We map migrations from one organization in a given Pager service to FireHydrant.
type Organization struct {
	ID    string
	Teams []*Team
	Users []*User
	Resource
}

func (o *Organization) String() string {
	return o.Resource.Name
}

func (o *Organization) Slug() string {
	return slug.Make(o.Name)
}

// Team refers to a team within a Pager service. This is a concept which
// FireHydrant Signals use to group users and their schedules, but other
// providers may label them differently, e.g. Service Owners / User Groups.
type Team struct {
	Slug      string
	Schedules []*Schedule
	Members   []*User
	Resource
}

func (t *Team) String() string {
	return fmt.Sprintf("%s %s (%s)", t.Resource.ID, t.Resource.Name, t.Slug)
}

func (t *Team) TFSlug() string {
	return strings.ReplaceAll(t.Slug, "-", "_")
}

// User refers to a user within a Pager service, may also be referred
// to as member or contact.
type User struct {
	Email string
	Resource
}

func (u *User) String() string {
	name := u.Resource.Name
	email := u.Email
	if name != "" && email != "" {
		return fmt.Sprintf("%s (%s)", u.Name, u.Email)
	}
	if name != "" {
		return name
	}
	return email
}

func (u *User) Slug() string {
	return slug.Make(u.Email)
}

func (u *User) TFSlug() string {
	return strings.ReplaceAll(u.Slug(), "-", "_")
}

// ScheduleStrategy refers to the strategy used to determine the scheduling
// order of users within a team.
type ScheduleStrategy int

//go:generate stringer -type=ScheduleStrategy
const (
	Daily ScheduleStrategy = iota
	Weekly
	Fortnightly
)

// Schedule refers to collection of on-call shifts within a team. It dictates
// the parameters on how shifts are scheduled automatically by pager provider.
type Schedule struct {
	Strategy    ScheduleStrategy
	TimeZone    string
	HandoffTime time.Time
	HandoffDay  time.Weekday
}
