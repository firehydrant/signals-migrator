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
  # [PagerDuty] avery@example.com https://example.pagerduty.com/users/U4AVERY
}

data "firehydrant_user" "blake" {
  email = "blake@example.com"
  # [PagerDuty] blake@example.com https://example.pagerduty.com/users/U4BLAKE
}

data "firehydrant_user" "cam" {
  email = "cam@example.com"
  # [PagerDuty] cam@example.com https://example.pagerduty.com/users/U4CAM
}

data "firehydrant_user" "drew" {
  email = "drew@example.com"
  # [PagerDuty] drew@example.com https://example.pagerduty.com/users/U4DREW
}

data "firehydrant_user" "evan" {
  email = "evan@example.com"
  # [PagerDuty] evan@example.com https://example.pagerduty.com/users/U4EVAN
}

resource "firehydrant_team" "fh_team_T4MIX" {
  name = "Mixed Strategies"

  # [PagerDuty] Mixed Strategies https://example.pagerduty.com/teams/T4MIX

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
}

import {
  id = "fh-team-T4MIX"
  to = firehydrant_team.fh_team_T4MIX
}

resource "firehydrant_on_call_schedule" "fh_team_T4MIX_mixed_strategies_schedule" {
  name                 = "Mixed Strategies Schedule"
  team_id              = firehydrant_team.fh_team_T4MIX.id
  rotation_name        = "Primary Business Hours"
  rotation_description = "(Primary Business Hours)"
  time_zone            = "Etc/UTC"
  start_time           = "2025-01-06T09:00:00Z"

  member_ids = [data.firehydrant_user.avery.id, data.firehydrant_user.blake.id, data.firehydrant_user.cam.id, data.firehydrant_user.drew.id, data.firehydrant_user.evan.id]

  strategy {
    type         = "weekly"
    handoff_day  = "monday"
    handoff_time = "09:00:00"
  }

  restrictions {
    start_day  = "monday"
    start_time = "09:00:00"
    end_day    = "monday"
    end_time   = "17:00:00"
  }

  restrictions {
    start_day  = "tuesday"
    start_time = "09:00:00"
    end_day    = "tuesday"
    end_time   = "17:00:00"
  }

  restrictions {
    start_day  = "wednesday"
    start_time = "09:00:00"
    end_day    = "wednesday"
    end_time   = "17:00:00"
  }

  restrictions {
    start_day  = "thursday"
    start_time = "09:00:00"
    end_day    = "thursday"
    end_time   = "17:00:00"
  }

  restrictions {
    start_day  = "friday"
    start_time = "09:00:00"
    end_day    = "friday"
    end_time   = "17:00:00"
  }

  # [PagerDuty] Mixed Strategies https://example.pagerduty.com/teams/T4MIX
}

resource "firehydrant_rotation" "fh_team_T4MIX_mixed_strategies_schedule_after_hours" {
  name        = "After Hours"
  description = "(After Hours)"
  team_id     = firehydrant_team.fh_team_T4MIX.id
  schedule_id = firehydrant_on_call_schedule.fh_team_T4MIX_mixed_strategies_schedule.id
  time_zone   = "Etc/UTC"
  start_time  = "2025-01-06T17:00:00Z"

  members {
    user_id = data.firehydrant_user.cam.id
  }
  members {
    user_id = data.firehydrant_user.drew.id
  }
  members {
    user_id = data.firehydrant_user.evan.id
  }

  strategy {
    type         = "daily"
    handoff_day  = "monday"
    handoff_time = "17:00:00"
  }

  restrictions {
    start_day  = "sunday"
    start_time = "17:00:00"
    end_day    = "monday"
    end_time   = "07:00:00"
  }

  restrictions {
    start_day  = "monday"
    start_time = "17:00:00"
    end_day    = "tuesday"
    end_time   = "07:00:00"
  }

  restrictions {
    start_day  = "tuesday"
    start_time = "17:00:00"
    end_day    = "wednesday"
    end_time   = "07:00:00"
  }

  restrictions {
    start_day  = "wednesday"
    start_time = "17:00:00"
    end_day    = "thursday"
    end_time   = "07:00:00"
  }

  restrictions {
    start_day  = "thursday"
    start_time = "17:00:00"
    end_day    = "friday"
    end_time   = "07:00:00"
  }

  restrictions {
    start_day  = "friday"
    start_time = "17:00:00"
    end_day    = "saturday"
    end_time   = "07:00:00"
  }

  restrictions {
    start_day  = "saturday"
    start_time = "17:00:00"
    end_day    = "sunday"
    end_time   = "07:00:00"
  }

  # [PagerDuty] Mixed Strategies https://example.pagerduty.com/teams/T4MIX
}
