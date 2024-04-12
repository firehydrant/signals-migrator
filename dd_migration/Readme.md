# Datadog Migration Tools

Scripts and terraform files to aid in migration for customers using datadog for alerting

## Tools

## `update_monitors.sh`

This script will search a datadog instance for all monitors with a given notification method (such as `@slack-platform-alerts`) and create a new monitor with the same name and all the same settings, but with a new notification method (such as `@firehydrant-platform-alerts`).  Optionally, the `UPDATE_IN_PLACE` variable in the script can be set to `true`, in which case the existing monitor will be modified to use the new notification method instead.

This script requires that a datadog API key be exported as `DD_API_KEY`, as well as an application key exported as `DD_APP_KEY`.  Note that the application key must have all scopes needed for modifying all expected alerts and so should be created by someone with admin access.  In the script, the datadog site parameter `DD_SITE_PARAMETER` needs to be set to connect with any site other than the default US1.  See this page for the appropriate value: https://docs.datadoghq.com/getting_started/site/#access-the-datadog-site

Known issue: All monitors in question need to have some text (apart from the notification method) in the `message` field for the monitor.  Datadog allows this field to contain the notification method only, but then returns an empty string for the message as part of the API and thus it can't be modified in this way.

## `datadog-webhook.tf`

This is an example terraform file for creating a signals compatible webhook at datadog.  It requires the API key and Application key as described above, as well as the base datadog API url as described in the provider setup here: https://registry.terraform.io/providers/DataDog/datadog/latest/docs

The team-specific URL would be added where indicated.  This would create a notification method addressable in datadog as `@webhook-firehydrant-myteam`