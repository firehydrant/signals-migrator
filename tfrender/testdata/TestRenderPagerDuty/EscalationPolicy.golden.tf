terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant"
      version = ">= 0.8.0"
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

  memberships {
    user_id = data.firehydrant_user.alice_bob.id
  }
}

import {
  id = "f159b173-1ffd-41ac-9254-ce8ec1142267"
  to = firehydrant_team.cowboy_coders
}

resource "firehydrant_on_call_schedule" "cowboy_coders_atalice_bob_is_always_on_call_layer_1" {
  name        = "🐴 @alice.bob is always on call - Layer 1"
  description = "Always 😭 (Layer 1)"
  team_id     = firehydrant_team.cowboy_coders.id
  time_zone   = "America/Los_Angeles"

  member_ids = [data.firehydrant_user.alice_bob.id]

  strategy {
    type         = "weekly"
    handoff_day  = "friday"
    handoff_time = "12:00:00"
  }
}

resource "firehydrant_escalation_policy" "atalice_bob_test_service_ep" {
  name = "🐴 @alice.bob Test Service-ep"

  step {
    timeout = "PT30M"

    targets {
      type = "OnCallSchedule"
      id   = firehydrant_on_call_schedule.atalice_bob_is_always_on_call_layer_1.id
    }
  }

  repetitions = 0
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
}
