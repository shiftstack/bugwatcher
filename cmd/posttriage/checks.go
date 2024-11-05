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

func priorityCheck(issue jira.Issue) (bool, string, error) {
	// If a bug has been closed as a non-bug, we shouldn't insist on a priority.
	// Taken from https://issues.redhat.com/rest/api/2/resolution
	switch issue.Fields.Resolution.Name {
	case "Won't Do", "Cannot Reproduce", "Can't Do", "Duplicate", "Not a bug", "Obsolete":
		return true, "", nil
	}
	if issue.Fields.Priority == nil || issue.Fields.Priority.Name == "Undefined" {
		return false, "the Priority assessment is missing", nil
	}
	return true, "", nil
}

const (
	releaseBlockerNone     releaseBlocker = ""
	releaseBlockerApproved releaseBlocker = "Approved"
	releaseBlockerProposed releaseBlocker = "Proposed"
	releaseBlockerRejected releaseBlocker = "Rejected"
)

type releaseBlocker string

// ReleaseBlockerFromIssue parses releaseBlocker information from a Jira issue.
func ReleaseBlockerFromIssue(issue jira.Issue) (releaseBlocker, error) {
	// https://confluence.atlassian.com/jirakb/how-to-find-any-custom-field-s-ids-744522503.html
	if issue.Fields.Unknowns["customfield_12319743"] == nil {
		return releaseBlockerNone, nil
	}

	releaseBlockerMap, ok := issue.Fields.Unknowns["customfield_12319743"].(map[string]any)
	if !ok {
		return releaseBlockerNone, fmt.Errorf("failed to parse (not a map)")
	}

	// https://confluence.atlassian.com/jirakb/how-to-retrieve-available-options-for-a-multi-select-customfield-via-jira-rest-api-815566715.html
	switch releaseBlockerMap["id"] {
	case "25755":
		return releaseBlockerApproved, nil
	case "25756":
		return releaseBlockerProposed, nil
	case "25757":
		return releaseBlockerRejected, nil
	default:
		return releaseBlockerNone, fmt.Errorf("unknown Release Blocker value: %s", releaseBlockerMap["id"])
	}

}

func releaseBlockerCheck(issue jira.Issue) (bool, string, error) {
	rb, err := ReleaseBlockerFromIssue(issue)
	if err != nil {
		return false, "", fmt.Errorf("failed to parse Release Blocker: %s", err)
	}
	if rb == releaseBlockerProposed {
		return false, "the issue is a proposed release blocker", nil
	}
	return true, "", nil
}

const (
	testCoverageNone       testCoverage = 0
	testCoverageAutomated  testCoverage = '+'
	testCoverageManual     testCoverage = '-'
	testCoverageNoCoverage testCoverage = '?'
)

type testCoverage byte

// TestCoverageFromIssue parses testCoverage information from a Jira issue.
func TestCoverageFromIssue(issue jira.Issue) (testCoverage, error) {
	// https://confluence.atlassian.com/jirakb/how-to-find-any-custom-field-s-ids-744522503.html
	if issue.Fields.Unknowns["customfield_12320940"] == nil {
		return testCoverageNone, nil
	}

	testCoverageSlice, err := issue.Fields.Unknowns.Slice("customfield_12320940")
	if err != nil {
		return testCoverageNone, err
	}
	if len(testCoverageSlice) < 1 {
		return testCoverageNone, nil
	}

	testCoverageMap, ok := testCoverageSlice[0].(map[string]any)
	if !ok {
		return testCoverageNone, fmt.Errorf("failed to parse (not a slice of maps)")
	}

	// https://confluence.atlassian.com/jirakb/how-to-retrieve-available-options-for-a-multi-select-customfield-via-jira-rest-api-815566715.html
	switch testCoverageMap["id"] {
	case "27678":
		return testCoverageAutomated, nil
	case "27679":
		return testCoverageManual, nil
	case "27680":
		return testCoverageNoCoverage, nil
	default:
		return testCoverageNone, fmt.Errorf("unknown test coverage value: %s", testCoverageMap["id"])
	}
}

func testCoverageCheck(issue jira.Issue) (bool, string, error) {
	tc, err := TestCoverageFromIssue(issue)
	if err != nil {
		return false, "", fmt.Errorf("failed to parse Test coverage: %w", err)
	}
	if tc == testCoverageNone {
		return false, "the Test coverage assessment is missing", nil
	}
	return true, "", nil
}
