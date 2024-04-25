terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant"
      version = ">= 0.8.0"
    }
  }
}

data "firehydrant_user" "jsmith" {
  email = "jsmith@firehydrant.com"
}

data "firehydrant_user" "fh_demo" {
  email = "fh-demo@firehydrant.io"
}

data "firehydrant_user" "fh_eng" {
  email = "fh-eng@firehydrant.io"
}

data "firehydrant_user" "fh_success" {
  email = "fh-success@firehydrant.com"
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

resource "firehydrant_on_call_schedule" "aj_team_aj_team_schedule_rota3" {
  name        = "AJ Team_schedule - Rota3"
  description = "(Rota3)"
  team_id     = firehydrant_team.aj_team.id
  time_zone   = "America/Los_Angeles"

  member_ids = [data.firehydrant_user.fh_demo.id, data.firehydrant_user.fh_eng.id, data.firehydrant_user.fh_success.id]

  strategy {
    type           = "custom"
    shift_duration = "PT2H"
  }
}

resource "firehydrant_on_call_schedule" "aj_team_aj_team_schedule_daytime_rotation" {
  name        = "AJ Team_schedule - Daytime rotation"
  description = "(Daytime rotation)"
  team_id     = firehydrant_team.aj_team.id
  time_zone   = "America/Los_Angeles"

  member_ids = [data.firehydrant_user.fh_demo.id, data.firehydrant_user.fh_eng.id, data.firehydrant_user.jsmith.id, data.firehydrant_user.fh_success.id]

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

resource "firehydrant_on_call_schedule" "aj_team_aj_team_schedule_nighttime_rotation" {
  name        = "AJ Team_schedule - Nighttime rotation"
  description = "(Nighttime rotation)"
  team_id     = firehydrant_team.aj_team.id
  time_zone   = "America/Los_Angeles"

  member_ids = [data.firehydrant_user.fh_demo.id, data.firehydrant_user.fh_eng.id, data.firehydrant_user.jsmith.id, data.firehydrant_user.fh_success.id]

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
      id   = firehydrant_on_call_schedule.aj_team_aj_team_schedule_daytime_rotation.id
    }

    targets {
      type = "OnCallSchedule"
      id   = firehydrant_on_call_schedule.aj_team_aj_team_schedule_nighttime_rotation.id
    }

    targets {
      type = "OnCallSchedule"
      id   = firehydrant_on_call_schedule.aj_team_aj_team_schedule_rota3.id
    }
  }

  repetitions = 0
}
