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

type Leave struct {
	Start Date `json:"start"`
	End   Date `json:"end"`
}

type TeamMember struct {
	SlackId  string
	JiraName string
	vacation []Leave
}

func (m TeamMember) IsAvailable(t time.Time) bool {
	for _, leave := range m.vacation {
		if leave.End.After(t) && leave.Start.Before(t) {
			return false
		}
	}
	return true
}

func (t *Team) Load(teamJSON, vacationJSON io.Reader) error {
	var members map[string]struct {
		SlackId    string   `json:"slack_id"`
		JiraName   string   `json:"jira_name"`
		Components []string `json:"jira_components"`
		vacation   []Leave
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

		for _, leave := range vacation {
			if m, ok := members[leave.Kerberos]; ok {
				m.vacation = append(m.vacation, Leave{
					Start: leave.Start,
					End:   leave.End,
				})
				members[leave.Kerberos] = m
			}
		}
	}

	team := make(map[string][]TeamMember)
	for _, member := range members {
		for _, component := range append(member.Components, "") {
			team[component] = append(team[component], TeamMember{SlackId: member.SlackId, JiraName: member.JiraName, vacation: member.vacation})
		}
	}

	*t = team
	return nil
}

// Specialists returns the team members assigned to that component. If the
// component isn't known, Specialists returns the whole team.
func (t Team) Specialists(component string) []TeamMember {
	if s, ok := t[component]; ok {
		return s
	}
	return t[""]
}

var (
	ErrEmptyTeam    error = fmt.Errorf("no team members to choose from")
	ErrVacatingTeam error = fmt.Errorf("no available specialist team members")
)

// RandomAvailable returns a random team member from the given slice, that
// isn't vacating. If the slice is empty, or if all members are vacating,
// RandomAvailable returns a non-nil error.
func RandomAvailable(team []TeamMember, t time.Time) (TeamMember, error) {
	if len(team) == 0 {
		return TeamMember{}, ErrEmptyTeam
	}

	availableMembers := make([]TeamMember, 0, len(team))
	for _, member := range team {
		if member.IsAvailable(t) {
			availableMembers = append(availableMembers, member)
		}
	}
	if len(availableMembers) == 0 {
		return TeamMember{}, ErrVacatingTeam
	}

	return availableMembers[rand.Intn(len(availableMembers))], nil
}

// MemberByJiraName returns the first team member found that has the given Jira
// name, and true. The boolean is false if no team member was found with that
// Jira name.
func (t Team) MemberByJiraName(jiraName string) (TeamMember, bool) {
	for _, members := range t {
		for _, member := range members {
			if member.JiraName == jiraName {
				return member, true
			}
		}
	}
	return TeamMember{}, false
}
