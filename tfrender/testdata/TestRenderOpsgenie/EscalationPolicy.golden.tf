terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant"
      version = ">= 0.8.0"
    }
  }
}

data "firehydrant_user" "jsmith" {
  email = "jsmith@example.com"
  # [Opsgenie] e0a51be7-3c7e-407f-8678-292ab421f55f jsmith@example.com
}

data "firehydrant_user" "fh_demo" {
  email = "fh-demo@example.com"
  # [Opsgenie] 1dc37638-ab52-44f3-848e-a16bcc584fb7 fh-demo@example.com
}

data "firehydrant_user" "fh_eng" {
  email = "fh-eng@example.com"
  # [Opsgenie] 9253cf00-6195-4123-a9a6-f9f1e25718d8 fh-eng@example.com
}

data "firehydrant_user" "fh_success" {
  email = "fh-success@example.com"
  # [Opsgenie] e94e17aa-418c-44f7-8e47-1eaebf6b5343 fh-success@example.com
}

resource "firehydrant_team" "aj_team" {
  name = "AJ Team"

  memberships {
    user_id = data.firehydrant_user.fh_eng.id
  }

  memberships {
    user_id = data.firehydrant_user.fh_demo.id
  }

  memberships {
    user_id = data.firehydrant_user.jsmith.id
  }

  memberships {
    user_id = data.firehydrant_user.fh_success.id
  }
}

resource "firehydrant_escalation_policy" "aj_team_escalation" {
  name    = "AJ Team_escalation"
  team_id = firehydrant_team.aj_team.id

  step {
    timeout = "PT1M"
  }

  repetitions = 0
}
