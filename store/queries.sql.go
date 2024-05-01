// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: queries.sql

package store

import (
	"context"
	"database/sql"
)

const deleteExtEscalationPolicyUnimported = `-- name: DeleteExtEscalationPolicyUnimported :exec
DELETE FROM ext_escalation_policies WHERE to_import = 0
`

func (q *Queries) DeleteExtEscalationPolicyUnimported(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteExtEscalationPolicyUnimported)
	return err
}

const deleteUnmatchedExtUsers = `-- name: DeleteUnmatchedExtUsers :exec
DELETE FROM ext_users
WHERE fh_user_id IS NULL
`

func (q *Queries) DeleteUnmatchedExtUsers(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteUnmatchedExtUsers)
	return err
}

const getExtSchedule = `-- name: GetExtSchedule :one
SELECT id, name, description, timezone, strategy, shift_duration, start_time, handoff_time, handoff_day FROM ext_schedules WHERE id = ?
`

func (q *Queries) GetExtSchedule(ctx context.Context, id string) (ExtSchedule, error) {
	row := q.db.QueryRowContext(ctx, getExtSchedule, id)
	var i ExtSchedule
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Timezone,
		&i.Strategy,
		&i.ShiftDuration,
		&i.StartTime,
		&i.HandoffTime,
		&i.HandoffDay,
	)
	return i, err
}

const getExtTeamAnnotation = `-- name: GetExtTeamAnnotation :one
SELECT annotations FROM ext_teams WHERE id = ?
`

func (q *Queries) GetExtTeamAnnotation(ctx context.Context, id string) (string, error) {
	row := q.db.QueryRowContext(ctx, getExtTeamAnnotation, id)
	var annotations string
	err := row.Scan(&annotations)
	return annotations, err
}

const getFhUserByEmail = `-- name: GetFhUserByEmail :one
SELECT id, name, email FROM fh_users WHERE email = ?
`

func (q *Queries) GetFhUserByEmail(ctx context.Context, email string) (FhUser, error) {
	row := q.db.QueryRowContext(ctx, getFhUserByEmail, email)
	var i FhUser
	err := row.Scan(&i.ID, &i.Name, &i.Email)
	return i, err
}

const getTeamByExtID = `-- name: GetTeamByExtID :one
SELECT id, name, slug, fh_team_id, is_group, to_import, annotations, fh_name, fh_slug FROM linked_teams WHERE id = ?
`

func (q *Queries) GetTeamByExtID(ctx context.Context, id string) (LinkedTeam, error) {
	row := q.db.QueryRowContext(ctx, getTeamByExtID, id)
	var i LinkedTeam
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Slug,
		&i.FhTeamID,
		&i.IsGroup,
		&i.ToImport,
		&i.Annotations,
		&i.FhName,
		&i.FhSlug,
	)
	return i, err
}

const getUserByExtID = `-- name: GetUserByExtID :one
SELECT id, name, email, fh_user_id, annotations, fh_name, fh_email FROM linked_users WHERE id = ?
`

func (q *Queries) GetUserByExtID(ctx context.Context, id string) (LinkedUser, error) {
	row := q.db.QueryRowContext(ctx, getUserByExtID, id)
	var i LinkedUser
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.FhUserID,
		&i.Annotations,
		&i.FhName,
		&i.FhEmail,
	)
	return i, err
}

const insertExtEscalationPolicy = `-- name: InsertExtEscalationPolicy :exec
INSERT INTO ext_escalation_policies (id, name, description, team_id, repeat_interval, repeat_limit, handoff_target_type, handoff_target_id, to_import)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
`

type InsertExtEscalationPolicyParams struct {
	ID                string         `json:"id"`
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	TeamID            sql.NullString `json:"team_id"`
	RepeatInterval    sql.NullString `json:"repeat_interval"`
	RepeatLimit       int64          `json:"repeat_limit"`
	HandoffTargetType string         `json:"handoff_target_type"`
	HandoffTargetID   string         `json:"handoff_target_id"`
	ToImport          int64          `json:"to_import"`
}

