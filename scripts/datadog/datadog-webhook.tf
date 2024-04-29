terraform {
  required_version = "~> 1.5.5"

  required_providers {
    datadog = {
      source  = "DataDog/datadog"
      version = ">= 3.38.0"
    }
  }
}

provider "datadog" {
  api_key = ""
  app_key = ""
  api_url = ""
}

resource "datadog_webhook" "firehydrant-myteam" {
  name           = "firehydrant-myteam"
  url            = "<team URL>"
  encode_as      = "json"
  payload        = <<EOF
{
    "summary": "$EVENT_TITLE",
    "body": "$EVENT_MSG",
    "unique_key": "$ALERT_ID",
    "level": "$ALERT_TYPE",
    "status": "$ALERT_TRANSITION",
    "links": [{"href": "$LINK", "text": "Datadog Monitor"}],
    "images": [{"src": "$SNAPSHOT", "alt": "Snapshot from $EVENT_TITLE"}]
}
EOF
}
