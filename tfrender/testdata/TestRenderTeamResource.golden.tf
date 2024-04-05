terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant"
      version = ">= 0.7.1"
    }
  }
}

data "firehydrant_user" "user_0" {
  email = "user-0@example.com"
}

data "firehydrant_user" "user_1" {
  email = "user-1@example.com"
}

data "firehydrant_user" "user_2" {
  email = "user-2@example.com"
}

data "firehydrant_user" "user_3" {
  email = "user-3@example.com"
}

resource "firehydrant_team" "team_0_slug" {
  name = "Team 0"

  membership {
    id = data.firehydrant_user.user_0.id
  }
}

import {
  id = "id-for-team-0"
  to = resource.firehydrant_team.team_0_slug.id
}

resource "firehydrant_team" "team_1_slug" {
  name = "Team 1"
}

resource "firehydrant_team" "team_2_slug" {
  name = "Team 2"
}

import {
  id = "id-for-team-2"
  to = resource.firehydrant_team.team_2_slug.id
}

resource "firehydrant_team" "team_3_slug" {
  name = "Team 3"
}
