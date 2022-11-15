package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	jira "github.com/andygrunwald/go-jira"
)

func notification(issues []jira.Issue, assignee TeamMember) string {
	var notification strings.Builder
	notification.WriteString("<@" + assignee.SlackId + "> please check the doctext for these bugs:")
	for _, issue := range issues {
		notification.WriteString(fmt.Sprintf(" <%s|%s>", jiraBaseURL+"browse/"+issue.Key, issue.Key))
	}
	return notification.String()
}

func notify(slackHook string, slackClient *http.Client, issues []jira.Issue, assignee TeamMember) error {
	var msg bytes.Buffer
	err := json.NewEncoder(&msg).Encode(struct {
		LinkNames bool   `json:"link_names"`
		Text      string `json:"text"`
	}{
		LinkNames: true,
		Text:      notification(issues, assignee),
	})
	if err != nil {
		return fmt.Errorf("error while preparing the Slack notification for %s: %w", assignee.SlackId, err)
	}

	res, err := slackClient.Post(
		slackHook,
		"application/JSON",
		&msg,
	)
	if err != nil {
		return fmt.Errorf("error while sending a Slack notification for %s: %w", assignee, err)
	}

	io.Copy(io.Discard, res.Body)
	res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK, http.StatusNoContent, http.StatusAccepted:
	default:
		return fmt.Errorf("unexpected status code %q while sending a Slack notification for %s", res.Status, assignee)
	}

	return nil
}