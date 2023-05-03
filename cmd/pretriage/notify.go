package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	jira "github.com/andygrunwald/go-jira"
	"github.com/shiftstack/bugwatcher/pkg/query"
)

func notify(slackHook string, slackClient *http.Client, issue jira.Issue, assignee TeamMember) error {
	var msg bytes.Buffer
	err := json.NewEncoder(&msg).Encode(struct {
		LinkNames bool   `json:"link_names"`
		Text      string `json:"text"`
	}{
		LinkNames: true,
		Text:      "<@" + assignee.SlackId + "> you have been assigned the triage of this bug: " + query.JiraBaseURL + "browse/" + issue.Key,
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
