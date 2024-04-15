package store

import (
	"strings"

	"github.com/gosimple/slug"
)

func (r *ListExtTeamsRow) ExtTeam() *ExtTeam {
	return &ExtTeam{
		ID:   r.ID,
		Name: r.Name,
		Slug: r.Slug,
	}
}

func (r *ListExtTeamsRow) FhTeam() *FhTeam {
	var id, name, slug string
	if r.FhTeamID.Valid {
		id = r.FhTeamID.String
	}
	if r.FhTeamName.Valid {
		name = r.FhTeamName.String
	}
	if r.FhTeamSlug.Valid {
		slug = r.FhTeamSlug.String
	}
	return &FhTeam{
		ID:   id,
		Name: name,
		Slug: slug,
	}
}

func (r *ListExtTeamsRow) ValidName() string {
	if r.FhTeamID.Valid {
		return r.FhTeamName.String
	}
	return r.Name
}

func (r *ListExtTeamsRow) TFSlug() string {
	s := r.Slug
	if r.FhTeamID.Valid {
		s = r.FhTeamSlug.String
	}
	if s == "" {
		s = slug.Make(r.ValidName())
	}
	return strings.ReplaceAll(s, "-", "_")
}

func (r *ListFhTeamsByExtScheduleIDRow) ExtTeam() *ExtTeam {
	return &ExtTeam{
		ID:   r.ID,
		Name: r.Name,
		Slug: r.Slug,
	}
}

func (r *ListFhTeamsByExtScheduleIDRow) FhTeam() *FhTeam {
	var id, name, slug string
	if r.FhTeamID.Valid {
		id = r.FhTeamID.String
	}
	if r.FhTeamName.Valid {
		name = r.FhTeamName.String
	}
	if r.FhTeamSlug.Valid {
		slug = r.FhTeamSlug.String
	}
	return &FhTeam{
		ID:   id,
		Name: name,
		Slug: slug,
	}
}

func (r *ListFhTeamsByExtScheduleIDRow) ValidName() string {
	if r.FhTeamID.Valid {
		return r.FhTeamName.String
	}
	return r.Name
}

func (r *ListFhTeamsByExtScheduleIDRow) TFSlug() string {
	s := r.Slug
	if r.FhTeamID.Valid {
		s = r.FhTeamSlug.String
	}
	if s == "" {
		s = slug.Make(r.ValidName())
	}
	return strings.ReplaceAll(s, "-", "_")
}

func (s *ExtSchedule) TFSlug() string {
	return strings.ReplaceAll(slug.Make(s.Name), "-", "_")
}

// Technically this should never be used, but available as a fallback.
func (t *ExtTeam) TFSlug() string {
	if t.Slug == "" {
		return strings.ReplaceAll(slug.Make(t.Name), "-", "_")
	}
	return strings.ReplaceAll(t.Slug, "-", "_")
}

func (t *FhTeam) TFSlug() string {
	if t.Slug == "" {
		return strings.ReplaceAll(slug.Make(t.Name), "-", "_")
	}
	return strings.ReplaceAll(t.Slug, "-", "_")
}

func (u *FhUser) TFSlug() string {
	username := strings.Split(u.Email, "@")[0]
	return strings.ReplaceAll(slug.Make(username), "-", "_")
}
