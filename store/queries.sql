-- name: ListFhUsers :many
SELECT * FROM fh_users;

-- name: GetFhUserByEmail :one
SELECT * FROM fh_users WHERE email = ?;

-- name: InsertFhUser :exec
INSERT INTO fh_users (id, name, email) VALUES (?, ?, ?);

-- name: GetUserByExtID :one
SELECT * FROM linked_users WHERE id = ?;

-- name: ListFhTeams :many
SELECT * FROM fh_teams;

-- name: InsertFhTeam :exec
INSERT INTO fh_teams (id, name, slug) VALUES (?, ?, ?);

-- name: InsertExtUser :exec
INSERT INTO ext_users (id, name, email, fh_user_id) VALUES (?, ?, ?, ?);

-- name: GetTeamByExtID :one
SELECT * FROM linked_teams WHERE id = ?;

-- name: ListTeams :many
SELECT * from linked_teams;

-- name: ListExtTeams :many
SELECT * FROM ext_teams;

-- name: InsertExtTeam :exec
INSERT INTO ext_teams (id, name, slug, fh_team_id, metadata) VALUES (?, ?, ?, ?, ?);

-- name: LinkExtUser :exec
UPDATE ext_users SET fh_user_id = ? WHERE id = ?;

-- name: LinkExtTeam :exec
UPDATE ext_teams SET fh_team_id = ? WHERE id = ?;

-- name: ListExtMemberships :many
SELECT sqlc.embed(ext_teams), sqlc.embed(ext_users) FROM ext_memberships
  JOIN ext_teams ON ext_teams.id = ext_memberships.team_id
  JOIN ext_users ON ext_users.id = ext_memberships.user_id;
;

-- name: InsertExtMembership :exec
INSERT INTO ext_memberships (user_id, team_id) VALUES (?, ?);

-- name: ListFhMembersByExtTeamID :many
SELECT fh_users.* FROM ext_memberships
  JOIN ext_teams ON ext_teams.id = ext_memberships.team_id
  JOIN ext_users ON ext_users.id = ext_memberships.user_id
  JOIN fh_users ON fh_users.id = ext_users.fh_user_id
  LEFT JOIN fh_teams ON fh_teams.id = ext_teams.fh_team_id
WHERE ext_teams.id = ?;

-- name: GetExtSchedule :one
SELECT * FROM ext_schedules WHERE id = ?;

-- name: ListExtSchedules :many
SELECT * FROM ext_schedules;

-- name: InsertExtSchedule :exec
INSERT INTO ext_schedules (id, name, description, timezone, strategy, shift_duration, start_time, handoff_time, handoff_day)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListExtScheduleRestrictionsByExtScheduleID :many
SELECT * FROM ext_schedule_restrictions WHERE schedule_id = ?;

-- name: InsertExtScheduleRestriction :exec
INSERT INTO ext_schedule_restrictions (schedule_id, restriction_index, start_time, start_day, end_time, end_day)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListTeamsByExtScheduleID :many
SELECT linked_teams.* FROM linked_teams
  JOIN ext_schedule_teams ON linked_teams.id = ext_schedule_teams.team_id
WHERE ext_schedule_teams.schedule_id = ?;

-- name: InsertExtScheduleTeam :exec
INSERT INTO ext_schedule_teams (schedule_id, team_id)
VALUES (?, ?);

-- name: ListFhMembersByExtScheduleID :many
SELECT fh_users.* FROM ext_schedule_members
  JOIN ext_users ON ext_users.id = ext_schedule_members.user_id
  JOIN fh_users ON fh_users.id = ext_users.fh_user_id
WHERE ext_schedule_members.schedule_id = ?;

-- name: InsertExtScheduleMember :exec
INSERT INTO ext_schedule_members (schedule_id, user_id)
VALUES (?, ?);

-- name: ListExtEscalationPolicies :many
SELECT * FROM ext_escalation_policies;

-- name: InsertExtEscalationPolicy :exec
INSERT INTO ext_escalation_policies (id, name, description, team_id, repeat_interval, repeat_limit, handoff_target_type, handoff_target_id, to_import)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: MarkAllExtEscalationPolicyToImport :exec
UPDATE ext_escalation_policies SET to_import = 1;

-- name: MarkExtEscalationPolicyToImport :exec
UPDATE ext_escalation_policies SET to_import = 1 WHERE id = ?;

-- name: DeleteExtEscalationPolicyUnimported :exec
DELETE FROM ext_escalation_policies WHERE to_import = 0;

-- name: ListExtEscalationPolicySteps :many
SELECT * FROM ext_escalation_policy_steps
WHERE escalation_policy_id = ?
ORDER BY position ASC;

-- name: InsertExtEscalationPolicyStep :exec
INSERT INTO ext_escalation_policy_steps (id, escalation_policy_id, position, timeout)
VALUES (?, ?, ?, ?);

-- name: ListExtEscalationPolicyStepTargets :many
SELECT * FROM ext_escalation_policy_step_targets
WHERE escalation_policy_step_id = ?;

-- name: InsertExtEscalationPolicyStepTarget :exec
INSERT INTO ext_escalation_policy_step_targets (escalation_policy_step_id, target_type, target_id)
VALUES (?, ?, ?);