func (q *Queries) InsertExtEscalationPolicy(ctx context.Context, arg InsertExtEscalationPolicyParams) error {
	_, err := q.db.ExecContext(ctx, insertExtEscalationPolicy,
		arg.ID,
		arg.Name,
		arg.Description,
		arg.TeamID,
		arg.RepeatInterval,
		arg.RepeatLimit,
		arg.HandoffTargetType,
		arg.HandoffTargetID,
		arg.ToImport,
	)
	return err
}

const insertExtEscalationPolicyStep = `-- name: InsertExtEscalationPolicyStep :exec
INSERT INTO ext_escalation_policy_steps (id, escalation_policy_id, position, timeout)
VALUES (?, ?, ?, ?)
`

type InsertExtEscalationPolicyStepParams struct {
	ID                 string `json:"id"`
	EscalationPolicyID string `json:"escalation_policy_id"`
	Position           int64  `json:"position"`
	Timeout            string `json:"timeout"`
}

func (q *Queries) InsertExtEscalationPolicyStep(ctx context.Context, arg InsertExtEscalationPolicyStepParams) error {
	_, err := q.db.ExecContext(ctx, insertExtEscalationPolicyStep,
		arg.ID,
		arg.EscalationPolicyID,
		arg.Position,
		arg.Timeout,
	)
	return err
}

const insertExtEscalationPolicyStepTarget = `-- name: InsertExtEscalationPolicyStepTarget :exec
INSERT INTO ext_escalation_policy_step_targets (escalation_policy_step_id, target_type, target_id)
VALUES (?, ?, ?)
`

type InsertExtEscalationPolicyStepTargetParams struct {
	EscalationPolicyStepID string `json:"escalation_policy_step_id"`
	TargetType             string `json:"target_type"`
	TargetID               string `json:"target_id"`
}

func (q *Queries) InsertExtEscalationPolicyStepTarget(ctx context.Context, arg InsertExtEscalationPolicyStepTargetParams) error {
	_, err := q.db.ExecContext(ctx, insertExtEscalationPolicyStepTarget, arg.EscalationPolicyStepID, arg.TargetType, arg.TargetID)
	return err
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

const insertExtSchedule = `-- name: InsertExtSchedule :exec
INSERT INTO ext_schedules (id, name, description, timezone, strategy, shift_duration, start_time, handoff_time, handoff_day)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
`

type InsertExtScheduleParams struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Timezone      string `json:"timezone"`
	Strategy      string `json:"strategy"`
	ShiftDuration string `json:"shift_duration"`
	StartTime     string `json:"start_time"`
	HandoffTime   string `json:"handoff_time"`
	HandoffDay    string `json:"handoff_day"`
}

func (q *Queries) InsertExtSchedule(ctx context.Context, arg InsertExtScheduleParams) error {
	_, err := q.db.ExecContext(ctx, insertExtSchedule,
		arg.ID,
		arg.Name,
		arg.Description,
		arg.Timezone,
		arg.Strategy,
		arg.ShiftDuration,
		arg.StartTime,
		arg.HandoffTime,
		arg.HandoffDay,
	)
	return err
}

const insertExtScheduleMember = `-- name: InsertExtScheduleMember :exec
INSERT INTO ext_schedule_members (schedule_id, user_id)
VALUES (?, ?)
`

type InsertExtScheduleMemberParams struct {
	ScheduleID string `json:"schedule_id"`
	UserID     string `json:"user_id"`
}

func (q *Queries) InsertExtScheduleMember(ctx context.Context, arg InsertExtScheduleMemberParams) error {
	_, err := q.db.ExecContext(ctx, insertExtScheduleMember, arg.ScheduleID, arg.UserID)
	return err
}

const insertExtScheduleRestriction = `-- name: InsertExtScheduleRestriction :exec
INSERT INTO ext_schedule_restrictions (schedule_id, restriction_index, start_time, start_day, end_time, end_day)
VALUES (?, ?, ?, ?, ?, ?)
`

