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
  # [PagerDuty] avery@example.com https://example.pagerduty.com/users/U3AVERY
}

data "firehydrant_user" "blake" {
  email = "blake@example.com"
  # [PagerDuty] blake@example.com https://example.pagerduty.com/users/U3BLAKE
}

data "firehydrant_user" "cam" {
  email = "cam@example.com"
  # [PagerDuty] cam@example.com https://example.pagerduty.com/users/U3CAM
}

resource "firehydrant_team" "fh_team_T3WK" {
  name = "Weekly Coverage"

  # [PagerDuty] Weekly Coverage https://example.pagerduty.com/teams/T3WK

  memberships {
    user_id = data.firehydrant_user.avery.id
  }

  memberships {
    user_id = data.firehydrant_user.blake.id
  }

  memberships {
    user_id = data.firehydrant_user.cam.id
  }
}

import {
  id = "fh-team-T3WK"
  to = firehydrant_team.fh_team_T3WK
}

resource "firehydrant_on_call_schedule" "fh_team_T3WK_all_weekdays_coverage" {
  name                 = "All Weekdays Coverage"
  team_id              = firehydrant_team.fh_team_T3WK.id
  rotation_name        = "Business Hours"
  rotation_description = "(Business Hours)"
  time_zone            = "America/Los_Angeles"
  start_time           = "2026-01-04T09:00:00-08:00"

  member_ids = [data.firehydrant_user.avery.id, data.firehydrant_user.blake.id, data.firehydrant_user.cam.id]

  strategy {
    type         = "weekly"
    handoff_day  = "sunday"
    handoff_time = "09:00:00"
  }

  restrictions {
    start_day  = "sunday"
    start_time = "09:00:00"
    end_day    = "sunday"
    end_time   = "21:00:00"
  }

  restrictions {
    start_day  = "monday"
    start_time = "09:00:00"
    end_day    = "monday"
    end_time   = "21:00:00"
  }

  restrictions {
    start_day  = "tuesday"
    start_time = "09:00:00"
    end_day    = "tuesday"
    end_time   = "21:00:00"
  }

  restrictions {
    start_day  = "wednesday"
    start_time = "09:00:00"
    end_day    = "wednesday"
    end_time   = "21:00:00"
  }

  restrictions {
    start_day  = "thursday"
    start_time = "09:00:00"
    end_day    = "thursday"
    end_time   = "21:00:00"
  }

  restrictions {
    start_day  = "friday"
    start_time = "09:00:00"
    end_day    = "friday"
    end_time   = "21:00:00"
  }

  restrictions {
    start_day  = "saturday"
    start_time = "09:00:00"
    end_day    = "saturday"
    end_time   = "21:00:00"
  }

  # [PagerDuty] Weekly Coverage https://example.pagerduty.com/teams/T3WK
}
