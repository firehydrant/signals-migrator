terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant"
      version = ">= 0.14.7"
    }
  }
}

data "firehydrant_user" "local" {
  email = "local@example.io"
}

data "firehydrant_user" "mika" {
  email = "mika@example.com"
}

data "firehydrant_user" "kiran" {
  email = "kiran@example.com"
}

data "firehydrant_user" "horse" {
  email = "horse@example.com"
}

data "firehydrant_user" "jack" {
  email = "jack@example.com"
}

data "firehydrant_user" "wong" {
  email = "wong@example.com"
}

resource "firehydrant_team" "aaaa_ipv6_migration_strategy" {
  name = "AAAA IPv6 migration strategy"

  # [PagerDuty] Jen https://pdt-apidocs.pagerduty.com/service-directory/PT54U20

  memberships {
    user_id = data.firehydrant_user.local.id
  }

  memberships {
    user_id = data.firehydrant_user.kiran.id
  }

  memberships {
    user_id = data.firehydrant_user.wong.id
  }
}

import {
  id = "47016143-6547-483a-b68a-5220b21681fd"
  to = firehydrant_team.aaaa_ipv6_migration_strategy
}

resource "firehydrant_team" "dunder_mifflin_scranton" {
  name = "Dunder Mifflin Scranton"

  # [PagerDuty] Jack Team https://pdt-apidocs.pagerduty.com/service-directory/PD2F80U

  memberships {
    user_id = data.firehydrant_user.jack.id
  }
}

import {
  id = "97d539b0-47a5-44f6-81e6-b6fcd98f23ac"
  to = firehydrant_team.dunder_mifflin_scranton
}

resource "firehydrant_team" "cowboy_coders" {
  name = "🐴 Cowboy Coders"

  # [PagerDuty] 🐴 Cowboy Coders https://pdt-apidocs.pagerduty.com/service-directory/PV9JOXL
}

import {
  id = "f159b173-1ffd-41ac-9254-ce8ec1142267"
  to = firehydrant_team.cowboy_coders
}

resource "firehydrant_on_call_schedule" "jen_jen___primary" {
  name        = "Jen - primary"
  description = "Primary on-call schedule for Jen team"
  team_id     = firehydrant_team.jen.id
  time_zone   = "America/Los_Angeles"
  start_time  = "2024-04-10T20:39:29-07:00"

  member_ids = [data.firehydrant_user.kiran.id]

  strategy {
    type           = "custom"
    shift_duration = "PT93600S"
  }

  restrictions {
    start_day  = "monday"
    start_time = "09:00:00"
    end_day    = "friday"
    end_time   = "17:00:00"
  }

  # [PagerDuty] Jen https://pdt-apidocs.pagerduty.com/service-directory/PT54U20
}

resource "firehydrant_on_call_schedule" "jack_team_jack_on_call_schedule" {
  name        = "Jack On-Call Schedule"
  description = "On-call schedule for Jack team"
  team_id     = firehydrant_team.jack_team.id
  time_zone   = "America/Los_Angeles"

  member_ids = [data.firehydrant_user.jack.id]

  strategy {
    type         = "weekly"
    handoff_day  = "friday"
    handoff_time = "14:00:00"
  }

  # [PagerDuty] Jack Team https://pdt-apidocs.pagerduty.com/service-directory/PD2F80U
}

resource "firehydrant_on_call_schedule" "cowboy_coders_🐴_is_always_on_call" {
  name        = "🐴 is always on call"
  description = "Always on call schedule"
  team_id     = firehydrant_team.cowboy_coders.id
  time_zone   = "America/Los_Angeles"

  member_ids = [data.firehydrant_user.horse.id]

  strategy {
    type         = "weekly"
    handoff_day  = "friday"
    handoff_time = "12:00:00"
  }

  # [PagerDuty] 🐴 Cowboy Coders https://pdt-apidocs.pagerduty.com/service-directory/PV9JOXL
}

resource "firehydrant_rotation" "jen_jen___primary_layer_1" {
  name        = "Layer 1"
  description = "(Layer 1)"
  team_id     = firehydrant_team.jen.id
  schedule_id = firehydrant_on_call_schedule.jen_jen___primary.id
  time_zone   = "America/Los_Angeles"
  start_time  = "2024-04-10T20:39:29-07:00"

  members = [data.firehydrant_user.local.id, data.firehydrant_user.wong.id]

  strategy {
    type           = "custom"
    shift_duration = "PT7200S"
  }

  restrictions {
    start_day  = "sunday"
    start_time = "09:00:00"
    end_day    = "sunday"
    end_time   = "17:00:00"
  }

  restrictions {
    start_day  = "monday"
    start_time = "09:00:00"
    end_day    = "monday"
    end_time   = "17:00:00"
  }

  restrictions {
    start_day  = "tuesday"
    start_time = "09:00:00"
    end_day    = "tuesday"
    end_time   = "17:00:00"
  }

  restrictions {
    start_day  = "wednesday"
    start_time = "09:00:00"
    end_day    = "wednesday"
    end_time   = "17:00:00"
  }

  restrictions {
    start_day  = "thursday"
    start_time = "09:00:00"
    end_day    = "thursday"
    end_time   = "17:00:00"
  }

  restrictions {
    start_day  = "friday"
    start_time = "09:00:00"
    end_day    = "friday"
    end_time   = "17:00:00"
  }

  restrictions {
    start_day  = "saturday"
    start_time = "09:00:00"
    end_day    = "saturday"
    end_time   = "17:00:00"
  }

  # [PagerDuty] Jen https://pdt-apidocs.pagerduty.com/service-directory/PT54U20
}
