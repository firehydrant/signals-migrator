#! /usr/bin/env bash
set -eo pipefail

# By default, we will make a copy of any alert with notification OLD_NOTIFICATION that is identical in all other ways but replaces OLD_NOTIFICATION
# with NEW_NOTIFICATION.  Setting UPDATE_IN_PLACE here changes that behavior to modify the existing alert instead of making a new one.
UPDATE_IN_PLACE="false"
# The existing notification as it appears in the Datadog message.  For example @slack-platform-alerts
OLD_NOTIFICATION=""
# The new notification as it would appear in the Datadog message.  For example @firehydrant-platform-alerts
NEW_NOTIFICATION=""
# The appropriate Datadog site parameter from the table here: https://docs.datadoghq.com/getting_started/site/#access-the-datadog-site
DD_SITE_PARAMETER="datadoghq.com"

if [ -z "$DD_API_KEY" ]
then
  echo "Please export your datadog API key as DD_API_KEY"
  exit 1
fi

if [ -z "$DD_APP_KEY" ]
then
  echo "Please export your datadog application key as DD_APP_KEY"
  exit 1
fi

set -u

URL_ENCODED_OLD_NOTIFICATION=$(echo ${OLD_NOTIFICATION:1}| jq -Rr @uri)

RAW_SEARCH=$(curl -s --fail-with-body 'https://api.'"$DD_SITE_PARAMETER"'/api/v1/monitor/search?query=notification%3A"'"$URL_ENCODED_OLD_NOTIFICATION"'"' -H "Accept: application/json" -H "DD-API-KEY: $DD_API_KEY" -H "DD-APPLICATION-KEY: $DD_APP_KEY")

PAGES=$(echo $RAW_SEARCH | jq .metadata.page_count)
MONITOR_COUNT=$(echo $RAW_SEARCH | jq .metadata.total_count)

MONITORS=()
for ((i=0; i<$PAGES; i++)); do
  MONITORS+=$(curl -s --fail-with-body 'https://api.'"$DD_SITE_PARAMETER"'/api/v1/monitor/search?query=notification%3A"'"$URL_ENCODED_OLD_NOTIFICATION"'"&page='"$i" -H "Accept: application/json" -H "DD-API-KEY: $DD_API_KEY" -H "DD-APPLICATION-KEY: $DD_APP_KEY" | jq '.monitors | .[] | .id')
  MONITORS+=" "
done

echo "Found $MONITOR_COUNT monitors..."

monitornames=()
for monitorid in $MONITORS; do
  rawmonitor=$(curl -s --fail-with-body "https://api.$DD_SITE_PARAMETER/api/v1/monitor/$monitorid" -H "Accept: application/json" -H "DD-API-KEY: $DD_API_KEY" -H "DD-APPLICATION-KEY: $DD_APP_KEY")
  monitornames+=$(echo $rawmonitor | jq .name)
  monitornames+='\n'
done

if [ $UPDATE_IN_PLACE == "true" ]
then
  echo -e "The following monitors will be updated to point to $NEW_NOTIFICATION:\n"
else
  echo -e "Copies of the following monitors will be created pointing to $NEW_NOTIFICATION:\n"
fi

echo -e $monitornames

read -p "Type Yes to continue: " yn
case $yn in
    Yes ) ;;
    * ) exit 1;;
esac

for monitorid in $MONITORS; do
  rawmonitor=$(curl -s --fail-with-body "https://api.$DD_SITE_PARAMETER/api/v1/monitor/$monitorid" -H "Accept: application/json" -H "DD-API-KEY: $DD_API_KEY" -H "DD-APPLICATION-KEY: $DD_APP_KEY")
  new_message=$(echo $rawmonitor | jq .message | sed "s/$OLD_NOTIFICATION/$NEW_NOTIFICATION/g")
  if [ $UPDATE_IN_PLACE == "true" ]
  then
    curl -X PUT -s --fail-with-body "https://api.$DD_SITE_PARAMETER/api/v1/monitor/$monitorid" \
    -H "Accept: application/json" \
    -H "Content-Type: application/json" \
    -H "DD-API-KEY: $DD_API_KEY" \
    -H "DD-APPLICATION-KEY: $DD_APP_KEY" \
    -d { "message": $new_message }

    echo "updated $(echo $rawmonitor | jq .name) in place"
  else
    payload='{
  "name": '"$(echo $rawmonitor | jq .name)"',
  "message": '"$new_message"',
  "options": '"$(echo $rawmonitor | jq .options)"',
  "priority": '"$(echo $rawmonitor | jq .priority)"',
  "query": '"$(echo $rawmonitor | jq .query)"',
  "restricted_roles": '"$(echo $rawmonitor | jq .restricted_roles)"',
  "tags": '"$(echo $rawmonitor | jq .tags)"',
  "type": '"$(echo $rawmonitor | jq .type)"'
}'
    
    curl -X POST -s --fail-with-body "https://api.$DD_SITE_PARAMETER/api/v1/monitor" \
    -H "Accept: application/json" \
    -H "Content-Type: application/json" \
    -H "DD-API-KEY: $DD_API_KEY" \
    -H "DD-APPLICATION-KEY: $DD_APP_KEY" \
    -d "$payload"

    echo "created new copy of $(echo $rawmonitor | jq .name) with notifications targeting $NEW_NOTIFICATION"
  fi
done
