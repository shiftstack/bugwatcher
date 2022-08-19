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
	if issue.Fields.Priority == nil || issue.Fields.Priority.Name == "Undefined" {
		return false, "the Priority assessment is missing", nil
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
		return testCoverageNone, fmt.Errorf("failed to parse Test Coverage")
	}

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
