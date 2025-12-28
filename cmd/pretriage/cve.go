package main

import (
	"fmt"
	"strings"

	jira "github.com/andygrunwald/go-jira"
)

// CVEFieldID is the JIRA custom field ID for the CVE identifier
const CVEFieldID = "customfield_12324749"

// CVEGroup represents a group of related CVE issues
type CVEGroup struct {
	CVEID     string
	Component string
	Issues    []jira.Issue
}

// isVulnerability checks if an issue is of type "Vulnerability"
func isVulnerability(issue jira.Issue) bool {
	if issue.Fields == nil || issue.Fields.Type.Name == "" {
		return false
	}
	return issue.Fields.Type.Name == "Vulnerability"
}

// extractCVEID extracts the CVE identifier from an issue's custom field
func extractCVEID(issue jira.Issue) string {
	if issue.Fields == nil || issue.Fields.Unknowns == nil {
		return ""
	}

	if cveValue, ok := issue.Fields.Unknowns[CVEFieldID]; ok {
		if cveStr, ok := cveValue.(string); ok {
			return strings.TrimSpace(cveStr)
		}
	}
	return ""
}

// extractComponent extracts the first component name from an issue
func extractComponent(issue jira.Issue) string {
	if issue.Fields == nil || len(issue.Fields.Components) == 0 {
		return "unknown"
	}
	return issue.Fields.Components[0].Name
}

// groupKey creates a unique key for grouping: "CVE-ID|Component"
func groupKey(issue jira.Issue) string {
	cveID := extractCVEID(issue)
	component := extractComponent(issue)
	return fmt.Sprintf("%s|%s", cveID, component)
}

// GroupCVEIssues groups issues by CVE ID + Component
// Returns a map where key is "CVE-ID|Component" and value is the CVEGroup
func GroupCVEIssues(issues []jira.Issue) map[string]*CVEGroup {
	groups := make(map[string]*CVEGroup)

	for _, issue := range issues {
		key := groupKey(issue)

		if groups[key] == nil {
			groups[key] = &CVEGroup{
				CVEID:     extractCVEID(issue),
				Component: extractComponent(issue),
				Issues:    []jira.Issue{issue},
			}
		} else {
			groups[key].Issues = append(groups[key].Issues, issue)
		}
	}

	return groups
}
