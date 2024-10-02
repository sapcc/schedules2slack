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
func NewTicketSyncJob(cfg config.TicketSync, dryrun bool, sn *servicenow.Client, slackClient *slackclient.Client) (*ServicenowTicketToSlackJob, error) {
	schedule, err := cron.Parse(cfg.CrontabExpressionForRepetition)
	if err != nil {
		return nil, fmt.Errorf("job: invalid cron schedule '%s': %w", cfg.CrontabExpressionForRepetition, err)
	}
	return &ServicenowTicketToSlackJob{
		syncOpts:         cfg.SyncOptions,
		schedule:         schedule,
		dryrun:           dryrun,
		err:              err,
		slackClient:      slackClient,
		servicenowClient: sn,
		tickets:          nil,
		syncjob:          nil,
	}, nil
}

type ServicenowTicketToSlackJob struct {
	syncOpts config.TicketSyncOptions // options for tasks during sync
	schedule cron.Schedule            // on which this job runs
	dryrun   bool                     // when enabled changes are not manifested
	err      error                    // err used for slack info message

	slackClient      *slackclient.Client // slack API access
	servicenowClient *servicenow.Client

	tickets []servicenow.Ticket

	syncjob SyncJob
}

// Run syncs schedule members to slack user group
func (s *ServicenowTicketToSlackJob) Run() error {
	log.Info(s.Name())
	s.err = nil

    //var t []servicenow.Ticket
	t, err := s.servicenowClient.ListTickets(s.syncOpts.SysParmQuery, s.syncOpts.SysParmLimit, s.syncOpts.TicketType)
	if err != nil {
		s.err = err
		return err
	}
    s.tickets = *t
	return nil
}

// Name of the job
func (s *ServicenowTicketToSlackJob) Name() string {
	return fmt.Sprintf("job: sync schedule(s) '%s' to slack group: '%s'", s.syncOpts.SysParmQuery, s.syncOpts.TicketType)
}

// Icon returns name of icon to show in Slack messages
func (s *ServicenowTicketToSlackJob) Icon() string {
	return ":calendar:"
}

// JobType as string
func (s *ServicenowTicketToSlackJob) JobType() string {
	return string(TicketSync)
}

// Dryrun is true when the job is not performing changes
func (s *ServicenowTicketToSlackJob) Dryrun() bool {
	return s.dryrun
}

// NextRun returns the time from now when the cron is next executed
func (s *ServicenowTicketToSlackJob) NextRun() time.Time {
	return s.schedule.Next(time.Now())
}

// Error if any occurred during the sync
func (s *ServicenowTicketToSlackJob) Error() error {
	return s.err
}
