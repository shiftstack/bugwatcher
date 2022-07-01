package main

import (
	"fmt"
	"io"
	"net/http"

	jira "github.com/andygrunwald/go-jira"
)

func assign(jiraClient *jira.Client, issue jira.Issue, assignee TeamMember) error {
	res, err := jiraClient.Issue.UpdateAssignee(issue.ID, &jira.User{Name: assignee.JiraName})
	if err != nil {
		return fmt.Errorf("error while assigning bug %s: %w", issue.Key, err)
	}

	io.Copy(io.Discard, res.Body)
	res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK, http.StatusNoContent, http.StatusAccepted:
	default:
		return fmt.Errorf("unexpected status code %q from Jira while assigning bug %s: %w", res.Status, issue.Key, err)
	}

	return nil
}
