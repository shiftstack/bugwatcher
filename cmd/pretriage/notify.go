package main

import (
	"fmt"
	"strings"

	jira "github.com/andygrunwald/go-jira"
	"github.com/shiftstack/bugwatcher/pkg/query"
	"github.com/shiftstack/bugwatcher/pkg/slack"
)

func notification(issue jira.Issue, slackId string) string {
	var notification strings.Builder
	notification.WriteByte('<')
	notification.WriteString(slackId)
	notification.WriteString("> you have been assigned triage of this bug: ")
	notification.WriteString(slack.Link(query.JiraBaseURL+"browse/"+issue.Key, issue.Key))
	return notification.String()
}

// cveGroupNotification creates a Slack message for a group of related CVE issues
func cveGroupNotification(group *CVEGroup, slackId string) string {
	var notification strings.Builder
	notification.WriteByte('<')
	notification.WriteString(slackId)
	notification.WriteString("> ")
	notification.WriteString(" You have been assigned to triage this CVE group: ")

	// Format: "CVE-2024-XXXX (Component): ISSUE-1 ISSUE-2 ISSUE-3"
	notification.WriteString(group.CVEID)
	notification.WriteString(" (")
	notification.WriteString(group.Component)
	notification.WriteString("): ")

	for i, issue := range group.Issues {
		if i > 0 {
			notification.WriteByte(' ')
		}
		notification.WriteString(slack.Link(query.JiraBaseURL+"browse/"+issue.Key, issue.Key))
	}

	if len(group.Issues) > 1 {
		notification.WriteString(fmt.Sprintf(" (%d related issues)", len(group.Issues)))
	}

	return notification.String()
}
