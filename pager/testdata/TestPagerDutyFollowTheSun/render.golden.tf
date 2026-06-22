terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant"
      version = ">= 0.15.2"
    }
  }
}

data "firehydrant_user" "avery" {
  email = "avery@example.com"
  # [PagerDuty] avery@example.com https://example.pagerduty.com/users/U1AVERY
}

data "firehydrant_user" "blake" {
  email = "blake@example.com"
  # [PagerDuty] blake@example.com https://example.pagerduty.com/users/U1BLAKE
}

data "firehydrant_user" "cam" {
  email = "cam@example.com"
  # [PagerDuty] cam@example.com https://example.pagerduty.com/users/U1CAM
}

data "firehydrant_user" "drew" {
  email = "drew@example.com"
  # [PagerDuty] drew@example.com https://example.pagerduty.com/users/U1DREW
}

data "firehydrant_user" "evan" {
  email = "evan@example.com"
  # [PagerDuty] evan@example.com https://example.pagerduty.com/users/U1EVAN
}

data "firehydrant_user" "faye" {
  email = "faye@example.com"
  # [PagerDuty] faye@example.com https://example.pagerduty.com/users/U1FAYE
}

data "firehydrant_user" "gale" {
  email = "gale@example.com"
  # [PagerDuty] gale@example.com https://example.pagerduty.com/users/U1GALE
}

data "firehydrant_user" "harlan" {
  email = "harlan@example.com"
  # [PagerDuty] harlan@example.com https://example.pagerduty.com/users/U1HARLAN
}

data "firehydrant_user" "indra" {
  email = "indra@example.com"
  # [PagerDuty] indra@example.com https://example.pagerduty.com/users/U1INDRA
}

data "firehydrant_user" "jules" {
  email = "jules@example.com"
  # [PagerDuty] jules@example.com https://example.pagerduty.com/users/U1JULES
}

data "firehydrant_user" "kai" {
  email = "kai@example.com"
  # [PagerDuty] kai@example.com https://example.pagerduty.com/users/U1KAI
}

data "firehydrant_user" "liu" {
  email = "liu@example.com"
  # [PagerDuty] liu@example.com https://example.pagerduty.com/users/U1LIU
}

data "firehydrant_user" "morgan" {
  email = "morgan@example.com"
  # [PagerDuty] morgan@example.com https://example.pagerduty.com/users/U1MORGAN
}

data "firehydrant_user" "noor" {
  email = "noor@example.com"
  # [PagerDuty] noor@example.com https://example.pagerduty.com/users/U1NOOR
}

resource "firehydrant_team" "fh_team_T1FTS" {
  name = "Platform On-Call"

  # [PagerDuty] Platform On-Call https://example.pagerduty.com/teams/T1FTS

  memberships {
    user_id = data.firehydrant_user.avery.id
  }

  memberships {
    user_id = data.firehydrant_user.blake.id
  }

  memberships {
    user_id = data.firehydrant_user.cam.id
  }

  memberships {
    user_id = data.firehydrant_user.drew.id
  }

  memberships {
    user_id = data.firehydrant_user.evan.id
  }

  memberships {
    user_id = data.firehydrant_user.faye.id
  }

  memberships {
    user_id = data.firehydrant_user.gale.id
  }

  memberships {
    user_id = data.firehydrant_user.harlan.id
  }

  memberships {
    user_id = data.firehydrant_user.indra.id
  }

  memberships {
    user_id = data.firehydrant_user.jules.id
  }

  memberships {
    user_id = data.firehydrant_user.kai.id
  }

  memberships {
    user_id = data.firehydrant_user.liu.id
  }

  memberships {
    user_id = data.firehydrant_user.morgan.id
  }

  memberships {
    user_id = data.firehydrant_user.noor.id
  }
}

import {
  id = "fh-team-T1FTS"
  to = firehydrant_team.fh_team_T1FTS
}

resource "firehydrant_on_call_schedule" "fh_team_T1FTS_follow_the_sun_primary" {
  name                 = "Follow-The-Sun Primary"
  team_id              = firehydrant_team.fh_team_T1FTS.id
  rotation_name        = "APAC"
  rotation_description = "(APAC)"
  time_zone            = "Etc/UTC"
  start_time           = "2025-06-02T00:00:00Z"

  member_ids = [data.firehydrant_user.avery.id, data.firehydrant_user.blake.id, data.firehydrant_user.cam.id]

  strategy {
    type         = "weekly"
    handoff_day  = "monday"
    handoff_time = "00:00:00"
  }

  restrictions {
    start_day  = "monday"
    start_time = "00:00:00"
    end_day    = "monday"
    end_time   = "08:00:00"
  }

  restrictions {
    start_day  = "tuesday"
    start_time = "00:00:00"
    end_day    = "tuesday"
    end_time   = "08:00:00"
  }

  restrictions {
    start_day  = "wednesday"
    start_time = "00:00:00"
    end_day    = "wednesday"
    end_time   = "08:00:00"
  }

  restrictions {
    start_day  = "thursday"
    start_time = "00:00:00"
    end_day    = "thursday"
    end_time   = "08:00:00"
  }

  restrictions {
    start_day  = "friday"
    start_time = "00:00:00"
    end_day    = "friday"
    end_time   = "08:00:00"
  }

  # [PagerDuty] Platform On-Call https://example.pagerduty.com/teams/T1FTS
}

