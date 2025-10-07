terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant"
      version = ">= 0.14.7"
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

resource "firehydrant_on_call_schedule" "aj_team_aj_team_schedule" {
  name        = "AJ Team_schedule"
  description = "AJ Team schedule with multiple rotations"
  team_id     = firehydrant_team.aj_team.id
  time_zone   = "America/Los_Angeles"

  member_ids = [data.firehydrant_user.fh_eng.id, data.firehydrant_user.fh_demo.id, data.firehydrant_user.fh_success.id]

  strategy {
    type           = "custom"
    shift_duration = "PT2H"
  }
}

resource "firehydrant_rotation" "aj_team_aj_team_schedule_daytime_rotation" {
  name        = "Daytime rotation"
  description = "Weekly daytime rotation"
  team_id     = firehydrant_team.aj_team.id
  schedule_id = firehydrant_on_call_schedule.aj_team_aj_team_schedule.id
  time_zone   = "America/Los_Angeles"

  members = [data.firehydrant_user.fh_eng.id, data.firehydrant_user.fh_demo.id, data.firehydrant_user.fh_success.id, data.firehydrant_user.jsmith.id]

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
}

resource "firehydrant_rotation" "aj_team_aj_team_schedule_nighttime_rotation" {
  name        = "Nighttime rotation"
  description = "Daily nighttime rotation"
  team_id     = firehydrant_team.aj_team.id
  schedule_id = firehydrant_on_call_schedule.aj_team_aj_team_schedule.id
  time_zone   = "America/Los_Angeles"

  members = [data.firehydrant_user.fh_eng.id, data.firehydrant_user.fh_demo.id, data.firehydrant_user.fh_success.id, data.firehydrant_user.jsmith.id]

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

resource "firehydrant_escalation_policy" "aj_team_escalation" {
  name    = "AJ Team_escalation"
  team_id = firehydrant_team.aj_team.id

  step {
    timeout = "PT1M"

    targets {
      type = "OnCallSchedule"
      id   = firehydrant_on_call_schedule.aj_team_aj_team_schedule.id
    }
  }

  repetitions = 0
}
