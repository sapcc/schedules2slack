# schedules2slack

Syncs ServiceNow on-call schedules to Slack user groups so people can mention whoever is currently on shift via `@group-handle`.

When no one is on shift, the Slack group can be automatically disabled and a notification posted to a configurable duty channel. The Slack App Home tab shows a live overview of all configured groups.

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `SLACK_BOT_TOKEN` | yes | Bot token (`xoxb-…`) |
| `SLACK_USER_TOKEN` | yes | User token (`xoxp-…`) — needed to update user groups |
| `SLACK_APP_TOKEN` | no | App-level token (`xapp-…`) — required for App Home view via Socket Mode |
| `SERVICENOW_API_CERT_PKC12_PWD` | yes | PFX certificate password |
| `SERVICENOW_API_CERT_PKC12_B64` | yes* | PFX certificate as base64 string |
| `SERVICENOW_API_CERT_PKC12` | yes* | Path to PFX certificate file |

*Either `_B64` or the file path must be set.

Encode a PFX file to base64:
```sh
base64 -b 0 -i CCGRNHOUSE_P.pfx -o CCGRNHOUSE_P.pfx.b64
# decode to verify:
base64 --decode -i CCGRNHOUSE_P.pfx.b64 -o CCGRNHOUSE_P.pfx.decoded
```

## Config File

```yaml
global:
  logLevel: info       # debug | info | warn | error
  write: true          # false = dry-run, no changes applied
  runAtStart: true     # run all jobs once on startup

slack:
  infoChannelID: "CXXXXXXXX"        # channel for per-run status messages
  workspaceForChatLinks: "my-workspace"

servicenow:
  apiEndpoint: "https://your-instance.service-now.com"

jobs:
  servicenow-schedules-on-duty-to-slack-group:

    - crontabExpressionForRepetition: "5 * * * *"
      syncOptions:
        disableSlackHandleTemporaryIfNoneOnShift: true   # disable group when shift is empty
        syncStyle: OnlyPrimary                           # OnlyPrimary | AllActiveLayers
      syncObjects:
        slackGroupHandle: "onduty-team-no1"
        groupId: "sys_id from ServiceNow URL"
      dutyChannelID: "CXXXXXXXX"    # channel for "is online" / "is out of duty" notifications
```

Cron format reference: <https://crontab.guru>

```
┌───────────── minute (0–59)
│ ┌───────────── hour (0–23)
│ │ ┌───────────── day of month (1–31)
│ │ │ ┌───────────── month (1–12)
│ │ │ │ ┌───────────── day of week (0–6, Sunday=0)
│ │ │ │ │
* * * * *
```

## Slack App Setup

The bot requires the following OAuth scopes:

- `usergroups:read` / `usergroups:write`
- `users:read`
- `channels:read`
- `chat:write`
- `app_mentions:read` *(optional)*

For the App Home view (Socket Mode):

1. Enable **Socket Mode** in your Slack App settings and generate an App-Level Token with the `connections:write` scope — this becomes `SLACK_APP_TOKEN`.
2. Under **Event Subscriptions → Subscribe to bot events**, add `app_home_opened`.
3. Under **App Home**, enable the **Home Tab**.

## Run Locally

```sh
# build
make build-all

# set required env vars
export SLACK_BOT_TOKEN=xoxb-...
export SLACK_USER_TOKEN=xoxp-...
export SLACK_APP_TOKEN=xapp-...        # optional, for App Home view
export SERVICENOW_API_CERT_PKC12_B64=...
export SERVICENOW_API_CERT_PKC12_PWD=...

# dry-run (no changes applied)
./build/schedules2slack -config ./config.yml

# with writes enabled (overrides config)
./build/schedules2slack -config ./config.yml -write
```

The `-write` flag overrides the `global.write` setting in the config file.

## Docker

See `makefile` for `docker-build` and `docker-push` targets. The container expects the config file at `/etc/config/_run_config.yaml`.
