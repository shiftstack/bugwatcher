package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	jira "github.com/andygrunwald/go-jira"
	"github.com/shiftstack/bugwatcher/pkg/query"
)

func notification(issue jira.Issue, assignee TeamMember) string {
	var slackId string
	if strings.HasPrefix(assignee.SlackId, "!subteam^") {
		slackId = "<" + assignee.SlackId + ">"
	} else {
		slackId = "<@" + assignee.SlackId + ">"
	}

	var notification strings.Builder
	notification.WriteString(slackId + " you have been assigned triage of this bug:")
	notification.WriteString(fmt.Sprintf(" <%s|%s>", query.JiraBaseURL+"browse/"+issue.Key, issue.Key))
	return notification.String()
}

func notify(slackHook string, slackClient *http.Client, issue jira.Issue, assignee TeamMember) error {
	var msg bytes.Buffer
	err := json.NewEncoder(&msg).Encode(struct {
		LinkNames bool   `json:"link_names"`
		Text      string `json:"text"`
	}{
		LinkNames: true,
		Text:      notification(issue, assignee),
	})
	if err != nil {
		return fmt.Errorf("error while preparing the Slack notification for bug %s: %w", issue.Key, err)
	}

	res, err := slackClient.Post(
		slackHook,
		"application/JSON",
		&msg,
	)
	if err != nil {
		return fmt.Errorf("error while sending a Slack notification for bug %s: %w", issue.Key, err)
	}

	io.Copy(io.Discard, res.Body)
	res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK, http.StatusNoContent, http.StatusAccepted:
	default:
		return fmt.Errorf("unexpected status code %q while sending a Slack notification for bug %s", res.Status, issue.Key)
	}

	return nil
}
