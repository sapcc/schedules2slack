package jobs

import (
	"fmt"
	"strings"

	"time"

	slackclient "github.com/sapcc/schedules2slack/internal/clients/slack"
	"github.com/sapcc/schedules2slack/internal/config"

	"github.com/slack-go/slack"
)

// ObjectSyncType
type ObjectSyncType string

const (
	ScheduleSync ObjectSyncType = "Schedule"
)

type SyncJob interface {
	// Name of the job
	Name() string
	// Icon returns name of icon to show in Slack messages
	Icon() string
	// JobType as string
	JobType() string
	// SlackHandle of the slack user group
	SlackHandle() string
	// ScheduleObjects returns the schedule/teams synced
	ScheduleObjects() []config.SyncObject
	// SlackInfoMessageBody custom to the Job
	SlackInfoMessageBody() *slack.TextBlockObject
	// Dryrun is true when the job is not performing changes
	Dryrun() bool
	// NextRun returns the time from now when the cron is next executed
	NextRun() time.Time
	// Error if any occurred during the sync
	Error() error
}

// PostInfoMessage posts a message to slack with the current sync state of the job
func PostInfoMessage(c *slackclient.Client, j *ServicenowScheduleToSlackJob) error {
	divSection := slack.NewDividerBlock()

	sHeaderText := fmt.Sprintf("%s %s > Slack Handle: `%s`", j.Icon(), j.JobType(), j.SlackHandle())
	if j.Dryrun() {
		sHeaderText += " - !!! DRY RUN !!! No update done !!!"
	}
	headerText := slack.NewTextBlockObject(slack.MarkdownType, sHeaderText, false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	var errorText *slack.TextBlockObject
	var errorSection *slack.SectionBlock
	if j.Error() != nil {
		errorText = slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf(":stop-sign: *Error:* %s", j.Error()), false, false)
		errorSection = slack.NewSectionBlock(errorText, nil, nil)
	}

	var fields []*slack.TextBlockObject

	/* TODO: Schedule Link
	var sL []string
	for _, aO := range pMI.Text {
		sL = append(sL, fmt.Sprintf("@%s", aO))
	}

	fields = append(fields, &slack.TextBlockObject{
		Type:     slack.MarkdownType,
		Text:     fmt.Sprintf("*Source*\n%s", strings.Join(sL, "\n")),
		Emoji:    false,
		Verbatim: false,
	}) */

	var sL []string
	for _, aO := range j.servicenowSchedule.OnOnCall {
		sL = append(sL, aO.SlackDisplayValue)
	}

	fields = append(fields, &slack.TextBlockObject{
		Type:     slack.MarkdownType,
		Text:     fmt.Sprintf("*Who is on shift:*\n - %s", strings.Join(sL, ",\n - ")),
		Emoji:    false,
		Verbatim: false,
	})

	fields = append(fields, &slack.TextBlockObject{
		Type:     slack.MarkdownType,
		Text:     fmt.Sprintf(":alarm_clock: *Next run:* %s", j.NextRun().Format(time.RFC822)),
		Emoji:    false,
		Verbatim: false,
	})
	//jobSection := slack.NewSectionBlock(jobText, fields, nil)
	jobSection := slack.NewSectionBlock(nil, fields, nil)

	if errorSection != nil {
		return c.PostMessage(slack.MsgOptionBlocks(headerSection, errorSection, jobSection, divSection))
	}
	return c.PostMessage(slack.MsgOptionBlocks(headerSection, jobSection, divSection))
}
