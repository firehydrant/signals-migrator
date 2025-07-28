terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant-v2"
      version = "~> 0.3.1"
    }
  }
}

data "firehydrant_user" "local" {
  id = "0946be55-ea20-4483-b9ab-617d5f0969e2"
}

data "firehydrant_user" "mika" {
  id = "e6009411-0015-43e3-815e-ca9db72f4088"
}

data "firehydrant_user" "kiran" {
  id = "4c3f28fa-b402-453c-9652-f014ecbe65a9"
}

data "firehydrant_user" "horse" {
  id = "35b5390f-d134-4bc6-966d-0b4048788b62"
}

data "firehydrant_user" "jack" {
  id = "6c08bff2-98f6-4ee9-8de1-12202186d084"
}

data "firehydrant_user" "wong" {
  id = "032a1f07-987e-4f76-8273-136e08e50baa"
}

resource "firehydrant_team" "aaaa_ipv6_migration_strategy" {
  name = "AAAA IPv6 migration strategy"

  # [PagerDuty] Jen https://pdt-apidocs.pagerduty.com/service-directory/PT54U20

  memberships_input = [{ user_id = data.firehydrant_user.local.id }, { user_id = data.firehydrant_user.kiran.id }, { user_id = data.firehydrant_user.wong.id }]
}

import {
  id = "47016143-6547-483a-b68a-5220b21681fd"
  to = firehydrant_team.aaaa_ipv6_migration_strategy
}

resource "firehydrant_team" "dunder_mifflin_scranton" {
  name = "Dunder Mifflin Scranton"

  # [PagerDuty] Jack Team https://pdt-apidocs.pagerduty.com/service-directory/PD2F80U

  memberships_input = [{ user_id = data.firehydrant_user.jack.id }]
}

import {
  id = "97d539b0-47a5-44f6-81e6-b6fcd98f23ac"
  to = firehydrant_team.dunder_mifflin_scranton
}

resource "firehydrant_signals_api_on_call_schedule" "aaaa_ipv6_migration_strategy_jen_primary_layer_2" {
  name        = "Jen - primary - Layer 2"
  description = "(Layer 2)"
  team_id     = firehydrant_team.aaaa_ipv6_migration_strategy.id
  time_zone   = "America/Los_Angeles"
  start_time  = "2024-04-10T20:39:29-07:00"

  # [PagerDuty] Jen https://pdt-apidocs.pagerduty.com/service-directory/PT54U20

  members_input = [{ user_id = data.firehydrant_user.kiran.id }]

  strategy_input = {
    shift_duration = "PT93600S"
    type           = "custom"
  }

  restrictions_input = [{
    end_day    = "friday"
    end_time   = "17:00:00"
    start_day  = "monday"
    start_time = "09:00:00"
  }]
}

resource "firehydrant_signals_api_on_call_schedule" "aaaa_ipv6_migration_strategy_jen_primary_layer_1" {
  name        = "Jen - primary - Layer 1"
  description = "(Layer 1)"
  team_id     = firehydrant_team.aaaa_ipv6_migration_strategy.id
  time_zone   = "America/Los_Angeles"
  start_time  = "2024-04-10T20:39:29-07:00"

  # [PagerDuty] Jen https://pdt-apidocs.pagerduty.com/service-directory/PT54U20

  members_input = [{ user_id = data.firehydrant_user.local.id }, { user_id = data.firehydrant_user.wong.id }]

  strategy_input = {
    shift_duration = "PT7200S"
    type           = "custom"
  }

  restrictions_input = [{
    end_day    = "sunday"
    end_time   = "17:00:00"
    start_day  = "sunday"
    start_time = "09:00:00"
    }, {
    end_day    = "monday"
    end_time   = "17:00:00"
    start_day  = "monday"
    start_time = "09:00:00"
    }, {
    end_day    = "tuesday"
    end_time   = "17:00:00"
    start_day  = "tuesday"
    start_time = "09:00:00"
    }, {
    end_day    = "wednesday"
    end_time   = "17:00:00"
    start_day  = "wednesday"
    start_time = "09:00:00"
    }, {
    end_day    = "thursday"
    end_time   = "17:00:00"
    start_day  = "thursday"
    start_time = "09:00:00"
    }, {
    end_day    = "friday"
    end_time   = "17:00:00"
    start_day  = "friday"
    start_time = "09:00:00"
    }, {
    end_day    = "saturday"
    end_time   = "17:00:00"
    start_day  = "saturday"
    start_time = "09:00:00"
  }]
}

resource "firehydrant_signals_api_on_call_schedule" "dunder_mifflin_scranton_jack_on_call_schedule_layer_1" {
  name        = "Jack On-Call Schedule - Layer 1"
  description = " (Layer 1)"
  team_id     = firehydrant_team.dunder_mifflin_scranton.id
  time_zone   = "America/Los_Angeles"

  # [PagerDuty] Jack Team https://pdt-apidocs.pagerduty.com/service-directory/PD2F80U

  members_input = [{ user_id = data.firehydrant_user.jack.id }]

  strategy_input = {
    handoff_day  = "friday"
    handoff_time = "14:00:00"
    type         = "weekly"
  }
}
