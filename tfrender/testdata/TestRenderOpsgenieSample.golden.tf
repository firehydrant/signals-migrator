terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant"
      version = ">= 0.8.0"
    }
  }
}

data "firehydrant_user" "agajewski" {
  email = "agajewski@firehydrant.com"
}

data "firehydrant_user" "ajasso" {
  email = "ajasso@firehydrant.com"
}

data "firehydrant_user" "abakhtiari" {
  email = "abakhtiari@firehydrant.com"
}

data "firehydrant_user" "akoenigbautista" {
  email = "akoenigbautista@firehydrant.com"
}

data "firehydrant_user" "breese" {
  email = "breese@firehydrant.com"
}

data "firehydrant_user" "btrowbridge" {
  email = "btrowbridge@firehydrant.com"
}

data "firehydrant_user" "boblad" {
  email = "boblad@firehydrant.com"
}

data "firehydrant_user" "cbeavers" {
  email = "cbeavers@firehydrant.com"
}

data "firehydrant_user" "cweber" {
  email = "cweber@firehydrant.com"
}

data "firehydrant_user" "cquinn" {
  email = "cquinn@firehydrant.com"
}

data "firehydrant_user" "cgrant" {
  email = "cgrant@firehydrant.com"
}

data "firehydrant_user" "dcondomitti" {
  email = "dcondomitti@firehydrant.com"
}

data "firehydrant_user" "dleong" {
  email = "dleong@firehydrant.com"
}

data "firehydrant_user" "davidcelis" {
  email = "davidcelis@firehydrant.com"
}

data "firehydrant_user" "deleted_dlinch_at_firehydrant_io" {
  email = "deleted+dlinch_at_firehydrant.io@firehydrant.com"
}

data "firehydrant_user" "dylan" {
  email = "dylan@firehydrant.com"
}

data "firehydrant_user" "epegues" {
  email = "epegues@firehydrant.com"
}

data "firehydrant_user" "it_integrations" {
  email = "it-integrations@firehydrant.com"
}

data "firehydrant_user" "jhamilton" {
  email = "jhamilton@firehydrant.com"
}

data "firehydrant_user" "jscott" {
  email = "jscott@firehydrant.com"
}

data "firehydrant_user" "lcastellanos" {
  email = "lcastellanos@firehydrant.com"
}

data "firehydrant_user" "lincoln" {
  email = "lincoln@firehydrant.com"
}

data "firehydrant_user" "mhughes" {
  email = "mhughes@firehydrant.com"
}

data "firehydrant_user" "mevans" {
  email = "mevans@firehydrant.com"
}

data "firehydrant_user" "mrulonmiller" {
  email = "mrulonmiller@firehydrant.com"
}

data "firehydrant_user" "mtison" {
  email = "mtison@firehydrant.com"
}

data "firehydrant_user" "robert" {
  email = "robert@firehydrant.com"
}

data "firehydrant_user" "ralipour" {
  email = "ralipour@firehydrant.com"
}

data "firehydrant_user" "tdamico" {
  email = "tdamico@firehydrant.com"
}

data "firehydrant_user" "vthanh" {
  email = "vthanh@firehydrant.com"
}

data "firehydrant_user" "whusin" {
  email = "whusin@firehydrant.com"
}

data "firehydrant_user" "ymuller" {
  email = "ymuller@firehydrant.com"
}

resource "firehydrant_team" "vinny_test_team" {
  name = "Vinny Test Team"

  memberships {
    user_id = data.firehydrant_user.mhughes.id
  }

  memberships {
    user_id = data.firehydrant_user.vthanh.id
  }
}

resource "firehydrant_team" "platform" {
  name = "Platform"
}

resource "firehydrant_team" "aj_team" {
  name = "AJ Team"

  memberships {
    user_id = data.firehydrant_user.cweber.id
  }

  memberships {
    user_id = data.firehydrant_user.akoenigbautista.id
  }

  memberships {
    user_id = data.firehydrant_user.abakhtiari.id
  }

  memberships {
    user_id = data.firehydrant_user.ajasso.id
  }
}

