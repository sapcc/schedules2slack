# schedules2slack

Syncs user from ServiceNow Schedules via CPI people on shift from Schedules to slack groups.
Heir of pagerduty2slack.

## Feature List

* We use a cron format to schedule each sync jobs
* handover time frame for schedule sync possible
* there is also the possibility to check on if a phone is set as contact
* disable a slack

## Some words on the job config

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
  schedules-to-slack-group:

    - crontabExpressionForRepetition: 5 7,8,13,14,19,20 \* \* \*
      syncOptions:
        slackHandleNoOneOnShiftStrategy: disable --> LastOnShift | disable (default)
        syncStyle: OnlyPrimary | AllActiveLayers (default)
      syncObjects:
        slackGroupHandle: "onduty-team-no1"
        groupIds:
          - "id from url"
          - "id from url"
    - crontabExpressionForRepetition: 5 7,8,13,14,19,20 \* \* \*
      syncOptions:
        slackHandleNoOneOnShiftStrategy: disable --> LastOnShift | disable (default)
        syncStyle: OnlyPrimary | AllActiveLayers (default)
      syncObjects:
        slackGroupHandle: "onduty-team-no2"
        groupIds:
          - "id from url"

