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

  memberships_input = [{ user_id = data.firehydrant_user.user_0.id }]
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
