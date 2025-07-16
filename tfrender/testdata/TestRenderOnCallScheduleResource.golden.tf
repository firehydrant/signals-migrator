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

resource "firehydrant_team" "team_0_slug" {
  name = "Team 0"
}

import {
  id = "id-for-team-0"
  to = firehydrant_team.team_0_slug
}

resource "firehydrant_team" "team_1_slug" {
  name = "Team 1"
}

resource "firehydrant_team" "team_2_slug" {
  name = "Team 2"
}

import {
  id = "id-for-team-2"
  to = firehydrant_team.team_2_slug
}

resource "firehydrant_team" "team_3_slug" {
  name = "Team 3"
}

resource "firehydrant_signals_api_on_call_schedule" "team_1_slug_schedule_0" {
  name        = "Schedule 0"
  description = "Schedule 0 description"
  team_id     = firehydrant_team.team_1_slug.id
  time_zone   = "UTC"

  members_input = [{ user_id = data.firehydrant_user.user_0.id }]

  strategy_input = {
    handoff_time = "11:00"
    type         = "daily"
  }
}
