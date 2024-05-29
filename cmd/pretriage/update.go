package main

import (
	"fmt"
	"io"
	"net/http"

	jira "github.com/andygrunwald/go-jira"
)

func update(jiraClient *jira.Client, issue jira.Issue, updates map[string]interface{}) error {
	res, err := jiraClient.Issue.UpdateIssue(issue.ID, updates)
	if err != nil && res == nil {
		// we only error out early if there's no response to work with
		return fmt.Errorf("error while updating bug %s: %w", issue.Key, err)
	}

	var body string
	if res != nil {
		// we don't check errors since this is best effort
		bodyBytes, _ := io.ReadAll(res.Body)
		body = string(bodyBytes)
		res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK, http.StatusNoContent, http.StatusAccepted:
	default:
		return fmt.Errorf("unexpected status code %q from Jira while updating bug %s: err=%w body=%s", res.Status, issue.Key, err, body)
	}

	return nil
}
