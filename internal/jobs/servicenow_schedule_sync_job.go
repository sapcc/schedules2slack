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
	servicenowSchedule servicenow.Schedule // servicenow schedule

	syncjob SyncJob
}

// Run syncs schedule members to slack user group
func (s *ServicenowScheduleToSlackJob) Run() error {
	log.Info(s.Name())
	s.err = nil

	/*schedule, err := s.servicenowClient.ListSchedules(s.scheduleGroupId)
	if err != nil {
		s.err = err
		return err
	}*/

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

	// get all SLACK users, bcz. we need the SLACK user id and match them with the ldap users
	slackUsers, err := s.slackClient.MatchUsers(onCallMember)
	if err != nil {
		s.err = err
		return err
	}

	// put ldap users which also have a slack account to our slack group (who's not in the ldap group is out)
	if _, err = s.slackClient.AddToGroup(s.slackHandle, slackUsers, s.dryrun); err != nil {
		s.err = err
		return fmt.Errorf("job: adding OnDuty members to slack group %s failed: %w", s.slackHandle, err)
	}
	return nil
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
