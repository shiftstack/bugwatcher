package main

import jira "github.com/andygrunwald/go-jira"

// triageCheck verifies one Triage condition.
// Returns true if the issue is triaged according to that particular condition.
// If triaged is false, msg contains the reason.
// err is non-nil in case of failure.
type triageCheck func(jira.Issue) (triaged bool, msg string, err error)

func docTextCheck(issue jira.Issue) (bool, string, error) {
	if issue.Fields.Unknowns["customfield_12310211"] == nil {
		return false, "the Release Note Text is missing", nil
	}
	return true, "", nil
}
