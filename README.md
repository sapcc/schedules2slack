# schedules2slack (ugly-beta-version)

Syncs from ServiceNow Schedules of a AssignmentGroup People on shift to slack groups.
Heir of schedules2slack.


base64 -b 0 -i CCGRNHOUSE_P.pfx -o CCGRNHOUSE_P.pfx_b64
base64 --decode -i CCGRNHOUSE_P.pfx_b64 -o CCGRNHOUSE_P.pfx_b64_decode

## Feature List

* We use a cron format to schedule each sync jobs

## Some words on the job config

### env values

SLACK_BOT_TOKEN=xoxb-xxxxx
SLACK_USER_TOKEN=xoxp-xxxx
SLACK_SOCKET_TOKEN=xapp-xxxx
SERVICENOW_API_CERT_PKC12_PWD=<secret>
SERVICENOW_API_CERT_PKC12_B64=<base64string> or
SERVICENOW_API_CERT_PKC12=cert_file.pfx


### config file

if you're not a cron hero, check <https://crontab.guru/> as example.

    ┌───────────── minute (0 - 59)
    │ ┌───────────── hour (0 - 23)
    │ │ ┌───────────── day of the month (1 - 31)
    │ │ │ ┌───────────── month (1 - 12)
    │ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday;
    │ │ │ │ │                                   7 is also Sunday on some systems)
    │ │ │ │ │
    │ │ │ │ │
    \* \* \* \* \* command to execute

jobs:
  servicenow-schedules-on-duty-to-slack-group:

    - crontabExpressionForRepetition: 5 7,8,13,14,19,20 \* \* \*
      syncOptions:
        slackHandleNoOneOnShiftStrategy: disable --> LastOnShift | disable (default)
        syncStyle: OnlyPrimary | AllActiveLayers (default)
      syncObjects:
        slackGroupHandle: "onduty-team-no1"
        groupId: "id from url"
    - crontabExpressionForRepetition: 5 7,8,13,14,19,20 \* \* \*
      syncOptions:
        slackHandleNoOneOnShiftStrategy: disable --> LastOnShift | disable (default)
        syncStyle: OnlyPrimary | AllActiveLayers (default)
      syncObjects:
        slackGroupHandle: "onduty-team-no2"
        groupIds:  groupId: "id from url"

