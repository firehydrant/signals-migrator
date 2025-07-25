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
  time_zone   = "America/New_York"

  members_input = [{ user_id = data.firehydrant_user.user_0.id }, { user_id = data.firehydrant_user.user_1.id }]

  strategy_input = {
    handoff_day  = "monday"
    handoff_time = "09:00"
    type         = "weekly"
  }

  restrictions_input = [{
    end_day    = "friday"
    end_time   = "17:00"
    start_day  = "monday"
    start_time = "09:00"
    }, {
    end_day    = "sunday"
    end_time   = "16:00"
    start_day  = "saturday"
    start_time = "10:00"
  }]
}
