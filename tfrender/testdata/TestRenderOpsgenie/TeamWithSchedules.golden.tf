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
  # [Opsgenie] 1dc37638-ab52-44f3-848e-a16bcc584fb7 admin@example.io
}

data "firehydrant_user" "mika" {
  id = "e6009411-0015-43e3-815e-ca9db72f4088"
  # [Opsgenie] 9253cf00-6195-4123-a9a6-f9f1e25718d8 mika@example.io
}

data "firehydrant_user" "kiran" {
  id = "4c3f28fa-b402-453c-9652-f014ecbe65a9"
  # [Opsgenie] e94e17aa-418c-44f7-8e47-1eaebf6b5343 kiran@example.com
}

data "firehydrant_user" "jack" {
  id = "6c08bff2-98f6-4ee9-8de1-12202186d084"
  # [Opsgenie] d68757c4-5eec-4560-8c5b-91c463f87dd8 jack@example.com
}

data "firehydrant_user" "wong" {
  id = "032a1f07-987e-4f76-8273-136e08e50baa"
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

  memberships_input = [{ user_id = data.firehydrant_user.kiran.id }, { user_id = data.firehydrant_user.mika.id }, { user_id = data.firehydrant_user.local.id }]
}

resource "firehydrant_team" "customer_success" {
  name = "Customer Success"
}

resource "firehydrant_team" "noodlebrigade" {
  name = "noodlebrigade"

  memberships_input = [{ user_id = data.firehydrant_user.jack.id }]
}

resource "firehydrant_team" "christine_test_team" {
  name = "Christine Test Team"
}

resource "firehydrant_signals_api_on_call_schedule" "customer_success_customer_success_schedule_rot1" {
  name        = "Customer Success_schedule - Rot1"
  description = "(Rot1)"
  team_id     = firehydrant_team.customer_success.id
  time_zone   = "America/Los_Angeles"

  members_input = []

  strategy_input = {
    handoff_day  = "tuesday"
    handoff_time = "04:45:32"
    type         = "weekly"
  }
}

resource "firehydrant_signals_api_on_call_schedule" "wong_squad_wong_team_schedule_first" {
  name        = "Wong Team_schedule - First"
  description = "(First)"
  team_id     = firehydrant_team.wong_squad.id
  time_zone   = "America/Los_Angeles"

  members_input = []

  strategy_input = {
    handoff_time = "08:00:00"
    type         = "daily"
  }
}

resource "firehydrant_signals_api_on_call_schedule" "aj_team_aj_team_schedule_daytime_rotation" {
  name        = "AJ Team_schedule - Daytime rotation"
  description = "(Daytime rotation)"
  team_id     = firehydrant_team.aj_team.id
  time_zone   = "America/Los_Angeles"

  members_input = [{ user_id = data.firehydrant_user.mika.id }, { user_id = data.firehydrant_user.local.id }, { user_id = data.firehydrant_user.kiran.id }]

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

  members_input = [{ user_id = data.firehydrant_user.mika.id }, { user_id = data.firehydrant_user.local.id }, { user_id = data.firehydrant_user.kiran.id }]

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
