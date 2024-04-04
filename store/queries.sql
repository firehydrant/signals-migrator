-- name: ListFhUsers :many
SELECT * FROM fh_users;

-- name: GetFhUserByEmail :one
SELECT * FROM fh_users WHERE email = ?;

-- name: InsertFhUser :exec
INSERT INTO fh_users (id, name, email) VALUES (?, ?, ?);

-- name: ListFhTeams :many
SELECT * FROM fh_teams;

-- name: InsertFhTeam :exec
INSERT INTO fh_teams (id, name, slug) VALUES (?, ?, ?);

-- name: InsertExtUser :exec
INSERT INTO ext_users (id, name, email, fh_user_id) VALUES (?, ?, ?, ?);

-- name: ListExtTeams :many
SELECT ext_teams.*, fh_teams.name as fh_team_name, fh_teams.slug as fh_team_slug FROM ext_teams
  LEFT JOIN fh_teams ON fh_teams.id = ext_teams.fh_team_id;

-- name: InsertExtTeam :exec
INSERT INTO ext_teams (id, name, slug, fh_team_id) VALUES (?, ?, ?, ?);

-- name: LinkExtUser :exec
UPDATE ext_users SET fh_user_id = ? WHERE id = ?;

-- name: LinkExtTeam :exec
UPDATE ext_teams SET fh_team_id = ? WHERE id = ?;

-- name: InsertExtMembership :exec
INSERT INTO ext_memberships (user_id, team_id) VALUES (?, ?);

-- name: ListFhMembersByExtTeamID :many
SELECT fh_users.* FROM ext_memberships
  JOIN ext_teams ON ext_teams.id = ext_memberships.team_id
  JOIN ext_users ON ext_users.id = ext_memberships.user_id
  JOIN fh_users ON fh_users.id = ext_users.fh_user_id
  LEFT JOIN fh_teams ON fh_teams.id = ext_teams.fh_team_id
WHERE ext_teams.id = ?;
