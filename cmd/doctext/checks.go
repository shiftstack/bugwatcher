package main

import jira "github.com/andygrunwald/go-jira"

// triageCheck verifies one Triage condition.
// Returns true if the issue is triaged according to that particular condition.
// If triaged is false, msg contains the reason.
// err is non-nil in case of failure.
type triageCheck func(jira.Issue) (triaged bool, msg string, err error)

func docTextCheck(issue jira.Issue) (bool, string, error) {
	// We must set the type and (optionally) the text for release notes. The
	// text must always be set unless the type is "No Doc Update".
	//
	// Release Note Type -> customfield_12320850
	// Release Note Text -> customfield_12317313
	if issue.Fields.Unknowns["customfield_12320850"] != nil {
		if issue.Fields.Unknowns["customfield_12320850"] == "No Doc Update" {
			return true, "", nil
		}
		if issue.Fields.Unknowns["customfield_12317313"] != nil {
			return true, "", nil
		}
	}
	return false, "the Release Note Text is missing", nil
}
