package pager

// TODO: [noted with apologies]
//   This file existed before the pivot to using SQLite. Now that we're using
//   SQLite, we can instead make use of store/models.go instead of this file.

import (
	"fmt"

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
	Slug    string
	Members []*User
	Resource
}

func (t *Team) String() string {
	return fmt.Sprintf("%s %s (%s)", t.Resource.ID, t.Resource.Name, t.Slug)
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
