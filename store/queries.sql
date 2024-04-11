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

-- name: CheckExtTeamExists :one
SELECT COUNT(*) > 0 FROM ext_teams WHERE id = ?;

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

-- name: ListExtSchedules :many
SELECT * FROM ext_schedules;

-- name: InsertExtSchedule :exec
INSERT INTO ext_schedules (id, name, description, timezone, strategy, shift_duration, start_time, handoff_time, handoff_day) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListExtScheduleRestrictionsByExtScheduleID :many
SELECT * FROM ext_schedule_restrictions WHERE schedule_id = ?;

-- name: InsertExtScheduleRestriction :exec
INSERT INTO ext_schedule_restrictions (schedule_id, restriction_index, start_time, start_day, end_time, end_day) VALUES (?, ?, ?, ?, ?, ?);

-- name: ListFhTeamsByExtScheduleID :many
SELECT ext_teams.*, fh_teams.name as fh_team_name, fh_teams.slug as fh_team_slug FROM ext_schedule_teams
  JOIN ext_teams ON ext_teams.id = ext_schedule_teams.team_id
  LEFT JOIN fh_teams ON fh_teams.id = ext_teams.fh_team_id
WHERE ext_schedule_teams.schedule_id = ?;

-- name: InsertExtScheduleTeam :exec
INSERT INTO ext_schedule_teams (schedule_id, team_id) VALUES (?, ?);

-- name: ListFhMembersByExtScheduleID :many
SELECT fh_users.* FROM ext_schedule_members
  JOIN ext_users ON ext_users.id = ext_schedule_members.user_id
  JOIN fh_users ON fh_users.id = ext_users.fh_user_id
WHERE ext_schedule_members.schedule_id = ?;

-- name: InsertExtScheduleMember :exec
INSERT INTO ext_schedule_members (schedule_id, user_id) VALUES (?, ?);
