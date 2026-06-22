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
  # [PagerDuty] avery@example.com https://example.pagerduty.com/users/U2AVERY
}

data "firehydrant_user" "blake" {
  email = "blake@example.com"
  # [PagerDuty] blake@example.com https://example.pagerduty.com/users/U2BLAKE
}

data "firehydrant_user" "cam" {
  email = "cam@example.com"
  # [PagerDuty] cam@example.com https://example.pagerduty.com/users/U2CAM
}

data "firehydrant_user" "drew" {
  email = "drew@example.com"
  # [PagerDuty] drew@example.com https://example.pagerduty.com/users/U2DREW
}

resource "firehydrant_team" "fh_team_T2CAD" {
  name = "Cadence Squad"

  # [PagerDuty] Cadence Squad https://example.pagerduty.com/teams/T2CAD

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
}

import {
  id = "fh-team-T2CAD"
  to = firehydrant_team.fh_team_T2CAD
}

resource "firehydrant_on_call_schedule" "fh_team_T2CAD_two_week_cadence" {
  name                 = "Two-Week Cadence"
  description          = "Custom 14-day cadence with overnight coverage."
  team_id              = firehydrant_team.fh_team_T2CAD.id
  rotation_name        = "Biweekly Overnight"
  rotation_description = "Custom 14-day cadence with overnight coverage. (Biweekly Overnight)"
  time_zone            = "America/New_York"
  start_time           = "2024-03-15T22:00:00-04:00"

  member_ids = [data.firehydrant_user.avery.id, data.firehydrant_user.blake.id, data.firehydrant_user.cam.id, data.firehydrant_user.drew.id]

  strategy {
    type           = "custom"
    shift_duration = "PT1209600S"
  }

  restrictions {
    start_day  = "sunday"
    start_time = "22:00:00"
    end_day    = "monday"
    end_time   = "07:00:00"
  }

  restrictions {
    start_day  = "monday"
    start_time = "22:00:00"
    end_day    = "tuesday"
    end_time   = "07:00:00"
  }

  restrictions {
    start_day  = "tuesday"
    start_time = "22:00:00"
    end_day    = "wednesday"
    end_time   = "07:00:00"
  }

  restrictions {
    start_day  = "wednesday"
    start_time = "22:00:00"
    end_day    = "thursday"
    end_time   = "07:00:00"
  }

  restrictions {
    start_day  = "thursday"
    start_time = "22:00:00"
    end_day    = "friday"
    end_time   = "07:00:00"
  }

  restrictions {
    start_day  = "friday"
    start_time = "22:00:00"
    end_day    = "saturday"
    end_time   = "07:00:00"
  }

  restrictions {
    start_day  = "saturday"
    start_time = "22:00:00"
    end_day    = "sunday"
    end_time   = "07:00:00"
  }

  # [PagerDuty] Cadence Squad https://example.pagerduty.com/teams/T2CAD
}
