package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	jira "github.com/andygrunwald/go-jira"
	"github.com/shiftstack/bugwatcher/pkg/query"
)

const queryTriaged = query.ShiftStack + `AND status in ("Release Pending", Verified, ON_QA) AND "Release Note Text" is EMPTY`

var (
	BUGWATCHER_DEBUG  = os.Getenv("BUGWATCHER_DEBUG")
	SLACK_HOOK        = os.Getenv("SLACK_HOOK")
	JIRA_TOKEN        = os.Getenv("JIRA_TOKEN")
	TEAM_MEMBERS_DICT = os.Getenv("TEAM_MEMBERS_DICT")
)

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
		docTextCheck,
	}

	var (
		found     int
		gotErrors bool
		wg        sync.WaitGroup
	)
	issues := make(map[string][]jira.Issue)
	for issue := range searchIssues(ctx, jiraClient, queryTriaged) {
		wg.Add(1)
		found++
		go func(issue jira.Issue) {
			defer wg.Done()
			reasons := make([]string, 0, len(triageChecks))

			for _, check := range triageChecks {
				triaged, msg, err := check(issue)
				if err != nil {
					log.Printf("WARNING: DocText check failed: %v", err)
					continue
				}
				if !triaged {
					reasons = append(reasons, msg)
				}
			}

			if len(reasons) > 0 {
				log.Printf("INFO: Missing DocText for %q", issue.Key)
				var assignee string
				if issue.Fields.Assignee == nil {
					assignee = "team"
				} else {
					assignee = issue.Fields.Assignee.Name
				}
				issues[assignee] = append(issues[assignee], issue)

			}
		}(issue)
	}
	wg.Wait()

	if BUGWATCHER_DEBUG == "" {
		var team Team
		if err := team.Load(strings.NewReader(TEAM_MEMBERS_DICT)); err != nil {
			log.Fatalf("error unmarshaling TEAM_MEMBERS_DICT: %v", err)
		}

		slackClient := &http.Client{}
		for assignee, issue := range issues {
			teamMember, ok := team[assignee]
			if !ok {
				teamMember = team["team"]
			}
			if err := notify(SLACK_HOOK, slackClient, issue, teamMember); err != nil {
				gotErrors = true
				log.Print(err)
				return
			}
		}
	} else {
		for assignee, issue := range issues {
			log.Printf("Found %d issues for assignee %s", len(issue), assignee)
			for _, x := range issue {
				log.Printf("- %s (%s)", x.Key, x.Fields.Summary)
			}
		}
	}

	log.Printf("INFO: The query found %d bugs", found)

	if gotErrors {
		os.Exit(1)
	}
}

func init() {
	ex_usage := false
	if JIRA_TOKEN == "" {
		ex_usage = true
		log.Print("Required environment variable not found: JIRA_TOKEN")
	}

	if BUGWATCHER_DEBUG == "" {
		if SLACK_HOOK == "" {
			ex_usage = true
			log.Print("Required environment variable not found: SLACK_HOOK")
		}

		if TEAM_MEMBERS_DICT == "" {
			ex_usage = true
			log.Print("Required environment variable not found: TEAM_MEMBERS_DICT")
		}
	}

	if ex_usage {
		log.Print("Exiting.")
		os.Exit(64)
	}
}