resource "firehydrant_team" "customer_success" {
  name = "Customer Success"

  memberships {
    user_id = data.firehydrant_user.cweber.id
  }
}

resource "firehydrant_team" "chris_test_team" {
  name = "Chris Test Team"

  memberships {
    user_id = data.firehydrant_user.cweber.id
  }
}

resource "firehydrant_team" "andrewkb_test_team" {
  name = "AndrewKB Test Team"
}

resource "firehydrant_team" "noodlebrigade" {
  name = "noodlebrigade"

  memberships {
    user_id = data.firehydrant_user.breese.id
  }
}

resource "firehydrant_team" "christine_test_team" {
  name = "Christine Test Team"

  memberships {
    user_id = data.firehydrant_user.agajewski.id
  }
}

resource "firehydrant_team" "krista_test_team" {
  name = "krista-test-team"

  memberships {
    user_id = data.firehydrant_user.abakhtiari.id
  }
}

resource "firehydrant_on_call_schedule" "customer_success_customer_success_schedule_rot1" {
  name        = "Customer Success_schedule - Rot1"
  description = "(Rot1)"
  team_id     = firehydrant_team.customer_success.id
  time_zone   = "America/Los_Angeles"

  member_ids = [data.firehydrant_user.cweber.id]

  strategy {
    type         = "weekly"
    handoff_day  = "tuesday"
    handoff_time = "04:45:32"
  }
}

resource "firehydrant_on_call_schedule" "chris_test_team_chris_test_team_schedule_rot1" {
  name        = "Chris Test Team_schedule - Rot1"
  description = "(Rot1)"
  team_id     = firehydrant_team.chris_test_team.id
  time_zone   = "America/Los_Angeles"

  member_ids = [data.firehydrant_user.cweber.id]

  strategy {
    type         = "weekly"
    handoff_day  = "monday"
    handoff_time = "07:00:00"
  }
}

resource "firehydrant_on_call_schedule" "vinny_test_team_vinny_test_team_schedule_first" {
  name        = "Vinny Test Team_schedule - First"
  description = "(First)"
  team_id     = firehydrant_team.vinny_test_team.id
  time_zone   = "America/Los_Angeles"

  member_ids = [data.firehydrant_user.vthanh.id]

  strategy {
    type         = "daily"
    handoff_time = "08:00:00"
  }
}

resource "firehydrant_on_call_schedule" "aj_team_aj_team_schedule_daytime_rotation" {
  name        = "AJ Team_schedule - Daytime rotation"
  description = "(Daytime rotation)"
  team_id     = firehydrant_team.aj_team.id
  time_zone   = "America/Los_Angeles"

  member_ids = [data.firehydrant_user.ajasso.id, data.firehydrant_user.abakhtiari.id, data.firehydrant_user.cweber.id, data.firehydrant_user.akoenigbautista.id]

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

  member_ids = [data.firehydrant_user.ajasso.id, data.firehydrant_user.abakhtiari.id, data.firehydrant_user.cweber.id, data.firehydrant_user.akoenigbautista.id]

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

resource "firehydrant_on_call_schedule" "noodlebrigade_noodlebrigade_schedule_rot1" {
  name        = "noodlebrigade_schedule - Rot1"
  description = "(Rot1)"
  team_id     = firehydrant_team.noodlebrigade.id
  time_zone   = "America/Los_Angeles"

  member_ids = [data.firehydrant_user.breese.id]

  strategy {
    type         = "weekly"
    handoff_day  = "monday"
    handoff_time = "08:00:00"
  }
}

resource "firehydrant_on_call_schedule" "krista_test_team_krista_test_team_schedule_rot1" {
  name        = "krista-test-team_schedule - Rot1"
  description = "(Rot1)"
  team_id     = firehydrant_team.krista_test_team.id
  time_zone   = "America/Los_Angeles"

  member_ids = [data.firehydrant_user.abakhtiari.id]

  strategy {
    type         = "weekly"
    handoff_day  = "tuesday"
    handoff_time = "01:24:00"
  }
}