type InsertExtScheduleRestrictionParams struct {
	ScheduleID       string `json:"schedule_id"`
	RestrictionIndex string `json:"restriction_index"`
	StartTime        string `json:"start_time"`
	StartDay         string `json:"start_day"`
	EndTime          string `json:"end_time"`
	EndDay           string `json:"end_day"`
}

func (q *Queries) InsertExtScheduleRestriction(ctx context.Context, arg InsertExtScheduleRestrictionParams) error {
	_, err := q.db.ExecContext(ctx, insertExtScheduleRestriction,
		arg.ScheduleID,
		arg.RestrictionIndex,
		arg.StartTime,
		arg.StartDay,
		arg.EndTime,
		arg.EndDay,
	)
	return err
}

const insertExtScheduleTeam = `-- name: InsertExtScheduleTeam :exec
INSERT INTO ext_schedule_teams (schedule_id, team_id)
VALUES (?, ?)
`

type InsertExtScheduleTeamParams struct {
	ScheduleID string `json:"schedule_id"`
	TeamID     string `json:"team_id"`
}

func (q *Queries) InsertExtScheduleTeam(ctx context.Context, arg InsertExtScheduleTeamParams) error {
	_, err := q.db.ExecContext(ctx, insertExtScheduleTeam, arg.ScheduleID, arg.TeamID)
	return err
}

const insertExtTeam = `-- name: InsertExtTeam :exec
INSERT INTO ext_teams (id, name, slug, is_group, fh_team_id, annotations)
VALUES (?, ?, ?, ?, ?, ?)
`

type InsertExtTeamParams struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	IsGroup     int64          `json:"is_group"`
	FhTeamID    sql.NullString `json:"fh_team_id"`
	Annotations string         `json:"annotations"`
}

func (q *Queries) InsertExtTeam(ctx context.Context, arg InsertExtTeamParams) error {
	_, err := q.db.ExecContext(ctx, insertExtTeam,
		arg.ID,
		arg.Name,
		arg.Slug,
		arg.IsGroup,
		arg.FhTeamID,
		arg.Annotations,
	)
	return err
}

const insertExtTeamGroup = `-- name: InsertExtTeamGroup :exec
INSERT INTO ext_team_groups (group_team_id, member_team_id) VALUES (?, ?)
`

type InsertExtTeamGroupParams struct {
	GroupTeamID  string `json:"group_team_id"`
	MemberTeamID string `json:"member_team_id"`
}

func (q *Queries) InsertExtTeamGroup(ctx context.Context, arg InsertExtTeamGroupParams) error {
	_, err := q.db.ExecContext(ctx, insertExtTeamGroup, arg.GroupTeamID, arg.MemberTeamID)
	return err
}

const insertExtUser = `-- name: InsertExtUser :exec
INSERT INTO ext_users (id, name, email, fh_user_id, annotations)
VALUES (?, ?, ?, ?, ?)
`

type InsertExtUserParams struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Email       string         `json:"email"`
	FhUserID    sql.NullString `json:"fh_user_id"`
	Annotations string         `json:"annotations"`
}

