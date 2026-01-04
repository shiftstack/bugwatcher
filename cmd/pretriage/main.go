package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/shiftstack/bugwatcher/pkg/jiraclient"
	"github.com/shiftstack/bugwatcher/pkg/query"
	"github.com/shiftstack/bugwatcher/pkg/slack"
	"github.com/shiftstack/bugwatcher/pkg/team"
)

const queryUntriaged = query.ShiftStack + `AND ( assignee is EMPTY OR assignee = "shiftstack-dev@redhat.com" ) AND (labels not in ("Triaged") OR labels is EMPTY)`
const queryARTReconciliation = query.ShiftStack + `AND labels in ("art:reconciliation")
	AND (
		priority is EMPTY OR
		"Release Note Type" is EMPTY OR
		"Test Coverage" is EMPTY
	)
`

var (
	SLACK_HOOK = os.Getenv("SLACK_HOOK")
	JIRA_TOKEN = os.Getenv("JIRA_TOKEN")
	PEOPLE     = os.Getenv("PEOPLE")
)

func main() {
	ctx := context.Background()

	var people, triagers []team.Person
	{
		var err error
		people, err = team.Load(strings.NewReader(PEOPLE))
		if err != nil {
			log.Fatalf("error fetching team information: %v", err)
		}

		triagers = make([]team.Person, 0, len(people))

		now := time.Now()
		for _, p := range people {
			if p.BugTriage && p.IsAvailable(now) {
				triagers = append(triagers, p)
			}
		}
		if len(triagers) < 1 {
			log.Fatal("no triagers available")
		}
	}

	jiraClient, err := jiraclient.NewWithToken(query.JiraBaseURL, JIRA_TOKEN)
	if err != nil {
		log.Fatalf("error building a Jira client: %v", err)
	}

	var wg sync.WaitGroup
	var gotErrors bool

	log.Print("pre-setting any necessary fields for the ART reconciliation bugs...")
	for issue := range query.SearchIssues(ctx, jiraClient, queryARTReconciliation) {
		wg.Add(1)
		go func(issue jira.Issue) {
			defer wg.Done()

			log.Printf("Updating issue %q", issue.Key)

			// These changes are idempotent, so we don't need to check for the current value
			updates := map[string]any{
				"update": map[string]any{
					"priority": []map[string]any{
						{
							"set": map[string]any{"name": "Normal"},
						},
					},
					"customfield_12320850": []map[string]any{ // Release Note Type
						{
							"set": map[string]any{"value": "Release Note Not Required"},
						},
					},
					"customfield_12320940": []map[string]any{ // Test Coverage
						{
							"set": []map[string]any{
								{"value": "-"},
							},
						},
					},
				},
			}
			if err := update(jiraClient, issue, updates); err != nil {
				gotErrors = true
				log.Print(err)
				return
			}
		}(issue)
	}
	wg.Wait()

	if gotErrors {
		os.Exit(1)
	}

	slackClient := slack.New()

	log.Print("Running the actual triage assignment...")

	// Collect all issues first, separating CVEs from regular bugs
	var cveIssues []jira.Issue
	var regularIssues []jira.Issue

	for issue := range query.SearchIssues(ctx, jiraClient, queryUntriaged) {
		if isVulnerability(issue) {
			cveIssues = append(cveIssues, issue)
		} else {
			regularIssues = append(regularIssues, issue)
		}
	}

	// Process CVE issues: group by CVE ID + Component, assign group together
	if len(cveIssues) > 0 {
		log.Printf("Found %d CVE issues, grouping...", len(cveIssues))
		cveGroups := GroupCVEIssues(cveIssues)
		log.Printf("Grouped into %d CVE groups", len(cveGroups))

		for key, group := range cveGroups {
			assignee := &triagers[rand.Intn(len(triagers))]

			log.Printf("Assigning CVE group %q (%d issues) to %q",
				key, len(group.Issues), censorEmail(assignee.Jira))

			// Assign all issues in the group to the same person
			for _, issue := range group.Issues {
				wg.Add(1)
				go func(issue jira.Issue) {
					defer wg.Done()
					if err := assign(jiraClient, issue, assignee.Jira); err != nil {
						gotErrors = true
						log.Print(err)
					}
				}(issue)
			}
			wg.Wait()

			// Send single grouped notification
			if err := slackClient.Send(SLACK_HOOK, cveGroupNotification(group, assignee.Slack)); err != nil {
				gotErrors = true
				log.Print(err)
			}
		}
	}

	// Process regular bugs: existing individual assignment flow
	for _, issue := range regularIssues {
		wg.Add(1)
		go func(issue jira.Issue) {
			defer wg.Done()
			assignee := &triagers[rand.Intn(len(triagers))]
			if parent, isBackport, err := backportParent(jiraClient, issue); isBackport {
				if err != nil {
					log.Print(err)
					gotErrors = true
					return
				}
				if parent.Fields.Assignee != nil {
					log.Printf("Issue %q has parent %q, which is assigned to %q", issue.Key, parent.Key, censorEmail(parent.Fields.Assignee.Name))
					if p, ok := team.PersonByJiraName(triagers, parent.Fields.Assignee.Name); ok {
						assignee = &p
					}
				}
			}

			log.Printf("Assigning issue %q to %q", issue.Key, censorEmail(assignee.Jira))

			if err := assign(jiraClient, issue, assignee.Jira); err != nil {
				gotErrors = true
				log.Print(err)
				return
			}

			if err := slackClient.Send(SLACK_HOOK, notification(issue, assignee.Slack)); err != nil {
				gotErrors = true
				log.Print(err)
				return
			}
		}(issue)
	}
	wg.Wait()

	if gotErrors {
		os.Exit(1)
	}
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)

	ex_usage := false
	if SLACK_HOOK == "" {
		ex_usage = true
		log.Print("Required environment variable not found: SLACK_HOOK")
	}

	if JIRA_TOKEN == "" {
		ex_usage = true
		log.Print("Required environment variable not found: JIRA_TOKEN")
	}

	if PEOPLE == "" {
		ex_usage = true
		log.Print("Required environment variable not found: PEOPLE")
	}

	if ex_usage {
		log.Print("Exiting.")
		os.Exit(64)
	}
}
