package main

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"

	jira "github.com/andygrunwald/go-jira"
	"github.com/shiftstack/bugwatcher/pkg/query"
)

const queryTriaged = query.ShiftStack + `AND labels = "Triaged"`

var JIRA_TOKEN = os.Getenv("JIRA_TOKEN")

func main() {
	ctx := context.Background()

	var jiraClient *jira.Client
	{
		var err error
		jiraClient, err = jira.NewClient(
			(&jira.BearerAuthTransport{Token: JIRA_TOKEN}).Client(),
			query.JiraBaseURL,
		)
		if err != nil {
			log.Fatalf("FATAL: error building a Jira client: %v", err)
		}
	}

	triageChecks := [...]triageCheck{
		priorityCheck,
		releaseBlockerCheck,
		testCoverageCheck,
	}

	var (
		found     int
		gotErrors bool
		wg        sync.WaitGroup
	)
	for issue := range query.SearchIssues(ctx, jiraClient, queryTriaged) {
		wg.Add(1)
		found++
		go func(issue jira.Issue) {
			defer wg.Done()
			reasons := make([]string, 0, len(triageChecks))

			for _, check := range triageChecks {
				triaged, msg, err := check(issue)
				if err != nil {
					log.Printf("WARNING: issue %s: Triage check failed: %v", issue.Key, err)
					continue
				}
				if !triaged {
					reasons = append(reasons, msg)
				}
			}

			if len(reasons) > 0 {
				log.Printf("INFO: Untriaging %q because %s", issue.Key, reasons)

				var comment strings.Builder
				comment.WriteString("Removing the Triaged label because:\n")
				for _, reason := range reasons {
					comment.WriteString("* " + reason + "\n")
				}

				if err := untriage(ctx, jiraClient, issue, comment.String()); err != nil {
					gotErrors = true
					log.Printf("ERROR: Failed to untriage %q: %v", issue.Key, err)
				}
			}
		}(issue)
	}
	wg.Wait()

	log.Printf("INFO: The query found %d bugs", found)

	if gotErrors {
		os.Exit(1)
	}
}

func init() {
	if JIRA_TOKEN == "" {
		log.Print("FATAL: Required environment variable not found: JIRA_TOKEN")
		os.Exit(64)
	}
}
