terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant-v2"
      version = "~> 0.3.1"
    }
  }
}

data "firehydrant_user" "jsmith" {
  id = "49ef2cda-ab4f-4599-852c-8cc2c8884523"
  # [Opsgenie] e0a51be7-3c7e-407f-8678-292ab421f55f jsmith@example.com
}

data "firehydrant_user" "fh_demo" {
  id = "90c17208-46b4-4e82-b9b8-b8f5d8215a05"
  # [Opsgenie] 1dc37638-ab52-44f3-848e-a16bcc584fb7 fh-demo@example.com
}

data "firehydrant_user" "fh_eng" {
  id = "66506894-ecbc-4034-b8e6-30851dabf5f3"
  # [Opsgenie] 9253cf00-6195-4123-a9a6-f9f1e25718d8 fh-eng@example.com
}

data "firehydrant_user" "fh_success" {
  id = "a8fc03aa-8443-4c76-819c-8b7242fec459"
  # [Opsgenie] e94e17aa-418c-44f7-8e47-1eaebf6b5343 fh-success@example.com
}

resource "firehydrant_team" "aj_team" {
  name = "AJ Team"

  memberships_input = [{ user_id = data.firehydrant_user.fh_eng.id }, { user_id = data.firehydrant_user.fh_demo.id }, { user_id = data.firehydrant_user.jsmith.id }, { user_id = data.firehydrant_user.fh_success.id }]
}

resource "firehydrant_signals_api_on_call_schedule" "aj_team_aj_team_schedule_rota3" {
  name        = "AJ Team_schedule - Rota3"
  description = "(Rota3)"
  team_id     = firehydrant_team.aj_team.id
  time_zone   = "America/Los_Angeles"

  members_input = [{ user_id = data.firehydrant_user.fh_eng.id }, { user_id = data.firehydrant_user.fh_demo.id }, { user_id = data.firehydrant_user.fh_success.id }]

  strategy_input = {
    shift_duration = "PT2H"
    type           = "custom"
  }
}

resource "firehydrant_signals_api_on_call_schedule" "aj_team_aj_team_schedule_daytime_rotation" {
  name        = "AJ Team_schedule - Daytime rotation"
  description = "(Daytime rotation)"
  team_id     = firehydrant_team.aj_team.id
  time_zone   = "America/Los_Angeles"

  members_input = [{ user_id = data.firehydrant_user.fh_eng.id }, { user_id = data.firehydrant_user.fh_demo.id }, { user_id = data.firehydrant_user.fh_success.id }, { user_id = data.firehydrant_user.jsmith.id }]

  strategy_input = {
    handoff_day  = "monday"
    handoff_time = "07:00:00"
    type         = "weekly"
  }

  restrictions_input = [{
    end_day    = "monday"
    end_time   = "18:00:00"
    start_day  = "monday"
    start_time = "08:00:00"
    }, {
    end_day    = "tuesday"
    end_time   = "18:00:00"
    start_day  = "tuesday"
    start_time = "08:00:00"
    }, {
    end_day    = "wednesday"
    end_time   = "18:00:00"
    start_day  = "wednesday"
    start_time = "08:00:00"
    }, {
    end_day    = "thursday"
    end_time   = "18:00:00"
    start_day  = "thursday"
    start_time = "08:00:00"
    }, {
    end_day    = "friday"
    end_time   = "18:00:00"
    start_day  = "friday"
    start_time = "08:00:00"
  }]
}

resource "firehydrant_signals_api_on_call_schedule" "aj_team_aj_team_schedule_nighttime_rotation" {
  name        = "AJ Team_schedule - Nighttime rotation"
  description = "(Nighttime rotation)"
  team_id     = firehydrant_team.aj_team.id
  time_zone   = "America/Los_Angeles"

  members_input = [{ user_id = data.firehydrant_user.fh_eng.id }, { user_id = data.firehydrant_user.fh_demo.id }, { user_id = data.firehydrant_user.fh_success.id }, { user_id = data.firehydrant_user.jsmith.id }]

  strategy_input = {
    handoff_time = "15:00:00"
    type         = "daily"
  }

  restrictions_input = [{
    end_day    = "monday"
    end_time   = "08:00:00"
    start_day  = "sunday"
    start_time = "18:00:00"
    }, {
    end_day    = "tuesday"
    end_time   = "08:00:00"
    start_day  = "monday"
    start_time = "18:00:00"
    }, {
    end_day    = "wednesday"
    end_time   = "08:00:00"
    start_day  = "tuesday"
    start_time = "18:00:00"
    }, {
    end_day    = "thursday"
    end_time   = "08:00:00"
    start_day  = "wednesday"
    start_time = "18:00:00"
    }, {
    end_day    = "friday"
    end_time   = "08:00:00"
    start_day  = "thursday"
    start_time = "18:00:00"
    }, {
    end_day    = "saturday"
    end_time   = "08:00:00"
    start_day  = "friday"
    start_time = "18:00:00"
    }, {
    end_day    = "sunday"
    end_time   = "08:00:00"
    start_day  = "saturday"
    start_time = "18:00:00"
  }]
}

resource "firehydrant_signals_api_escalation_policy" "aj_team_escalation" {
  name    = "AJ Team_escalation"
  team_id = firehydrant_team.aj_team.id

  steps = [{ timeout = "PT1M", targets = [{ type = "OnCallSchedule", id = firehydrant_signals_api_on_call_schedule.aj_team_aj_team_schedule_daytime_rotation.id }, { type = "OnCallSchedule", id = firehydrant_signals_api_on_call_schedule.aj_team_aj_team_schedule_nighttime_rotation.id }, { type = "OnCallSchedule", id = firehydrant_signals_api_on_call_schedule.aj_team_aj_team_schedule_rota3.id }] }]

  repetitions = 0
}
