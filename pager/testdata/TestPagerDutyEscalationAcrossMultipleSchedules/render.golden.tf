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
  # [PagerDuty] avery@example.com https://example.pagerduty.com/users/U5AVERY
}

data "firehydrant_user" "blake" {
  email = "blake@example.com"
  # [PagerDuty] blake@example.com https://example.pagerduty.com/users/U5BLAKE
}

data "firehydrant_user" "cam" {
  email = "cam@example.com"
  # [PagerDuty] cam@example.com https://example.pagerduty.com/users/U5CAM
}

data "firehydrant_user" "drew" {
  email = "drew@example.com"
  # [PagerDuty] drew@example.com https://example.pagerduty.com/users/U5DREW
}

data "firehydrant_user" "evan" {
  email = "evan@example.com"
  # [PagerDuty] evan@example.com https://example.pagerduty.com/users/U5EVAN
}

data "firehydrant_user" "frey" {
  email = "frey@example.com"
  # [PagerDuty] frey@example.com https://example.pagerduty.com/users/U5FREY
}

data "firehydrant_user" "gale" {
  email = "gale@example.com"
  # [PagerDuty] gale@example.com https://example.pagerduty.com/users/U5GALE
}

data "firehydrant_user" "harlan" {
  email = "harlan@example.com"
  # [PagerDuty] harlan@example.com https://example.pagerduty.com/users/U5HARLAN
}

resource "firehydrant_team" "fh_team_T5EP" {
  name = "Escalation Quad"

  # [PagerDuty] Escalation Quad https://example.pagerduty.com/teams/T5EP

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
    user_id = data.firehydrant_user.frey.id
  }

  memberships {
    user_id = data.firehydrant_user.gale.id
  }

  memberships {
    user_id = data.firehydrant_user.harlan.id
  }
}

import {
  id = "fh-team-T5EP"
  to = firehydrant_team.fh_team_T5EP
}

resource "firehydrant_on_call_schedule" "fh_team_T5EP_primary_on_call" {
  name                 = "Primary On-Call"
  team_id              = firehydrant_team.fh_team_T5EP.id
  rotation_name        = "Primary Weekly"
  rotation_description = "(Primary Weekly)"
  time_zone            = "Etc/UTC"
  start_time           = "2025-06-02T09:00:00Z"

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

  # [PagerDuty] Escalation Quad https://example.pagerduty.com/teams/T5EP
}

resource "firehydrant_on_call_schedule" "fh_team_T5EP_manager_escalation" {
  name                 = "Manager Escalation"
  description          = "Monthly manager rotation."
  team_id              = firehydrant_team.fh_team_T5EP.id
  rotation_name        = "Manager Monthly"
  rotation_description = "Monthly manager rotation. (Manager Monthly)"
  time_zone            = "Etc/UTC"
  start_time           = "2025-06-01T00:00:00Z"

  member_ids = [data.firehydrant_user.avery.id, data.firehydrant_user.frey.id, data.firehydrant_user.gale.id]

  strategy {
    type           = "custom"
    shift_duration = "PT2592000S"
  }

  # [PagerDuty] Escalation Quad https://example.pagerduty.com/teams/T5EP
}

resource "firehydrant_escalation_policy" "tiered_response" {
  name        = "Tiered Response"
  description = "Three-tier escalation across schedules and a backstop user."
  team_id     = firehydrant_team.fh_team_T5EP.id

  # [PagerDuty]
  #   Tiered Response https://example.pagerduty.com/escalation_policies/EP5
  # [Teams]
  #   - T5EP

  step {
    timeout = "PT5M"

    targets {
      type = "OnCallSchedule"
      id   = firehydrant_on_call_schedule.fh_team_T5EP_primary_on_call.id
    }
  }

  step {
    timeout = "PT10M"

    targets {
      type = "OnCallSchedule"
      id   = firehydrant_on_call_schedule.fh_team_T5EP_manager_escalation.id
    }
  }

  step {
    timeout = "PT15M"

    targets {
      type = "User"
      id   = data.firehydrant_user.harlan.id
    }
  }

  repetitions = 0
  default     = "true"
}
