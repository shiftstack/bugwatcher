package main

import (
	"fmt"

	jira "github.com/andygrunwald/go-jira"
)

// triageCheck verifies one Triage condition.
// Returns true if the issue is triaged according to that particular condition.
// If triaged is false, msg contains the reason.
// err is non-nil in case of failure.
type triageCheck func(jira.Issue) (triaged bool, msg string, err error)

func docTextCheck(issue jira.Issue) (bool, string, error) {
	// We must set the type and (optionally) the text for release notes. The
	// text must always be set unless the type is "No Doc Update".
	//
	// Release Note Type -> customfield_10785
	// Release Note Text -> customfield_10783

	if issue.Fields.Unknowns["customfield_10785"] != nil {
		releaseNoteType, ok := issue.Fields.Unknowns["customfield_10785"].(map[string]any)
		if !ok {
			return false, "", fmt.Errorf("failed to parse release note type for issue %s", issue.Key)
		}
		releaseNoteText := issue.Fields.Unknowns["customfield_10783"]

		if releaseNoteType["id"] == "12510" { // Release Note Not Required
			return true, "", nil
		}
		if releaseNoteText != nil {
			return true, "", nil
		}
	}

	return false, "the Release Note Text is missing", nil
}