func (q *Queries) InsertExtUser(ctx context.Context, arg InsertExtUserParams) error {
	_, err := q.db.ExecContext(ctx, insertExtUser,
		arg.ID,
		arg.Name,
		arg.Email,
		arg.FhUserID,
		arg.Annotations,
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

const listExtEscalationPolicies = `-- name: ListExtEscalationPolicies :many
SELECT id, name, description, team_id, repeat_limit, repeat_interval, handoff_target_type, handoff_target_id, to_import FROM ext_escalation_policies
`

func (q *Queries) ListExtEscalationPolicies(ctx context.Context) ([]ExtEscalationPolicy, error) {
	rows, err := q.db.QueryContext(ctx, listExtEscalationPolicies)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExtEscalationPolicy
	for rows.Next() {
		var i ExtEscalationPolicy
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.TeamID,
			&i.RepeatLimit,
			&i.RepeatInterval,
			&i.HandoffTargetType,
			&i.HandoffTargetID,
			&i.ToImport,
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

const listExtEscalationPolicyStepTargets = `-- name: ListExtEscalationPolicyStepTargets :many
SELECT escalation_policy_step_id, target_type, target_id FROM ext_escalation_policy_step_targets
WHERE escalation_policy_step_id = ?
`

func (q *Queries) ListExtEscalationPolicyStepTargets(ctx context.Context, escalationPolicyStepID string) ([]ExtEscalationPolicyStepTarget, error) {
	rows, err := q.db.QueryContext(ctx, listExtEscalationPolicyStepTargets, escalationPolicyStepID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExtEscalationPolicyStepTarget
	for rows.Next() {
		var i ExtEscalationPolicyStepTarget
		if err := rows.Scan(&i.EscalationPolicyStepID, &i.TargetType, &i.TargetID); err != nil {
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

const listExtEscalationPolicySteps = `-- name: ListExtEscalationPolicySteps :many
SELECT id, escalation_policy_id, position, timeout FROM ext_escalation_policy_steps
WHERE escalation_policy_id = ?
ORDER BY position ASC
`

func (q *Queries) ListExtEscalationPolicySteps(ctx context.Context, escalationPolicyID string) ([]ExtEscalationPolicyStep, error) {
	rows, err := q.db.QueryContext(ctx, listExtEscalationPolicySteps, escalationPolicyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExtEscalationPolicyStep
	for rows.Next() {
		var i ExtEscalationPolicyStep
		if err := rows.Scan(
			&i.ID,
			&i.EscalationPolicyID,
			&i.Position,
			&i.Timeout,
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

const listExtScheduleRestrictionsByExtScheduleID = `-- name: ListExtScheduleRestrictionsByExtScheduleID :many
SELECT schedule_id, restriction_index, start_time, start_day, end_time, end_day FROM ext_schedule_restrictions WHERE schedule_id = ?
`

func (q *Queries) ListExtScheduleRestrictionsByExtScheduleID(ctx context.Context, scheduleID string) ([]ExtScheduleRestriction, error) {
	rows, err := q.db.QueryContext(ctx, listExtScheduleRestrictionsByExtScheduleID, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExtScheduleRestriction
	for rows.Next() {
		var i ExtScheduleRestriction
		if err := rows.Scan(
			&i.ScheduleID,
			&i.RestrictionIndex,
			&i.StartTime,
			&i.StartDay,
			&i.EndTime,
			&i.EndDay,
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

const listExtSchedules = `-- name: ListExtSchedules :many
SELECT id, name, description, timezone, strategy, shift_duration, start_time, handoff_time, handoff_day FROM ext_schedules
`

func (q *Queries) ListExtSchedules(ctx context.Context) ([]ExtSchedule, error) {
	rows, err := q.db.QueryContext(ctx, listExtSchedules)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExtSchedule
	for rows.Next() {
		var i ExtSchedule
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Timezone,
			&i.Strategy,
			&i.ShiftDuration,
			&i.StartTime,
			&i.HandoffTime,
			&i.HandoffDay,
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

const listExtSchedulesLikeID = `-- name: ListExtSchedulesLikeID :many
SELECT id, name, description, timezone, strategy, shift_duration, start_time, handoff_time, handoff_day FROM ext_schedules WHERE id LIKE ?
`

func (q *Queries) ListExtSchedulesLikeID(ctx context.Context, id string) ([]ExtSchedule, error) {
	rows, err := q.db.QueryContext(ctx, listExtSchedulesLikeID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExtSchedule
	for rows.Next() {
		var i ExtSchedule
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Timezone,
			&i.Strategy,
			&i.ShiftDuration,
			&i.StartTime,
			&i.HandoffTime,
			&i.HandoffDay,
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

const listExtTeamMemberships = `-- name: ListExtTeamMemberships :many
SELECT ext_teams.id, ext_teams.name, ext_teams.slug, ext_teams.fh_team_id, ext_teams.is_group, ext_teams.to_import, ext_teams.annotations, ext_users.id, ext_users.name, ext_users.email, ext_users.fh_user_id, ext_users.annotations FROM ext_memberships
  JOIN ext_teams ON ext_teams.id = ext_memberships.team_id
  JOIN ext_users ON ext_users.id = ext_memberships.user_id
`

type ListExtTeamMembershipsRow struct {
	ExtTeam ExtTeam `json:"ext_team"`
	ExtUser ExtUser `json:"ext_user"`
}

func (q *Queries) ListExtTeamMemberships(ctx context.Context) ([]ListExtTeamMembershipsRow, error) {
	rows, err := q.db.QueryContext(ctx, listExtTeamMemberships)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListExtTeamMembershipsRow
	for rows.Next() {
		var i ListExtTeamMembershipsRow
		if err := rows.Scan(
			&i.ExtTeam.ID,
			&i.ExtTeam.Name,
			&i.ExtTeam.Slug,
			&i.ExtTeam.FhTeamID,
			&i.ExtTeam.IsGroup,
			&i.ExtTeam.ToImport,
			&i.ExtTeam.Annotations,
			&i.ExtUser.ID,
			&i.ExtUser.Name,
			&i.ExtUser.Email,
			&i.ExtUser.FhUserID,
			&i.ExtUser.Annotations,
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

const listExtTeams = `-- name: ListExtTeams :many
SELECT id, name, slug, fh_team_id, is_group, to_import, annotations FROM ext_teams
`

func (q *Queries) ListExtTeams(ctx context.Context) ([]ExtTeam, error) {
	rows, err := q.db.QueryContext(ctx, listExtTeams)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExtTeam
	for rows.Next() {
		var i ExtTeam
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Slug,
			&i.FhTeamID,
			&i.IsGroup,
			&i.ToImport,
			&i.Annotations,
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

const listExtUsers = `-- name: ListExtUsers :many
SELECT id, name, email, fh_user_id, annotations FROM ext_users
`

func (q *Queries) ListExtUsers(ctx context.Context) ([]ExtUser, error) {
	rows, err := q.db.QueryContext(ctx, listExtUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExtUser
	for rows.Next() {
		var i ExtUser
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Email,
			&i.FhUserID,
			&i.Annotations,
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

const listFhMembersByExtScheduleID = `-- name: ListFhMembersByExtScheduleID :many
SELECT fh_users.id, fh_users.name, fh_users.email FROM ext_schedule_members
  JOIN ext_users ON ext_users.id = ext_schedule_members.user_id
  JOIN fh_users ON fh_users.id = ext_users.fh_user_id
WHERE ext_schedule_members.schedule_id = ?
`

func (q *Queries) ListFhMembersByExtScheduleID(ctx context.Context, scheduleID string) ([]FhUser, error) {
	rows, err := q.db.QueryContext(ctx, listFhMembersByExtScheduleID, scheduleID)
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

const listFhUserAnnotations = `-- name: ListFhUserAnnotations :many
SELECT annotations FROM linked_users WHERE fh_user_id = ?
`

func (q *Queries) ListFhUserAnnotations(ctx context.Context, fhUserID sql.NullString) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, listFhUserAnnotations, fhUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var annotations string
		if err := rows.Scan(&annotations); err != nil {
			return nil, err
		}
		items = append(items, annotations)
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

const listGroupExtTeamMemberships = `-- name: ListGroupExtTeamMemberships :many
SELECT ext_teams.id, ext_teams.name, ext_teams.slug, ext_teams.fh_team_id, ext_teams.is_group, ext_teams.to_import, ext_teams.annotations, ext_users.id, ext_users.name, ext_users.email, ext_users.fh_user_id, ext_users.annotations FROM ext_teams
  JOIN ext_team_groups ON ext_team_groups.group_team_id = ext_teams.id
  JOIN ext_teams AS member_team ON member_team.id = ext_team_groups.member_team_id
  JOIN ext_memberships ON ext_memberships.team_id = member_team.id
  JOIN ext_users ON ext_users.id = ext_memberships.user_id
`

type ListGroupExtTeamMembershipsRow struct {
	ExtTeam ExtTeam `json:"ext_team"`
	ExtUser ExtUser `json:"ext_user"`
}

func (q *Queries) ListGroupExtTeamMemberships(ctx context.Context) ([]ListGroupExtTeamMembershipsRow, error) {
	rows, err := q.db.QueryContext(ctx, listGroupExtTeamMemberships)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListGroupExtTeamMembershipsRow
	for rows.Next() {
		var i ListGroupExtTeamMembershipsRow
		if err := rows.Scan(
			&i.ExtTeam.ID,
			&i.ExtTeam.Name,
			&i.ExtTeam.Slug,
			&i.ExtTeam.FhTeamID,
			&i.ExtTeam.IsGroup,
			&i.ExtTeam.ToImport,
			&i.ExtTeam.Annotations,
			&i.ExtUser.ID,
			&i.ExtUser.Name,
			&i.ExtUser.Email,
			&i.ExtUser.FhUserID,
			&i.ExtUser.Annotations,
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

const listGroupExtTeams = `-- name: ListGroupExtTeams :many
SELECT id, name, slug, fh_team_id, is_group, to_import, annotations FROM ext_teams
WHERE is_group = 1
`

func (q *Queries) ListGroupExtTeams(ctx context.Context) ([]ExtTeam, error) {
	rows, err := q.db.QueryContext(ctx, listGroupExtTeams)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExtTeam
	for rows.Next() {
		var i ExtTeam
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Slug,
			&i.FhTeamID,
			&i.IsGroup,
			&i.ToImport,
			&i.Annotations,
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

const listMemberExtTeams = `-- name: ListMemberExtTeams :many
SELECT id, name, slug, fh_team_id, is_group, to_import, annotations FROM ext_teams
WHERE id IN (
  SELECT DISTINCT member_team_id FROM ext_team_groups
  WHERE group_team_id = ?
)
`

func (q *Queries) ListMemberExtTeams(ctx context.Context, groupTeamID string) ([]ExtTeam, error) {
	rows, err := q.db.QueryContext(ctx, listMemberExtTeams, groupTeamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExtTeam
	for rows.Next() {
		var i ExtTeam
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Slug,
			&i.FhTeamID,
			&i.IsGroup,
			&i.ToImport,
			&i.Annotations,
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

const listNonGroupExtTeams = `-- name: ListNonGroupExtTeams :many
SELECT id, name, slug, fh_team_id, is_group, to_import, annotations FROM ext_teams
WHERE is_group = 0
`

func (q *Queries) ListNonGroupExtTeams(ctx context.Context) ([]ExtTeam, error) {
	rows, err := q.db.QueryContext(ctx, listNonGroupExtTeams)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExtTeam
	for rows.Next() {
		var i ExtTeam
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Slug,
			&i.FhTeamID,
			&i.IsGroup,
			&i.ToImport,
			&i.Annotations,
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

const listTeams = `-- name: ListTeams :many
SELECT id, name, slug, fh_team_id, is_group, to_import, annotations, fh_name, fh_slug from linked_teams
`

func (q *Queries) ListTeams(ctx context.Context) ([]LinkedTeam, error) {
	rows, err := q.db.QueryContext(ctx, listTeams)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []LinkedTeam
	for rows.Next() {
		var i LinkedTeam
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Slug,
			&i.FhTeamID,
			&i.IsGroup,
			&i.ToImport,
			&i.Annotations,
			&i.FhName,
			&i.FhSlug,
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

const listTeamsByExtScheduleID = `-- name: ListTeamsByExtScheduleID :many
SELECT linked_teams.id, linked_teams.name, linked_teams.slug, linked_teams.fh_team_id, linked_teams.is_group, linked_teams.to_import, linked_teams.annotations, linked_teams.fh_name, linked_teams.fh_slug FROM linked_teams
  JOIN ext_schedule_teams ON linked_teams.id = ext_schedule_teams.team_id
WHERE ext_schedule_teams.schedule_id = ?
`

func (q *Queries) ListTeamsByExtScheduleID(ctx context.Context, scheduleID string) ([]LinkedTeam, error) {
	rows, err := q.db.QueryContext(ctx, listTeamsByExtScheduleID, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []LinkedTeam
	for rows.Next() {
		var i LinkedTeam
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Slug,
			&i.FhTeamID,
			&i.IsGroup,
			&i.ToImport,
			&i.Annotations,
			&i.FhName,
			&i.FhSlug,
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

const listTeamsToImport = `-- name: ListTeamsToImport :many
SELECT id, name, slug, fh_team_id, is_group, to_import, annotations, fh_name, fh_slug from linked_teams WHERE to_import = 1
`

func (q *Queries) ListTeamsToImport(ctx context.Context) ([]LinkedTeam, error) {
	rows, err := q.db.QueryContext(ctx, listTeamsToImport)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []LinkedTeam
	for rows.Next() {
		var i LinkedTeam
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Slug,
			&i.FhTeamID,
			&i.IsGroup,
			&i.ToImport,
			&i.Annotations,
			&i.FhName,
			&i.FhSlug,
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

const listUnmatchedExtUsers = `-- name: ListUnmatchedExtUsers :many
SELECT id, name, email, fh_user_id, annotations FROM ext_users
WHERE fh_user_id IS NULL
`

func (q *Queries) ListUnmatchedExtUsers(ctx context.Context) ([]ExtUser, error) {
	rows, err := q.db.QueryContext(ctx, listUnmatchedExtUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExtUser
	for rows.Next() {
		var i ExtUser
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Email,
			&i.FhUserID,
			&i.Annotations,
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

const listUsersJoinByEmail = `-- name: ListUsersJoinByEmail :many
SELECT ext_users.id, ext_users.name, ext_users.email, ext_users.fh_user_id, ext_users.annotations, fh_users.id, fh_users.name, fh_users.email FROM ext_users
  JOIN fh_users ON fh_users.email = ext_users.email
`

type ListUsersJoinByEmailRow struct {
	ExtUser ExtUser `json:"ext_user"`
	FhUser  FhUser  `json:"fh_user"`
}

func (q *Queries) ListUsersJoinByEmail(ctx context.Context) ([]ListUsersJoinByEmailRow, error) {
	rows, err := q.db.QueryContext(ctx, listUsersJoinByEmail)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListUsersJoinByEmailRow
	for rows.Next() {
		var i ListUsersJoinByEmailRow
		if err := rows.Scan(
			&i.ExtUser.ID,
			&i.ExtUser.Name,
			&i.ExtUser.Email,
			&i.ExtUser.FhUserID,
			&i.ExtUser.Annotations,
			&i.FhUser.ID,
			&i.FhUser.Name,
			&i.FhUser.Email,
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

const markAllExtEscalationPolicyToImport = `-- name: MarkAllExtEscalationPolicyToImport :exec
UPDATE ext_escalation_policies SET to_import = 1
`

func (q *Queries) MarkAllExtEscalationPolicyToImport(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, markAllExtEscalationPolicyToImport)
	return err
}

const markExtEscalationPolicyToImport = `-- name: MarkExtEscalationPolicyToImport :exec
UPDATE ext_escalation_policies SET to_import = 1 WHERE id = ?
`

func (q *Queries) MarkExtEscalationPolicyToImport(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, markExtEscalationPolicyToImport, id)
	return err
}

const markExtTeamToImport = `-- name: MarkExtTeamToImport :exec
UPDATE ext_teams SET to_import = 1 WHERE id = ?
`

func (q *Queries) MarkExtTeamToImport(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, markExtTeamToImport, id)
	return err
}
