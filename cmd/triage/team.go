package main

import (
	"encoding/json"
	"fmt"
	"io"
)

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
		"team": {SlackId: "openstack-dev-team", JiraName: "team"},
	}
	for _, member := range members {
		team[member.JiraName] = TeamMember{SlackId: member.SlackId, JiraName: member.JiraName}
	}

	*t = team
	return nil
}
