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
