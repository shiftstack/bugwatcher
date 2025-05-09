package main

import (
	"strings"

	jira "github.com/andygrunwald/go-jira"
	"github.com/shiftstack/bugwatcher/pkg/query"
	"github.com/shiftstack/bugwatcher/pkg/slack"
)

func notification(issue jira.Issue, assignee TeamMember) string {
	var slackId string
	if strings.HasPrefix(assignee.SlackId, "!subteam^") {
		slackId = "<" + assignee.SlackId + ">"
	} else {
		slackId = "<@" + assignee.SlackId + ">"
	}

	var notification strings.Builder
	notification.WriteString(slackId)
	notification.WriteString(" you have been assigned triage of this bug: ")
	notification.WriteString(slack.Link(query.JiraBaseURL+"browse/"+issue.Key, issue.Key))
	return notification.String()
}
