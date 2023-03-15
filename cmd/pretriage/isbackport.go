package main

import (
	jira "github.com/andygrunwald/go-jira"
)

const isBlockedBy = "12310720"

// returns the first detected blocking issue. Note that the returned Jira issue
// only contains the "assignee" field, for optimisation reasons; this can be
// changed when another field is useful.
// The returned error comes from the Jira client and can only be non-nil when a
// parent exists.
func backportParent(client *jira.Client, issue jira.Issue) (jira.Issue, bool, error) {
	for _, link := range issue.Fields.IssueLinks {
		if link.Type.ID == isBlockedBy && link.InwardIssue != nil {
			parent, _, err := client.Issue.Get(link.InwardIssue.ID, &jira.GetQueryOptions{Fields: "assignee"})
			if err != nil {
				return jira.Issue{}, true, err
			}
			return *parent, true, nil
		}
	}
	return jira.Issue{}, false, nil
}
