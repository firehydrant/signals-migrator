package types

import (
	"fmt"
	"time"

	"github.com/gosimple/slug"
)

type Organization struct {
	ID    string
	Teams []*Team
	Users []*User
}

func NewOrganization() Organization {
	return Organization{}
}

type Team struct {
	Slug      string
	Schedules []*Schedule
	Members   []*User
	Resource
}

func (t Team) FilterValue() string {
	return t.Name
}

func (t Team) Title() string {
	return t.Name
}

func (t Team) Description() string {
	return t.ID
}

type User struct {
	Email string
	Resource
}

func (u User) ToResource() string {
	return slug.Make(u.Email)
}

func (u User) FilterValue() string {
	return u.Email
}

func (u User) Title() string {
	return fmt.Sprintf("%s (%s)", u.Name, u.Email)
}

func (u User) Description() string {
	return u.ID
}

type ScheduleStrategy int

//go:generate stringer -type=ScheduleStrategy
const (
	Weekly ScheduleStrategy = iota
	Daily
)

type Schedule struct {
	TimeZone    string
	Strategy    ScheduleStrategy
	HandoffTime time.Time
	HandoffDay  time.Weekday
	Source      interface{}
	Resource
}

type Resource struct {
	ID          string
	RemoteID    string
	Name        string
	Description string
}

func (r *Resource) ToResource() string {
	return slug.Make(r.Name)
}
