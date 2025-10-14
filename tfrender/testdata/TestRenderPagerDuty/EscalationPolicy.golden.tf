terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant"
      version = ">= 0.14.7"
    }
  }
}

data "firehydrant_user" "alice_bob" {
  email = "alice.bob@example.com"
  # Insert PagerDuty URL here :)
  # https://acme-eng.pagerduty.com/users/PUIDISU
}

resource "firehydrant_team" "cowboy_coders" {
  name = "🐴 Cowboy Coders"

  # [PagerDuty] team-rocket https://pdt-apidocs.pagerduty.com/service-directory/PV9JOXL

  memberships {
    user_id = data.firehydrant_user.alice_bob.id
  }
}

import {
  id = "f159b173-1ffd-41ac-9254-ce8ec1142267"
  to = firehydrant_team.cowboy_coders
}

resource "firehydrant_on_call_schedule" "cowboy_coders_🐴_@alice.bob_is_always_on_call" {
  name        = "🐴 @alice.bob is always on call"
  description = "Always on call schedule"
  team_id     = firehydrant_team.cowboy_coders.id
  time_zone   = "America/Los_Angeles"

  member_ids = [data.firehydrant_user.alice_bob.id]

  strategy {
    type         = "weekly"
    handoff_day  = "friday"
    handoff_time = "12:00:00"
  }

  # [PagerDuty] team-rocket https://pdt-apidocs.pagerduty.com/service-directory/PV9JOXL
}

resource "firehydrant_escalation_policy" "atalice_bob_test_service_ep" {
  name = "🐴 @alice.bob Test Service-ep"

  step {
    timeout = "PT30M"

    targets {
      type = "OnCallSchedule"
      id   = firehydrant_on_call_schedule.cowboy_coders_🐴_@alice.bob_is_always_on_call.id
    }
  }

  repetitions = 0
  default     = "false"
}

resource "firehydrant_escalation_policy" "notify_atalice_bob" {
  name = "🐴 Notify @alice.bob"

  step {
    timeout = "PT30M"

    targets {
      type = "User"
      id   = data.firehydrant_user.alice_bob.id
    }
  }

  repetitions = 0
  default     = "false"
}
