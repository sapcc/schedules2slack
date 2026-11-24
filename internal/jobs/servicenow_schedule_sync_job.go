package jobs

import (
	"fmt"
	"time"

	"github.com/robfig/cron"
	"github.com/sapcc/schedules2slack/internal/clients/servicenow"
	slackclient "github.com/sapcc/schedules2slack/internal/clients/slack"
	"github.com/sapcc/schedules2slack/internal/config"
	log "github.com/sirupsen/logrus"
)

// NewSchedulesSyncJob creates a new job to sync members of schedules to a slack user group
func NewScheduleSyncJob(cfg config.ScheduleSync, dryrun bool, sn *servicenow.Client, slackClient *slackclient.Client) (*ServicenowScheduleToSlackJob, error) {
	schedule, err := cron.Parse(cfg.CrontabExpressionForRepetition)
	if err != nil {
		return nil, fmt.Errorf("job: invalid cron schedule '%s': %w", cfg.CrontabExpressionForRepetition, err)
	}
	return &ServicenowScheduleToSlackJob{
		syncOpts:         cfg.SyncOptions,
		dryrun:           dryrun,
		slackHandle:      cfg.SyncObjects.SlackGroupHandle,
		scheduleGroupID:  cfg.SyncObjects.GroupID,
		dutyChannelID:    cfg.DutyChannelID,
		schedule:         schedule,
		servicenowClient: sn,
		slackClient:      slackClient,
		syncjob:          nil,
	}, nil
}

type ServicenowScheduleToSlackJob struct {
	syncOpts config.ScheduleSyncOptions // options for tasks during sync
	schedule cron.Schedule              // on which this job runs
	dryrun   bool                       // when enabled changes are not manifested
	err      error                      // err used for slack info message

	slackClient      *slackclient.Client // slack API access
	servicenowClient *servicenow.Client

	slackHandle        string              // of the target user group
	scheduleGroupID    string              // IDs of the schedules to sync
	dutyChannelID      string              // channel to post on-/off-duty notifications
	servicenowSchedule servicenow.Schedule // servicenow schedule

	syncjob SyncJob
}

// Run syncs schedule members to slack user group
func (s *ServicenowScheduleToSlackJob) Run() error {
	log.Info(s.Name())
	s.err = nil

	groupmember, err := s.servicenowClient.ListScheduleMember(s.scheduleGroupID)
	if err != nil {
		s.err = err
		return err
	}
	s.servicenowSchedule.Members = groupmember

	onCallMember, err := s.servicenowClient.ListSpans(s.servicenowSchedule, s.scheduleGroupID)
	if err != nil {
		log.Error(err)
		return err
	}
	s.servicenowSchedule.OnOnCall = onCallMember

	slackUsers, err := s.slackClient.MatchUsers(onCallMember)
	if err != nil && len(onCallMember) > 0 {
		s.err = err
		return err
	}

	currentGroup, err := s.slackClient.GetSlackGroup(s.slackHandle)
	if err != nil {
		s.err = err
		return fmt.Errorf("job: retrieving slack group %s failed: %w", s.slackHandle, err)
	}

	if len(slackUsers) == 0 && s.syncOpts.DisableSlackHandleTemporaryIfNoneOnShift {
		groupWasActive := currentGroup.DateDelete.String() == "0"
		if groupWasActive {
			if !s.dryrun {
				if err = s.slackClient.DisableGroup(currentGroup.ID); err != nil {
					s.err = err
					return fmt.Errorf("job: disabling slack group %s failed: %w", s.slackHandle, err)
				}
				s.postDutyNotification(fmt.Sprintf("<!subteam^%s> is out of duty", currentGroup.ID))
			} else {
				log.Infof("job: dry run — would disable group '%s' and post out-of-duty notice", s.slackHandle)
			}
		}
		return nil
	}

	groupWasDisabled := currentGroup.DateDelete.String() != "0"

	if _, err = s.slackClient.AddToGroup(s.slackHandle, slackUsers, s.dryrun); err != nil {
		s.err = err
		return fmt.Errorf("job: adding OnDuty members to slack group %s failed: %w", s.slackHandle, err)
	}

	if groupWasDisabled && !s.dryrun && s.dutyChannelID != "" {
		s.postDutyNotification(fmt.Sprintf("<!subteam^%s> is online", currentGroup.ID))
	}

	return nil
}

func (s *ServicenowScheduleToSlackJob) postDutyNotification(text string) {
	if s.dutyChannelID == "" {
		return
	}
	msg := slackclient.NewTextMessage(text)
	if err := s.slackClient.PostToChannel(s.dutyChannelID, msg); err != nil {
		log.Warnf("job: posting duty notification failed: %s", err)
	}
}

// Name of the job
func (s *ServicenowScheduleToSlackJob) Name() string {
	return fmt.Sprintf("job: sync schedule(s) '%s' to slack group: '%s'", s.scheduleGroupID, s.slackHandle)
}

// Icon returns name of icon to show in Slack messages
func (s *ServicenowScheduleToSlackJob) Icon() string {
	return ":calendar:"
}

// JobType as string
func (s *ServicenowScheduleToSlackJob) JobType() string {
	return string(ScheduleSync)
}

// SlackHandle of the slack user group
func (s *ServicenowScheduleToSlackJob) SlackHandle() string {
	return s.slackHandle
}

// Dryrun is true when the job is not performing changes
func (s *ServicenowScheduleToSlackJob) Dryrun() bool {
	return s.dryrun
}

// NextRun returns the time from now when the cron is next executed
func (s *ServicenowScheduleToSlackJob) NextRun() time.Time {
	return s.schedule.Next(time.Now())
}

// Error if any occurred during the sync
func (s *ServicenowScheduleToSlackJob) Error() error {
	return s.err
}
