package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
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

type Team map[string][]TeamMember

type TeamMember struct {
	SlackId  string
	JiraName string
}

func (t *Team) Load(teamJSON, vacationJSON io.Reader) error {
	var members map[string]struct {
		SlackId    string   `json:"slack_id"`
		JiraName   string   `json:"jira_name"`
		Components []string `json:"jira_components"`
	}

	if err := json.NewDecoder(teamJSON).Decode(&members); err != nil {
		return fmt.Errorf("failed to unmarshal team: %w", err)
	}

	// Apply vacation
	{
		var vacation []struct {
			Kerberos string `json:"kerberos"`
			Start    Date   `json:"start"`
			End      Date   `json:"end"`
		}

		if err := json.NewDecoder(vacationJSON).Decode(&vacation); err != nil {
			return fmt.Errorf("failed to unmarshal vacation: %w", err)
		}

		now := time.Now()
		for _, v := range vacation {
			if v.End.After(now) && v.Start.Before(now) {
				delete(members, v.Kerberos)
			}
		}
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
