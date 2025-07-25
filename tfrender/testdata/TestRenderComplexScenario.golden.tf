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

data "firehydrant_user" "user_2" {
  id = "id-for-user-2"
}

data "firehydrant_user" "user_3" {
  id = "id-for-user-3"
}

data "firehydrant_user" "user_4" {
  id = "id-for-user-4"
}

resource "firehydrant_team" "team_0_slug" {
  name = "Team 0"

  memberships_input = [{ user_id = data.firehydrant_user.user_0.id }, { user_id = data.firehydrant_user.user_2.id }]
}

import {
  id = "id-for-team-0"
  to = firehydrant_team.team_0_slug
}

resource "firehydrant_team" "team_1_slug" {
  name = "Team 1"

  memberships_input = [{ user_id = data.firehydrant_user.user_1.id }]
}

resource "firehydrant_team" "team_2_slug" {
  name = "Team 2"
}

resource "firehydrant_signals_api_on_call_schedule" "team_0_slug_schedule_0" {
  name        = "Schedule 0"
  description = "Schedule 0 description"
  team_id     = firehydrant_team.team_0_slug.id
  time_zone   = "UTC"

  members_input = [{ user_id = data.firehydrant_user.user_0.id }, { user_id = data.firehydrant_user.user_1.id }]

  strategy_input = {
    handoff_day  = "monday"
    handoff_time = "09:00"
    type         = "weekly"
  }
}

resource "firehydrant_signals_api_on_call_schedule" "team_1_slug_schedule_1" {
  name        = "Schedule 1"
  description = "Schedule 1 description"
  team_id     = firehydrant_team.team_1_slug.id
  time_zone   = "UTC"

  members_input = [{ user_id = data.firehydrant_user.user_2.id }, { user_id = data.firehydrant_user.user_3.id }]

  strategy_input = {
    handoff_day  = "monday"
    handoff_time = "09:00"
    type         = "weekly"
  }
}
