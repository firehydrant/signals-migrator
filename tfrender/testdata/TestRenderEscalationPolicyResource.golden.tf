terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant-v2"
      version = "~> 0.3.0"
    }
  }
}

data "firehydrant_user" "user_0" {
  id = "id-for-user-0"
}

data "firehydrant_user" "user_1" {
  id = "id-for-user-1"
}

resource "firehydrant_team" "team_0_slug" {
  name = "Team 0"
}

resource "firehydrant_team" "team_1_slug" {
  name = "Team 1"
}

resource "firehydrant_signals_api_on_call_schedule" "team_0_slug_schedule_0" {
  name        = "Schedule 0"
  description = "Schedule 0 description"
  team_id     = firehydrant_team.team_0_slug.id
  time_zone   = "UTC"

  members_input = []

  strategy_input = {
    handoff_day  = "monday"
    handoff_time = "09:00"
    type         = "weekly"
  }
}

resource "firehydrant_signals_api_escalation_policy" "policy_0" {
  name        = "Policy 0"
  description = "Policy 0 description"
  team_id     = firehydrant_team.team_0_slug.id

  steps = [{ timeout = "PT5M", targets = [] }]

  repetitions = 3
}
