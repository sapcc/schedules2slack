package slack

import (
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	slackgo "github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// JobState holds the last known state of a sync job for display in the App Home view.
type JobState struct {
	SlackHandle string
	OnShift     []string
	Disabled    bool
	LastRun     time.Time
	NextRun     time.Time
	Err         error
}

// JobStateStore is an in-memory store for job states, safe for concurrent access.
type JobStateStore struct {
	mu     sync.RWMutex
	states map[string]*JobState // keyed by slackHandle
}

func NewJobStateStore() *JobStateStore {
	return &JobStateStore{states: make(map[string]*JobState)}
}

func (s *JobStateStore) Update(state JobState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state.SlackHandle] = &state
}

func (s *JobStateStore) All() []*JobState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*JobState, 0, len(s.states))
	for _, v := range s.states {
		result = append(result, v)
	}
	return result
}

// StartSocketMode starts the Socket Mode event loop and handles app_home_opened events.
// It blocks until the connection is closed, so callers should run it in a goroutine.
func (c *Client) StartSocketMode(appToken string, store *JobStateStore) {
	if appToken == "" {
		log.Warn("slack: SLACK_APP_TOKEN not set, App Home view disabled")
		return
	}

	// Socket Mode requires a client initialized with the app-level token (xapp-).
	// PublishView calls go through the regular botClient.
	socketAPI := slackgo.New("", slackgo.OptionAppLevelToken(appToken))
	smClient := socketmode.New(socketAPI, socketmode.OptionDebug(false))
	handler := socketmode.NewSocketmodeHandler(smClient)

	handler.HandleEvents(slackevents.AppHomeOpened, func(evt *socketmode.Event, client *socketmode.Client) {
		eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
		if !ok {
			return
		}
		client.Ack(*evt.Request)

		homeEvent, ok := eventsAPIEvent.InnerEvent.Data.(*slackevents.AppHomeOpenedEvent)
		if !ok {
			return
		}

		view := buildHomeView(store)
		_, err := c.botClient.PublishView(homeEvent.User, view, "")
		if err != nil {
			log.Warnf("slack: failed publishing App Home view: %s", err)
		}
	})

	log.Info("slack: starting Socket Mode event loop")
	if err := handler.RunEventLoop(); err != nil {
		log.Errorf("slack: Socket Mode event loop exited: %s", err)
	}
}

func buildHomeView(store *JobStateStore) slackgo.HomeTabViewRequest {
	states := store.All()

	var blocks []slackgo.Block

	headerText := slackgo.NewTextBlockObject(slackgo.MarkdownType, "*:calendar: On-Call Schedule Overview*", false, false)
	blocks = append(blocks, slackgo.NewSectionBlock(headerText, nil, nil))
	blocks = append(blocks, slackgo.NewDividerBlock())

	if len(states) == 0 {
		noDataText := slackgo.NewTextBlockObject(slackgo.MarkdownType, "_No sync jobs have run yet._", false, false)
		blocks = append(blocks, slackgo.NewSectionBlock(noDataText, nil, nil))
	}

	for _, s := range states {
		statusIcon := ":large_green_circle:"
		statusText := fmt.Sprintf("On shift: %s", strings.Join(s.OnShift, ", "))
		if s.Disabled {
			statusIcon = ":red_circle:"
			statusText = "Out of duty"
		} else if len(s.OnShift) == 0 {
			statusIcon = ":large_yellow_circle:"
			statusText = "No members matched"
		}

		line := fmt.Sprintf("%s  `@%s`   %s", statusIcon, s.SlackHandle, statusText)
		if !s.NextRun.IsZero() {
			line += fmt.Sprintf("   :alarm_clock: Next: %s", s.NextRun.UTC().Format("15:04 UTC"))
		}
		if s.Err != nil {
			line += fmt.Sprintf("   :warning: _%s_", s.Err)
		}

		rowText := slackgo.NewTextBlockObject(slackgo.MarkdownType, line, false, false)
		blocks = append(blocks, slackgo.NewSectionBlock(rowText, nil, nil))
	}

	updatedAt := slackgo.NewTextBlockObject(slackgo.MarkdownType,
		fmt.Sprintf("_Last updated: %s_", time.Now().UTC().Format("2006-01-02 15:04 UTC")), false, false)
	blocks = append(blocks, slackgo.NewDividerBlock())
	blocks = append(blocks, slackgo.NewSectionBlock(updatedAt, nil, nil))

	return slackgo.HomeTabViewRequest{
		Type:   slackgo.VTHomeTab,
		Blocks: slackgo.Blocks{BlockSet: blocks},
	}
}
