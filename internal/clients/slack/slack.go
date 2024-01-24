package slack

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/sapcc/schedules2slack/internal/clients/servicenow"
	"github.com/sapcc/schedules2slack/internal/config"
	log "github.com/sirupsen/logrus"
	slackgo "github.com/slack-go/slack"
)

type Client struct {
	botClient     *slackgo.Client     // slack client for bot
	userClient    *slackgo.Client     // slack client for user
	users         []slackgo.User      // list of all slack users in the workspace
	groups        []slackgo.UserGroup // list of all groups in the workspace
	infoChannel   *slackgo.Channel    // channel used to post info messages to
	infoChannelID string              // ID of info channel
}

// newAPIClient returns token specific slack client object and tests auth
func newAPIClient(token string, options ...slackgo.Option) (*slackgo.Client, error) {
	options = append(options, slackgo.OptionDebug(false))
	c := slackgo.New(token, options...)
	_, err := c.AuthTest()
	return c, err
}

// PostBlocksMessage takes the blocks and sends them to the default info channel
func (c *Client) PostBlocksMessage(blocks ...slackgo.Block) error {
	opts := slackgo.MsgOptionBlocks(blocks...)
	return c.PostMessage(opts)
}

// PostMessage takes the message options sends it to the info channel
func (c *Client) PostMessage(opts slackgo.MsgOption) error {
	if _, _, err := c.botClient.PostMessage(c.infoChannel.ID, opts); err != nil {
		return fmt.Errorf("slack: failed posting message: %w", err)
	}
	log.Debug("slack: message successfully sent to channel ", c.infoChannel.Name)
	return nil
}

// NewClient returns a new slackclient with intialized bot & user client and loaded masterdata
func NewClient(cfg *config.SlackConfig) (*Client, error) {
	bot, err := newAPIClient(cfg.BotSecurityToken)
	if err != nil {
		return nil, fmt.Errorf("slack: failed creating bot client: %w", err)
	}
	user, err := newAPIClient(cfg.UserSecurityToken)
	if err != nil {
		return nil, fmt.Errorf("slack: failed creating user client: %w", err)
	}

	c := &Client{
		botClient:     bot,
		userClient:    user,
		infoChannelID: cfg.InfoChannelID,
	}

	err = c.LoadMasterData()
	if err != nil {
		return nil, fmt.Errorf("slack: failed loading masterdata: %w", err)
	}
	return c, nil
}

// LoadMasterData singleton master data to speed up
func (c *Client) LoadMasterData() (err error) {

	var cI = slackgo.GetConversationInfoInput{
		ChannelID:         c.infoChannelID,
		IncludeLocale:     true,
		IncludeNumMembers: true,
	}
	slackChannelsTemp, err := c.botClient.GetConversationInfo(&cI)
	if err != nil {
		return fmt.Errorf("slack: failed retrieving info channel '%s': %w", c.infoChannelID, err)
	}
	c.infoChannel = slackChannelsTemp

	slackUserListTemp, err := c.botClient.GetUsers()
	if err != nil {
		return fmt.Errorf("slack: failed retrieving users: %w", err)
	}

	slackGrpsTemp, err := c.botClient.GetUserGroups(slackgo.GetUserGroupsOptionIncludeUsers(true))
	if err != nil {
		return fmt.Errorf("slack: failed retrieving user groups: %w", err)
	}

	var mutex = &sync.Mutex{}
	mutex.Lock()
	c.users = slackUserListTemp
	c.groups = slackGrpsTemp
	mutex.Unlock()
	log.Debug("slack: masterdata successfully updated")
	return nil
}

// GetSlackGroup requests existing Group for given name
func (c *Client) GetSlackGroup(slackGroupHandle string) (slackgo.UserGroup, error) {
	if slackGroupHandle == "" {
		return slackgo.UserGroup{}, fmt.Errorf("slack: finding group failed, empty handle")
	}

	// get the group we are interested in
	var targetGroup slackgo.UserGroup
	for _, g := range c.groups {
		if strings.EqualFold(g.Handle, slackGroupHandle) {
			targetGroup = g
			break
		}
	}

	if targetGroup.Handle == "" {
		return slackgo.UserGroup{}, fmt.Errorf("slack: finding group handle '%s' failed. check config", slackGroupHandle)
	}

	return targetGroup, nil
}

// MatchPDUsers returns slack users matching the given users
func (c *Client) MatchUsers(User []servicenow.Member) ([]slackgo.User, error) {
	// if no pdUsers given, we don't need to filter
	if User == nil {
		log.Warn("empty schedule member list given!")
		return nil, fmt.Errorf("empty PD user list; check shift schedule")
	}

	// get all SLACK User Ids which are in our PD Group - some people are not in slack
	userList := c.matchToSlackUsers(User)

	log.Infof("slack: found #%v matching slack user(s) for #%v user(s) in schedule group", len(userList), len(User))
	return userList, nil
}

