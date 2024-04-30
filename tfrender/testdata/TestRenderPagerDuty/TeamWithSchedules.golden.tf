terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant"
      version = ">= 0.8.0"
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

resource "firehydrant_on_call_schedule" "aaaa_ipv6_migration_strategy_jen_primary_layer_2" {
  name        = "Jen - primary - Layer 2"
  description = "(Layer 2)"
  team_id     = firehydrant_team.aaaa_ipv6_migration_strategy.id
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
}

resource "firehydrant_on_call_schedule" "aaaa_ipv6_migration_strategy_jen_primary_layer_1" {
  name        = "Jen - primary - Layer 1"
  description = "(Layer 1)"
  team_id     = firehydrant_team.aaaa_ipv6_migration_strategy.id
  time_zone   = "America/Los_Angeles"
  start_time  = "2024-04-10T20:39:29-07:00"

  member_ids = [data.firehydrant_user.wong.id, data.firehydrant_user.local.id]

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
}

resource "firehydrant_on_call_schedule" "dunder_mifflin_scranton_jack_on_call_schedule_layer_1" {
  name        = "Jack On-Call Schedule - Layer 1"
  description = " (Layer 1)"
  team_id     = firehydrant_team.dunder_mifflin_scranton.id
  time_zone   = "America/Los_Angeles"

  member_ids = [data.firehydrant_user.jack.id]

  strategy {
    type         = "weekly"
    handoff_day  = "friday"
    handoff_time = "14:00:00"
  }
}
