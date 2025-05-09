package main

import (
	"strings"

	jira "github.com/andygrunwald/go-jira"
	"github.com/shiftstack/bugwatcher/pkg/query"
	"github.com/shiftstack/bugwatcher/pkg/slack"
)

func notification(issues []jira.Issue, assignee TeamMember) string {
	var slackId string
	if strings.HasPrefix(assignee.SlackId, "!subteam^") {
		slackId = "<" + assignee.SlackId + ">"
	} else {
		slackId = "<@" + assignee.SlackId + ">"
	}

	var notification strings.Builder
	notification.WriteString(slackId)
	notification.WriteString(" please triage these bugs:")
	for _, issue := range issues {
		notification.WriteByte(' ')
		notification.WriteString(slack.Link(query.JiraBaseURL+"browse/"+issue.Key, issue.Key))
	}
	return notification.String()
}