// AddToGroup sets an array of Slack User to an Slack Group (found by name), returns true if noop
func (c *Client) AddToGroup(groupHandle string, slackUsers []slackgo.User, dryrun bool) (noChange bool, err error) {
	noChange = true

	// get the group we are interested in
	userGroupBefore, err := c.GetSlackGroup(groupHandle)
	if err != nil {
		return true, fmt.Errorf("slack: retrieving slack group '%s' failed: %w", groupHandle, err)
	}

	if len(slackUsers) == 0 {
		return true, fmt.Errorf("slack: user list empty; no update done")
	}

	// we need a list of IDs
	var slackUserIds []string

	if len(slackUsers) == len(userGroupBefore.Users) {
		for _, user := range slackUsers {
			if !groupContainsUser(userGroupBefore.Users, user) {
				slackUserIds = append(slackUserIds, user.ID)
				noChange = false
				continue
			}
			slackUserIds = append(slackUserIds, user.ID)
		}
	} else {
		noChange = false
		for _, u := range slackUsers {
			slackUserIds = append(slackUserIds, u.ID)
		}
	}

	if noChange {
		slackUserIds = nil
	}

	var userGroupAfter slackgo.UserGroup
	if !dryrun && !noChange {
		userGroupAfter, err = c.userClient.UpdateUserGroupMembers(userGroupBefore.ID, strings.Join(slackUserIds, ","))
		if err != nil {
			return noChange, fmt.Errorf("slack: writing changes for user group %s[%s] failed: %s", userGroupBefore.Name, userGroupBefore.ID, err.Error())
		}

		log.Infof("slack: updated %s successfully", userGroupAfter.Name)

		if userGroupAfter.DateDelete.String() == "" {
			_, err = c.userClient.EnableUserGroup(userGroupAfter.ID)
			if err != nil {
				return noChange, fmt.Errorf("slack: enabling user group %s[%s] failed: %s", userGroupBefore.Name, userGroupBefore.ID, err.Error())
			}
		}
	} else {
		userGroupAfter = userGroupBefore
	}

	var removedUsers []string
	if !noChange {
		for _, u := range userGroupBefore.Users {
			var removed = true
			for _, n := range userGroupAfter.Users {
				if u == n {
					removed = false
					break
				}
			}
			if removed {
				removedUsers = append(removedUsers, u)
			}
		}
	}

	if dryrun {
		log.Infof("slack: dry run. no changes executed.")
	}
	log.Infof("slack: added %v to and removed %v from group '%s'(%d member(s))", slackUserIds, removedUsers, userGroupAfter.Name, len(userGroupAfter.Users))

	return noChange, nil
}

func (c *Client) DisableGroup(groupID string) error {
	userGroup, err := c.userClient.DisableUserGroup(groupID)
	if err != nil {
		return err
	}
	log.Infof("slack: disabled slack user group %s[%s]", userGroup.Name, userGroup.ID)
	return nil
}

// groupContainsUser returns true user is contained in the groupUserIDs
func groupContainsUser(groupUserIDs []string, user slackgo.User) bool {
	if len(groupUserIDs) == 0 {
		return false
	}
	for _, id := range groupUserIDs {
		if id == user.ID {
			return true
		}
	}
	return false
}

// matchPDToSlackUsers returns a list of valid Slack users that match the list of users
func (c *Client) matchToSlackUsers(members []servicenow.Member) []slackgo.User {
	var matchedSlackUsers []slackgo.User
	for _, m := range members {

		if m.Name == "" {
			log.Infof("skipping user %s, name empty", m.Name)
			continue
		}

		for _, u := range c.users {
			if u.Deleted {
				continue
			}
			//if strings.EqualFold(m.Name, u.Profile.DisplayName) {

			if extractEmployeeId(m.Name) == strings.ToUpper(u.Name) {
				log.Debug(fmt.Sprintf("Found Slack user %s for schedule member %s", u.Profile.DisplayName, m.Name))
				matchedSlackUsers = append(matchedSlackUsers, u)
				continue
			}
		}
		log.Debug(fmt.Sprintf("Didn't found Slack user for schedule member %s", m.Name))
	}
	return matchedSlackUsers
}

/*
This makes it unusable outside SAP
need a better way to match Names
*/
func extractEmployeeId(input string) string {
	// Define the regular expression with a capturing group
	re := regexp.MustCompile(`([CID]\d{6})`)

	// Find the first match
	match := re.FindStringSubmatch(strings.ToUpper(input))

	// Check if there is a match
	if len(match) > 1 {
		log.Debug(fmt.Sprintf("regex employee Id for input `%s`: %s", input, match[1]))
		return match[1]
	}
	log.Debug(fmt.Sprintf("regex employee Id for input `%s`: no match", input))
	return ""
}
