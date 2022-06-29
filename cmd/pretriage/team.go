package main

import (
	"encoding/json"
	"math/rand"
)

type Team map[string][]TeamMember

type TeamMember struct {
	SlackId  string
	JiraName string
}

func (t *Team) UnmarshalJSON(data []byte) error {
	var members []struct {
		SlackId    string   `json:"slack_id"`
		JiraName   string   `json:"jira_name"`
		Components []string `json:"jira_components"`
	}

	if err := json.Unmarshal(data, &members); err != nil {
		return err
	}
	team := make(map[string][]TeamMember)
	for _, member := range members {
		for _, component := range append(member.Components, "") {
			team[component] = append(team[component], TeamMember{SlackId: member.SlackId, JiraName: member.JiraName})
		}
	}

	*t = team
	return nil
}

func (t Team) NewAssignee(component string) TeamMember {
	specialists := t[""]
	if s, ok := t[component]; ok {
		specialists = s
	}
	return specialists[rand.Intn(len(specialists))]
}
