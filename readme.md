# Signals Migrator

This can be used to import resources from legacy alerting providers into Signals.

# Todo

* Add support for getting transposer URLs (specifically for Datadog) to the team data resource or a signals ingest URL data resource
* Deduplicate FireHydrant team definitions
* Add support for importing escalation policies -- a default is created
* Test coverage
* Build + publish Docker image to simplify usage down to `docker run firehydrant/signals-migrator import`

## Usage

## Generate Datadog Webhooks

A Datadog API key and application key with the `create_webhook` scope are required.

Ensure FIREHYDRANT_API_KEY, DATADOG_API_KEY and DATADOG_APP_KEY are exported and run `go run main.go datadog`. This will generate Terraform files in output/*.tf.

*They will not be valid TF until ingest URL support is added to the provider and can not be applied*

## Generate Signals resources from PagerDuty
Ensure PAGERDUTY_API_KEY and FIREHYDRANT_API_KEY are exported and run `go run main.go import`. This will generate Terraform files in output/*.tf

### Process
* Fetch users from provider
* Fetch teams from provider
* Fetch schedules for each team from provider
* Map any users that don't exist in FH with the same email address
* Map any teams that don't exist in FH with the same name
* Render data resources for each user in the organization
* Render data resources for eacn team in the organization
* Render on_call_schedule resources for each schedule, referencing the users and teams above

## Debugging
Debugging this is a little tricky if you end up in the bubbletea resource picker because no TTY is available. You need to run the applicationw with delve and then `start debugging` to attach VSCode to it.

`dlv debug github.com/firehydrant/signals-migrator --headless --listen=0.0.0.0:2345 --log --api-version 2 -- import`


## Sample outputs

### Datadog
```
provider "datadog" {
}
locals {
  payload = "{\n  \"summary\": \"$EVENT_TITLE\",\n  \"body\": \"$EVENT_MSG\",\n  \"unique_key\": \"$ALERT_ID\",\n  \"level\": \"$ALERT_TYPE\",\n  \"status\": \"$ALERT_TRANSITION\",\n  \"links\": [{\"href\": \"$LINK\", \"text\": \"Datadog Monitor\"}],\n  \"images\": [{\"src\": \"$SNAPSHOT\", \"alt\": \"Snapshot from $EVENT_TITLE\"}],\n  \"tags\": \"$TAGS\"\n}"
}
resource "datadog_webhook" "team-edward" {
  name      = "firehydrant-team-edward"
  url       = data.firehydrant_team.team-edward.datadog_transpose_url
  encode_as = "json"
  payload   = local.payload
}
resource "datadog_webhook" "team-with-no-service" {
  name      = "firehydrant-team-with-no-service"
  url       = data.firehydrant_team.team-with-no-service.datadog_transpose_url
  encode_as = "json"
  payload   = local.payload
}
```

### PagerDuty
```
provider "firehydrant" {
}
data "firehydrant_user" "agajewskiatfyrehydrant-com" {
  email = "agajewski@fyrehydrant.com"
}
data "firehydrant_user" "akoenigbautistaatfyrehydrant-com" {
  email = "akoenigbautista@fyrehydrant.com"
}
data "firehydrant_user" "breeseatfyrehydrant-com" {
  email = "breese@fyrehydrant.com"
}
data "firehydrant_user" "cyiatfyrehydrant-com" {
  email = "cyi@fyrehydrant.com"
}
data "firehydrant_user" "alukensatfyrehydrant-com" {
  email = "alukens@fyrehydrant.com"
}
data "firehydrant_user" "bobladatfyrehydrant-com" {
  email = "boblad@fyrehydrant.com"
}
data "firehydrant_team" "team-with-no-service" {
  id = "718102e6-1d41-4ba4-bf26-06c40a16add4"
}
resource "firehydrant_on_call_schedule" "bmorton-primary" {
  name        = "bmorton - primary"
  description = ""
  team_id     = data.firehydrant_team.team-with-no-service.id
  time_zone   = "America/Los_Angeles"
  strategy {
    type         = "Weekly"
    handoff_time = "16:00:00"
    handoff_day  = "Friday"
  }
  member_ids = [data.firehydrant_user.fh-engatfirehydrant-io.id, data.firehydrant_user.fh-success-richard-engatfirehydrant-io.id]
}
resource "firehydrant_escalation_policy" "bmorton-primary" {
  name        = "default"
  description = "Default escalation policy"
  team_id     = data.firehydrant_team.team-with-no-service.id
  step {
    timeout = "PT5M"
    targets {
      type = "OnCallSchedule"
      id   = resource.firehydrant_on_call_schedule.bmorton-primary.id
    }
  }
  repetitions = 2
}
```