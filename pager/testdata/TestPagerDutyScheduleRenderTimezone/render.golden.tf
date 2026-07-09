terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant"
      version = ">= 0.15.2"
    }
  }
}

data "firehydrant_user" "quinn" {
  email = "quinn@example.com"
  # [PagerDuty] quinn@example.com https://example.pagerduty.com/users/U6QUINN
}

resource "firehydrant_team" "fh_team_T6TZ" {
  name = "Coverage Team"

  # [PagerDuty] Coverage Team https://example.pagerduty.com/teams/T6TZ

  memberships {
    user_id = data.firehydrant_user.quinn.id
  }
}

import {
  id = "fh-team-T6TZ"
  to = firehydrant_team.fh_team_T6TZ
}

resource "firehydrant_on_call_schedule" "fh_team_T6TZ_daytime_coverage" {
  name                 = "Daytime Coverage"
  description          = "Schedule whose list response is rendered in the API user's time zone"
  team_id              = firehydrant_team.fh_team_T6TZ.id
  rotation_name        = "Daytime Layer"
  rotation_description = "Schedule whose list response is rendered in the API user's time zone (Daytime Layer)"
  time_zone            = "America/Chicago"
  start_time           = "2026-01-09T14:00:00-06:00"

  member_ids = [data.firehydrant_user.quinn.id]

  strategy {
    type         = "daily"
    handoff_day  = "friday"
    handoff_time = "14:00:00"
  }

  restrictions {
    start_day  = "sunday"
    start_time = "08:00:00"
    end_day    = "sunday"
    end_time   = "20:00:00"
  }

  restrictions {
    start_day  = "monday"
    start_time = "08:00:00"
    end_day    = "monday"
    end_time   = "20:00:00"
  }

  restrictions {
    start_day  = "tuesday"
    start_time = "08:00:00"
    end_day    = "tuesday"
    end_time   = "20:00:00"
  }

  restrictions {
    start_day  = "wednesday"
    start_time = "08:00:00"
    end_day    = "wednesday"
    end_time   = "20:00:00"
  }

  restrictions {
    start_day  = "thursday"
    start_time = "08:00:00"
    end_day    = "thursday"
    end_time   = "20:00:00"
  }

  restrictions {
    start_day  = "friday"
    start_time = "08:00:00"
    end_day    = "friday"
    end_time   = "20:00:00"
  }

  restrictions {
    start_day  = "saturday"
    start_time = "08:00:00"
    end_day    = "saturday"
    end_time   = "20:00:00"
  }

  # [PagerDuty] Coverage Team https://example.pagerduty.com/teams/T6TZ
}
