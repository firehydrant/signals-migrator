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
