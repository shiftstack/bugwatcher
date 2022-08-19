package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	jira "github.com/andygrunwald/go-jira"
)

// untriage removes the Triage label and comments on the issue
func untriage(ctx context.Context, jiraClient *jira.Client, issue jira.Issue, comment string) error {
	// Remove the Triaged label
	{
		res, err := jiraClient.Issue.UpdateIssueWithContext(ctx, issue.ID, map[string]interface{}{
			"update": map[string]interface{}{
				"labels": json.RawMessage(`[{"remove":"Triaged"}]`),
			},
		})
		if err != nil {
			return fmt.Errorf("failed setting issue %q as non triaged: %w", issue.Key, err)
		}

		io.Copy(os.Stderr, res.Body)
		res.Body.Close()

		switch res.StatusCode {
		case http.StatusOK, http.StatusNoContent, http.StatusAccepted:
		default:
			return fmt.Errorf("unexpected status code %q while setting issue %q as non triaged", res.Status, issue.Key)
		}
	}

	// Add an explanatory comment
	if comment != "" {
		_, res, err := jiraClient.Issue.AddCommentWithContext(ctx, issue.ID, &jira.Comment{
			Body: comment,
		})
		if err != nil {
			return fmt.Errorf("failed commenting issue %q: %w", issue.Key, err)
		}

		io.Copy(os.Stderr, res.Body)
		res.Body.Close()

		switch res.StatusCode {
		case http.StatusOK, http.StatusNoContent, http.StatusAccepted, http.StatusCreated:
		default:
			return fmt.Errorf("unexpected status code %q while commenting issue %q", res.Status, issue.Key)
		}
	}

	return nil
}
