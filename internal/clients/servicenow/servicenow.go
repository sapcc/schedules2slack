package servicenow

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sapcc/schedules2slack/internal/config"

	//"github.com/sapcc/schedules2slack/internal/config"
	log "github.com/sirupsen/logrus"
)

// Client wraps the servicenow client.
type Client struct {
	cfg *config.ServiceNowConfig
}

// NewClient returns a new ServiceNowClient or an error.
func NewClient(cfg *config.ServiceNowConfig) (*Client, error) {
	c := &Client{
		cfg: cfg,
	}
	return c, nil
}

func (c *Client) ListOnCallUsers(s Schedule, layerSyncStyle config.SyncStyle) ([]Member, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: c.cfg.TLSconfig,
		},
	}

	for _, shift := range s.Shifts {
		var url = fmt.Sprintf("%s"+c.cfg.APIGetWhoIsOnCall, c.cfg.APIendpoint, shift.ID, shift.GroupID)
		log.Debug(url)
		response, err := client.Get(url)
		if err != nil {
			log.Error("error on servicenow request", err)
			return []Member{}, err
		}
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error parsing API response:", err)
			return []Member{}, err
		}
		log.Debug(string(body))

		var result WhoIsOnOnCallObjects
		if err := json.Unmarshal(body, &result); err != nil {
			log.Error("Can not unmarshal ServiceNow JSON")
			continue
		}
		fmt.Println(PrettyPrint(result))

		for _, who := range result.Members {
			for _, m := range s.Members {
				if who.UserID == m.UserID {
					s.OnOnCall = append(s.OnOnCall, m)
					log.Debug(fmt.Sprintf("%s is in oncall!", m.Name))
				} //else {
				//	log.Info(fmt.Sprintf("%s not found!", m.Name))
				//}
			}
		}
	}

	return s.OnOnCall, nil
}

func (c *Client) findMemberObjectByString(s Schedule, str string) (Member, error) {
	for _, m := range s.Members {
		if strings.Contains(str, m.Name) {
			//s.OnOnCall = append(s.OnOnCall, m)
			log.Debug(fmt.Sprintf("%s is in oncall!", m.Name))
			return m, nil
		}
	}
	log.Info(fmt.Sprintf("could find %s in group members", str))
	return Member{}, fmt.Errorf("could find %s in group members", str)
}

func (c *Client) ListSpans(s Schedule, groupID string) ([]Member, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: c.cfg.TLSconfig,
		},
	}

	today := time.Now().UTC()
	tomorrow := today.AddDate(0, 0, 1)
	todaystr := fmt.Sprintf("%04d-%02d-%02d", today.Year(), today.Month(), today.Day())
	tomorrowstr := fmt.Sprintf("%04d-%02d-%02d", tomorrow.Year(), tomorrow.Month(), tomorrow.Day())

	var url = fmt.Sprintf("%s"+c.cfg.APIGetSpans, c.cfg.APIendpoint, todaystr, groupID, tomorrowstr)
	log.Debug(url)
	response, err := client.Get(url)
	if err != nil {
		log.Error("error on servicenow request", err)
		return []Member{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error parsing API response:", err)
		return []Member{}, err
	}

	var result Spans
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		log.Error("Can not unmarshal ServiceNow JSON")
	}
	fmt.Println(PrettyPrint(result))

	// filter now
	result.Spans = filterByTimestamp(result.Spans)

	// get real Members
	var rM []Member
	for _, span := range result.Spans {
		var m, err = c.findMemberObjectByString(s, span.Title)
		if err == nil {
			//TODO: this is so dirty - we need to change that
			m.SlackDisplayValue = span.Title
			rM = append(rM, m)
		}
	}
	return rM, nil
}
func filterByTimestamp(spans []Span) []Span {
	var filteredSpans []Span

	for _, span := range spans {
		// Parse timestamp string into a time.Time object
		tsEnd, err := time.Parse(time.DateTime, span.End)
		if err != nil {
			// Handle parsing error if needed
			fmt.Printf("Error parsing timestamp for document ID %s: %v\n", span.Title, err)
			continue
		}
		tsBegin, err := time.Parse(time.DateTime, span.Start)
		if err != nil {
			// Handle parsing error if needed
			fmt.Printf("Error parsing timestamp for document ID %s: %v\n", span.Title, err)
			continue
		}

		// Compare the timestamp with the current time
		if tsEnd.After(time.Now()) && tsBegin.Before(time.Now()) && span.UserID != "" {
			filteredSpans = append(filteredSpans, span)
		}
	}

	return filteredSpans
}

// ListOnCallUsers returns the OnCall users being on shift now
func (c *Client) ListScheduleMember(groupID string) ([]Member, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: c.cfg.TLSconfig,
		},
	}

	var url = fmt.Sprintf("%s"+c.cfg.APIGetGroupMember, c.cfg.APIendpoint, groupID)
	log.Debug(url)
	response, err := client.Get(url)
	if err != nil {
		log.Error("error on servicenow request", err)
		return []Member{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error parsing API response:", err)
		return []Member{}, err
	}

	var result Members
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		log.Error("Can not unmarshal ServiceNow JSON")
	}
	fmt.Println(PrettyPrint(result))

	return result.Members, nil
}

// TeamMembers returns a schedule for the given name or an error.
func (c *Client) ListSchedules(scheduleID string) (Schedule, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: c.cfg.TLSconfig,
		},
	}

	var url = fmt.Sprintf("%s"+c.cfg.APIGetShifts, c.cfg.APIendpoint, scheduleID)
	log.Debug(url)
	response, err := client.Get(url)
	if err != nil {
		log.Error("error on servicenow request", err)
		return Schedule{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error parsing API response:", err)
		return Schedule{}, err
	}

	var result ScheduleShifts
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		log.Error("Can not unmarshal ServiceNow JSON")
	}
	fmt.Println(PrettyPrint(result))

	s := Schedule{
		Shifts:  result.Shifts,
		GroupID: scheduleID,
	}

	return s, nil
}
func PrettyPrint(i interface{}) string {
	s, err := json.MarshalIndent(i, "", "\t")
	if err != nil {
		return fmt.Sprintf("parsing error: %s", err)
	}
	return string(s)
}

/*
// listOnCallUsers returns unique list of users for OnCalls
func (c *Client) listOnCallUsers(onCalls []Member) (users []Member) {
	//opts := pd.GetUserOptions{Includes: []string{"contact_methods"}}

	distinctUsers := make(map[string]struct{})
	for _, u := range onCalls {
		if _, ok := distinctUsers[u.Name]; ok {
			// duplicate user
			log.Debugf("schedule: skipping duplicate onCall user %s", u.Name)
			continue
		}
		distinctUsers[u.Name] = struct{}{}

		users = append(users, u)
	}
	return users
}*/
