package main

import (
	"strings"

	jira "github.com/andygrunwald/go-jira"
	"github.com/shiftstack/bugwatcher/pkg/query"
	"github.com/shiftstack/bugwatcher/pkg/slack"
)

func notification(issues []jira.Issue, slackId string) string {
	var notification strings.Builder
	notification.WriteByte('<')
	notification.WriteString(slackId)
	notification.WriteString("> please triage these bugs:")
	for _, issue := range issues {
		notification.WriteByte(' ')
		notification.WriteString(slack.Link(query.JiraBaseURL+"browse/"+issue.Key, issue.Key))
	}
	return notification.String()
}
