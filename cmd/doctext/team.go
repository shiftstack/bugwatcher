package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type Date time.Time

func (d *Date) UnmarshalJSON(src []byte) error {
	var datestring string
	if err := json.Unmarshal(src, &datestring); err != nil {
		return err
	}

	t, err := time.Parse("2006-01-02", datestring)
	if err != nil {
		return err
	}

	*d = Date(t)
	return nil
}

func (d Date) Before(t time.Time) bool { return time.Time(d).Before(t) }
func (d Date) After(t time.Time) bool  { return time.Time(d).After(t) }

// Team is a map of JiraName to TeamMember
type Team map[string]TeamMember

type TeamMember struct {
	SlackId  string
	JiraName string
}

func (t *Team) Load(teamJSON io.Reader) error {
	var members map[string]struct {
		SlackId  string `json:"slack_id"`
		JiraName string `json:"jira_name"`
	}

	if err := json.NewDecoder(teamJSON).Decode(&members); err != nil {
		return fmt.Errorf("failed to unmarshal team: %w", err)
	}

	// team is a map of JiraID to TeamMember
	team := map[string]TeamMember{
		// ID of openstack-dev-team user group
		// https://api.slack.com/reference/surfaces/formatting#mentioning-groups
		"team": {SlackId: "!subteam^SKW6QC31Q", JiraName: "team"},
	}

	for _, member := range members {
		team[member.JiraName] = TeamMember{SlackId: member.SlackId, JiraName: member.JiraName}
	}

	*t = team
	return nil
}