resource "firehydrant_rotation" "fh_team_T1FTS_follow_the_sun_primary_emea" {
  name        = "EMEA"
  description = "(EMEA)"
  team_id     = firehydrant_team.fh_team_T1FTS.id
  schedule_id = firehydrant_on_call_schedule.fh_team_T1FTS_follow_the_sun_primary.id
  time_zone   = "Etc/UTC"
  start_time  = "2025-06-02T08:00:00Z"

  members {
    user_id = data.firehydrant_user.drew.id
  }
  members {
    user_id = data.firehydrant_user.evan.id
  }
  members {
    user_id = data.firehydrant_user.faye.id
  }
  members {
    user_id = data.firehydrant_user.gale.id
  }

  strategy {
    type         = "weekly"
    handoff_day  = "monday"
    handoff_time = "08:00:00"
  }

  restrictions {
    start_day  = "monday"
    start_time = "08:00:00"
    end_day    = "monday"
    end_time   = "16:00:00"
  }

  restrictions {
    start_day  = "tuesday"
    start_time = "08:00:00"
    end_day    = "tuesday"
    end_time   = "16:00:00"
  }

  restrictions {
    start_day  = "wednesday"
    start_time = "08:00:00"
    end_day    = "wednesday"
    end_time   = "16:00:00"
  }

  restrictions {
    start_day  = "thursday"
    start_time = "08:00:00"
    end_day    = "thursday"
    end_time   = "16:00:00"
  }

  restrictions {
    start_day  = "friday"
    start_time = "08:00:00"
    end_day    = "friday"
    end_time   = "16:00:00"
  }

  # [PagerDuty] Platform On-Call https://example.pagerduty.com/teams/T1FTS
}

resource "firehydrant_rotation" "fh_team_T1FTS_follow_the_sun_primary_americas" {
  name        = "Americas"
  description = "(Americas)"
  team_id     = firehydrant_team.fh_team_T1FTS.id
  schedule_id = firehydrant_on_call_schedule.fh_team_T1FTS_follow_the_sun_primary.id
  time_zone   = "Etc/UTC"
  start_time  = "2025-06-02T16:00:00Z"

  members {
    user_id = data.firehydrant_user.harlan.id
  }
  members {
    user_id = data.firehydrant_user.indra.id
  }
  members {
    user_id = data.firehydrant_user.jules.id
  }
  members {
    user_id = data.firehydrant_user.kai.id
  }
  members {
    user_id = data.firehydrant_user.liu.id
  }

  strategy {
    type         = "weekly"
    handoff_day  = "monday"
    handoff_time = "16:00:00"
  }

  restrictions {
    start_day  = "monday"
    start_time = "16:00:00"
    end_day    = "tuesday"
    end_time   = "00:00:00"
  }

  restrictions {
    start_day  = "tuesday"
    start_time = "16:00:00"
    end_day    = "wednesday"
    end_time   = "00:00:00"
  }

  restrictions {
    start_day  = "wednesday"
    start_time = "16:00:00"
    end_day    = "thursday"
    end_time   = "00:00:00"
  }

  restrictions {
    start_day  = "thursday"
    start_time = "16:00:00"
    end_day    = "friday"
    end_time   = "00:00:00"
  }

  restrictions {
    start_day  = "friday"
    start_time = "16:00:00"
    end_day    = "saturday"
    end_time   = "00:00:00"
  }

  # [PagerDuty] Platform On-Call https://example.pagerduty.com/teams/T1FTS
}

resource "firehydrant_rotation" "fh_team_T1FTS_follow_the_sun_primary_weekend" {
  name        = "Weekend"
  description = "(Weekend)"
  team_id     = firehydrant_team.fh_team_T1FTS.id
  schedule_id = firehydrant_on_call_schedule.fh_team_T1FTS_follow_the_sun_primary.id
  time_zone   = "Etc/UTC"
  start_time  = "2025-06-07T00:00:00Z"

  members {
    user_id = data.firehydrant_user.morgan.id
  }
  members {
    user_id = data.firehydrant_user.noor.id
  }

  strategy {
    type         = "weekly"
    handoff_day  = "saturday"
    handoff_time = "00:00:00"
  }

  restrictions {
    start_day  = "saturday"
    start_time = "00:00:00"
    end_day    = "sunday"
    end_time   = "00:00:00"
  }

  restrictions {
    start_day  = "sunday"
    start_time = "00:00:00"
    end_day    = "monday"
    end_time   = "00:00:00"
  }

  # [PagerDuty] Platform On-Call https://example.pagerduty.com/teams/T1FTS
}
