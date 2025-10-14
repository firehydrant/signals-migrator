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
  # [Opsgenie] 1dc37638-ab52-44f3-848e-a16bcc584fb7 admin@example.io
}

data "firehydrant_user" "mika" {
  email = "mika@example.com"
  # [Opsgenie] 9253cf00-6195-4123-a9a6-f9f1e25718d8 mika@example.io
}

data "firehydrant_user" "kiran" {
  email = "kiran@example.com"
  # [Opsgenie] e94e17aa-418c-44f7-8e47-1eaebf6b5343 kiran@example.com
}

data "firehydrant_user" "jack" {
  email = "jack@example.com"
  # [Opsgenie] d68757c4-5eec-4560-8c5b-91c463f87dd8 jack@example.com
}

data "firehydrant_user" "wong" {
  email = "wong@example.com"
  # [Opsgenie] a13020ca-cb08-48e3-9403-bed181a22e72 wong@example.io
}

resource "firehydrant_team" "wong_squad" {
  name = "Wong Squad"
}

resource "firehydrant_team" "platform" {
  name = "Platform"
}

resource "firehydrant_team" "aj_team" {
  name = "AJ Team"

  memberships {
    user_id = data.firehydrant_user.kiran.id
  }

  memberships {
    user_id = data.firehydrant_user.mika.id
  }

  memberships {
    user_id = data.firehydrant_user.local.id
  }
}

resource "firehydrant_team" "customer_success" {
  name = "Customer Success"
}

resource "firehydrant_team" "noodlebrigade" {
  name = "noodlebrigade"

  memberships {
    user_id = data.firehydrant_user.jack.id
  }
}

resource "firehydrant_team" "christine_test_team" {
  name = "Christine Test Team"
}

resource "firehydrant_on_call_schedule" "customer_success_customer_success_schedule" {
  name        = "Customer Success_schedule"
  description = "Customer Success team schedule"
  team_id     = firehydrant_team.customer_success.id
  time_zone   = "America/Los_Angeles"

  member_ids = []

  strategy {
    type         = "weekly"
    handoff_day  = "tuesday"
    handoff_time = "04:45:32"
  }
}

resource "firehydrant_on_call_schedule" "wong_squad_wong_team_schedule" {
  name        = "Wong Team_schedule"
  description = "Wong team schedule"
  team_id     = firehydrant_team.wong_squad.id
  time_zone   = "America/Los_Angeles"

  member_ids = []

  strategy {
    type         = "daily"
    handoff_time = "08:00:00"
  }
}

resource "firehydrant_on_call_schedule" "aj_team_aj_team_schedule" {
  name        = "AJ Team_schedule"
  description = "AJ team schedule"
  team_id     = firehydrant_team.aj_team.id
  time_zone   = "America/Los_Angeles"

  member_ids = [data.firehydrant_user.mika.id, data.firehydrant_user.local.id, data.firehydrant_user.kiran.id]

  strategy {
    type         = "weekly"
    handoff_day  = "monday"
    handoff_time = "07:00:00"
  }

  restrictions {
    start_day  = "monday"
    start_time = "08:00:00"
    end_day    = "monday"
    end_time   = "18:00:00"
  }

  restrictions {
    start_day  = "tuesday"
    start_time = "08:00:00"
    end_day    = "tuesday"
    end_time   = "18:00:00"
  }

  restrictions {
    start_day  = "wednesday"
    start_time = "08:00:00"
    end_day    = "wednesday"
    end_time   = "18:00:00"
  }

  restrictions {
    start_day  = "thursday"
    start_time = "08:00:00"
    end_day    = "thursday"
    end_time   = "18:00:00"
  }

  restrictions {
    start_day  = "friday"
    start_time = "08:00:00"
    end_day    = "friday"
    end_time   = "18:00:00"
  }

  # Overrides found for this schedule:
  # User: admin@example.net		Starting: Tues, 11 Oct 3025 18:30:00 +0000		Ending: Wed, 12 Oct 3025 18:30:00 +0000
  # You can see documention for adding overrides here: https://docs.firehydrant.com/docs/signals-on-call-schedules#overrides
}

resource "firehydrant_rotation" "aj_team_aj_team_schedule_nighttime_rotation" {
  name        = "Nighttime rotation"
  description = "(Nighttime rotation)"
  team_id     = firehydrant_team.aj_team.id
  schedule_id = firehydrant_on_call_schedule.aj_team_aj_team_schedule.id
  time_zone   = "America/Los_Angeles"

  members = [data.firehydrant_user.mika.id, data.firehydrant_user.local.id, data.firehydrant_user.kiran.id]

  strategy {
    type         = "daily"
    handoff_time = "15:00:00"
  }

  restrictions {
    start_day  = "sunday"
    start_time = "18:00:00"
    end_day    = "monday"
    end_time   = "08:00:00"
  }

  restrictions {
    start_day  = "monday"
    start_time = "18:00:00"
    end_day    = "tuesday"
    end_time   = "08:00:00"
  }

  restrictions {
    start_day  = "tuesday"
    start_time = "18:00:00"
    end_day    = "wednesday"
    end_time   = "08:00:00"
  }

  restrictions {
    start_day  = "wednesday"
    start_time = "18:00:00"
    end_day    = "thursday"
    end_time   = "08:00:00"
  }

  restrictions {
    start_day  = "thursday"
    start_time = "18:00:00"
    end_day    = "friday"
    end_time   = "08:00:00"
  }

  restrictions {
    start_day  = "friday"
    start_time = "18:00:00"
    end_day    = "saturday"
    end_time   = "08:00:00"
  }

  restrictions {
    start_day  = "saturday"
    start_time = "18:00:00"
    end_day    = "sunday"
    end_time   = "08:00:00"
  }
}
