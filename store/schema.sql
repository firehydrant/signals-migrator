CREATE TABLE IF NOT EXISTS fh_users (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS ext_users (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT NOT NULL,
  fh_user_id TEXT REFERENCES fh_users(id)
) STRICT;

CREATE TABLE IF NOT EXISTS fh_teams (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  slug TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS ext_teams (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  slug TEXT NOT NULL,
  fh_team_id TEXT REFERENCES fh_teams(id)
) STRICT;

CREATE TABLE IF NOT EXISTS ext_memberships (
  user_id TEXT NOT NULL,
  team_id TEXT NOT NULL,
  PRIMARY KEY (user_id, team_id),
  FOREIGN KEY (user_id) REFERENCES ext_users(id),
  FOREIGN KEY (team_id) REFERENCES ext_teams(id)
) STRICT;

CREATE TABLE IF NOT EXISTS ext_schedules (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL,
  timezone TEXT NOT NULL,
  strategy TEXT NOT NULL,
  shift_duration TEXT NOT NULL,
  start_time TEXT NOT NULL,
  handoff_time TEXT NOT NULL,
  handoff_day TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS ext_schedule_restrictions (
  schedule_id TEXT NOT NULL,
  restriction_index TEXT NOT NULL,
  start_time TEXT NOT NULL,
  start_day TEXT NOT NULL,
  end_time TEXT NOT NULL,
  end_day TEXT NOT NULL,
  PRIMARY KEY (schedule_id, restriction_index),
  FOREIGN KEY (schedule_id) REFERENCES ext_schedules(id)
) STRICT;

CREATE TABLE IF NOT EXISTS ext_schedule_teams (
  schedule_id TEXT NOT NULL,
  team_id TEXT NOT NULL,
  PRIMARY KEY (schedule_id, team_id),
  FOREIGN KEY (schedule_id) REFERENCES ext_schedules(id),
  FOREIGN KEY (team_id) REFERENCES ext_teams(id)
) STRICT;

CREATE TABLE IF NOT EXISTS ext_schedule_members (
  schedule_id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  PRIMARY KEY (schedule_id, user_id),
  FOREIGN KEY (schedule_id) REFERENCES ext_schedules(id),
  FOREIGN KEY (user_id) REFERENCES ext_users(id)
) STRICT;

CREATE TABLE IF NOT EXISTS ext_escalation_policies (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL,
  team_id TEXT REFERENCES ext_teams(id),
  repeat INTEGER NOT NULL,
  handoff_target_type TEXT NOT NULL,
  handoff_target_id TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS ext_escalation_policy_steps (
  id TEXT PRIMARY KEY,
  escalation_policy_id TEXT NOT NULL,
  position INTEGER NOT NULL,
  timeout TEXT NOT NULL,
  FOREIGN KEY (escalation_policy_id) REFERENCES ext_escalation_policies(id)
) STRICT;

CREATE TABLE IF NOT EXISTS ext_escalation_policy_targets (
  escalation_policy_step_id TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id TEXT NOT NULL,
  PRIMARY KEY (escalation_policy_step_id, target_type, target_id),
  FOREIGN KEY (escalation_policy_step_id) REFERENCES ext_escalation_policy_steps(id)
) STRICT;
