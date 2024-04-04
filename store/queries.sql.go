// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: queries.sql

package store

import (
	"context"
	"database/sql"
)

const getFhUserByEmail = `-- name: GetFhUserByEmail :one
SELECT id, name, email FROM fh_users WHERE email = ?
`

func (q *Queries) GetFhUserByEmail(ctx context.Context, email string) (FhUser, error) {
	row := q.db.QueryRowContext(ctx, getFhUserByEmail, email)
	var i FhUser
	err := row.Scan(&i.ID, &i.Name, &i.Email)
	return i, err
}

const insertExtMembership = `-- name: InsertExtMembership :exec
INSERT INTO ext_memberships (user_id, team_id) VALUES (?, ?)
`

type InsertExtMembershipParams struct {
	UserID string `json:"user_id"`
	TeamID string `json:"team_id"`
}

func (q *Queries) InsertExtMembership(ctx context.Context, arg InsertExtMembershipParams) error {
	_, err := q.db.ExecContext(ctx, insertExtMembership, arg.UserID, arg.TeamID)
	return err
}

const insertExtTeam = `-- name: InsertExtTeam :exec
INSERT INTO ext_teams (id, name, slug, fh_team_id) VALUES (?, ?, ?, ?)
`

type InsertExtTeamParams struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Slug     string         `json:"slug"`
	FhTeamID sql.NullString `json:"fh_team_id"`
}

func (q *Queries) InsertExtTeam(ctx context.Context, arg InsertExtTeamParams) error {
	_, err := q.db.ExecContext(ctx, insertExtTeam,
		arg.ID,
		arg.Name,
		arg.Slug,
		arg.FhTeamID,
	)
	return err
}

const insertExtUser = `-- name: InsertExtUser :exec
INSERT INTO ext_users (id, name, email, fh_user_id) VALUES (?, ?, ?, ?)
`

type InsertExtUserParams struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Email    string         `json:"email"`
	FhUserID sql.NullString `json:"fh_user_id"`
}

func (q *Queries) InsertExtUser(ctx context.Context, arg InsertExtUserParams) error {
	_, err := q.db.ExecContext(ctx, insertExtUser,
		arg.ID,
		arg.Name,
		arg.Email,
		arg.FhUserID,
	)
	return err
}

const insertFhTeam = `-- name: InsertFhTeam :exec
INSERT INTO fh_teams (id, name, slug) VALUES (?, ?, ?)
`

type InsertFhTeamParams struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (q *Queries) InsertFhTeam(ctx context.Context, arg InsertFhTeamParams) error {
	_, err := q.db.ExecContext(ctx, insertFhTeam, arg.ID, arg.Name, arg.Slug)
	return err
}

const insertFhUser = `-- name: InsertFhUser :exec
INSERT INTO fh_users (id, name, email) VALUES (?, ?, ?)
`

type InsertFhUserParams struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (q *Queries) InsertFhUser(ctx context.Context, arg InsertFhUserParams) error {
	_, err := q.db.ExecContext(ctx, insertFhUser, arg.ID, arg.Name, arg.Email)
	return err
}

const linkExtTeam = `-- name: LinkExtTeam :exec
UPDATE ext_teams SET fh_team_id = ? WHERE id = ?
`

type LinkExtTeamParams struct {
	FhTeamID sql.NullString `json:"fh_team_id"`
	ID       string         `json:"id"`
}

func (q *Queries) LinkExtTeam(ctx context.Context, arg LinkExtTeamParams) error {
	_, err := q.db.ExecContext(ctx, linkExtTeam, arg.FhTeamID, arg.ID)
	return err
}

const linkExtUser = `-- name: LinkExtUser :exec
UPDATE ext_users SET fh_user_id = ? WHERE id = ?
`

type LinkExtUserParams struct {
	FhUserID sql.NullString `json:"fh_user_id"`
	ID       string         `json:"id"`
}

func (q *Queries) LinkExtUser(ctx context.Context, arg LinkExtUserParams) error {
	_, err := q.db.ExecContext(ctx, linkExtUser, arg.FhUserID, arg.ID)
	return err
}

const listExtTeams = `-- name: ListExtTeams :many
SELECT ext_teams.id, ext_teams.name, ext_teams.slug, ext_teams.fh_team_id, fh_teams.name as fh_team_name, fh_teams.slug as fh_team_slug FROM ext_teams
  LEFT JOIN fh_teams ON fh_teams.id = ext_teams.fh_team_id
`

type ListExtTeamsRow struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Slug       string         `json:"slug"`
	FhTeamID   sql.NullString `json:"fh_team_id"`
	FhTeamName sql.NullString `json:"fh_team_name"`
	FhTeamSlug sql.NullString `json:"fh_team_slug"`
}

func (q *Queries) ListExtTeams(ctx context.Context) ([]ListExtTeamsRow, error) {
	rows, err := q.db.QueryContext(ctx, listExtTeams)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListExtTeamsRow
	for rows.Next() {
		var i ListExtTeamsRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Slug,
			&i.FhTeamID,
			&i.FhTeamName,
			&i.FhTeamSlug,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listFhMembersByExtTeamID = `-- name: ListFhMembersByExtTeamID :many
SELECT fh_users.id, fh_users.name, fh_users.email FROM ext_memberships
  JOIN ext_teams ON ext_teams.id = ext_memberships.team_id
  JOIN ext_users ON ext_users.id = ext_memberships.user_id
  JOIN fh_users ON fh_users.id = ext_users.fh_user_id
  LEFT JOIN fh_teams ON fh_teams.id = ext_teams.fh_team_id
WHERE ext_teams.id = ?
`

func (q *Queries) ListFhMembersByExtTeamID(ctx context.Context, id string) ([]FhUser, error) {
	rows, err := q.db.QueryContext(ctx, listFhMembersByExtTeamID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FhUser
	for rows.Next() {
		var i FhUser
		if err := rows.Scan(&i.ID, &i.Name, &i.Email); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listFhTeams = `-- name: ListFhTeams :many
SELECT id, name, slug FROM fh_teams
`

func (q *Queries) ListFhTeams(ctx context.Context) ([]FhTeam, error) {
	rows, err := q.db.QueryContext(ctx, listFhTeams)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FhTeam
	for rows.Next() {
		var i FhTeam
		if err := rows.Scan(&i.ID, &i.Name, &i.Slug); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listFhUsers = `-- name: ListFhUsers :many
SELECT id, name, email FROM fh_users
`

func (q *Queries) ListFhUsers(ctx context.Context) ([]FhUser, error) {
	rows, err := q.db.QueryContext(ctx, listFhUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FhUser
	for rows.Next() {
		var i FhUser
		if err := rows.Scan(&i.ID, &i.Name, &i.Email); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}