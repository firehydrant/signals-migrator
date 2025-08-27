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
