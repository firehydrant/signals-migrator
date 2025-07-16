terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant-v2"
      version = "~> 0.3.0"
    }
  }
}

data "firehydrant_user" "alice_bob" {
  id = "35b5390f-d134-4bc6-966d-0b4048788b62"
  # Insert PagerDuty URL here :)
  # https://acme-eng.pagerduty.com/users/PUIDISU
}

resource "firehydrant_team" "cowboy_coders" {
  name = "🐴 Cowboy Coders"

  # [PagerDuty] team-rocket https://pdt-apidocs.pagerduty.com/service-directory/PV9JOXL

  memberships_input = [{ user_id = data.firehydrant_user.alice_bob.id }]
}

import {
  id = "f159b173-1ffd-41ac-9254-ce8ec1142267"
  to = firehydrant_team.cowboy_coders
}

resource "firehydrant_signals_api_on_call_schedule" "cowboy_coders_atalice_bob_is_always_on_call_layer_1" {
  name        = "🐴 @alice.bob is always on call - Layer 1"
  description = "Always 😭 (Layer 1)"
  team_id     = firehydrant_team.cowboy_coders.id
  time_zone   = "America/Los_Angeles"

  # [PagerDuty] team-rocket https://pdt-apidocs.pagerduty.com/service-directory/PV9JOXL

  members_input = [{ user_id = data.firehydrant_user.alice_bob.id }]

  strategy_input = {
    handoff_day  = "friday"
    handoff_time = "12:00:00"
    type         = "weekly"
  }
}

resource "firehydrant_signals_api_escalation_policy" "atalice_bob_test_service_ep" {
  name = "🐴 @alice.bob Test Service-ep"

  steps = [{ timeout = "PT30M", targets = [] }]

  repetitions = 0
}

resource "firehydrant_signals_api_escalation_policy" "notify_atalice_bob" {
  name = "🐴 Notify @alice.bob"

  steps = [{ timeout = "PT30M", targets = [] }]

  repetitions = 0
}
