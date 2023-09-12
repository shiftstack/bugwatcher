package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/shiftstack/bugwatcher/pkg/query"
)

const queryUntriaged = query.ShiftStack + `AND assignee is EMPTY AND (labels not in ("Triaged") OR labels is EMPTY)`

var (
	SLACK_HOOK        = os.Getenv("SLACK_HOOK")
	JIRA_TOKEN        = os.Getenv("JIRA_TOKEN")
	TEAM_MEMBERS_DICT = os.Getenv("TEAM_MEMBERS_DICT")
	TEAM_VACATION     = os.Getenv("TEAM_VACATION")
)

func main() {
	var team Team
	if err := team.Load(strings.NewReader(TEAM_MEMBERS_DICT), strings.NewReader(TEAM_VACATION)); err != nil {
		log.Fatalf("error unmarshaling TEAM_MEMBERS_DICT: %v", err)
	}

	var jiraClient *jira.Client
	{
		var err error
		jiraClient, err = jira.NewClient(
			(&jira.BearerAuthTransport{Token: JIRA_TOKEN}).Client(),
			query.JiraBaseURL,
		)
		if err != nil {
			log.Fatalf("error building a Jira client: %v", err)
		}
	}

	slackClient := &http.Client{}

	now := time.Now()
	var gotErrors bool
	var wg sync.WaitGroup
	for issue := range query.SearchIssues(context.Background(), jiraClient, queryUntriaged) {
		wg.Add(1)
		go func(issue jira.Issue) {
			defer wg.Done()
			var assignee TeamMember
			if parent, isBackport, err := backportParent(jiraClient, issue); isBackport {
				if err != nil {
					log.Print(err)
					gotErrors = true
					return
				}
				log.Printf("Issue %q has parent %q, which is assigned to %q", issue.Key, parent.Key, censorEmail(parent.Fields.Assignee.Name))
				if teamMember, ok := team.MemberByJiraName(parent.Fields.Assignee.Name); ok {
					assignee = teamMember
				}
			}

			if assignee.JiraName == "" {
				// "It should be 1 component per bug" 🤞
				// The current JQL query filters in by component anyway.
				//
				// https://coreos.slack.com/archives/C02F4Q7EF5L/p1656519746123569
				issueComponent := issue.Fields.Components[0].Name
				var err error
				assignee, err = RandomAvailable(team.Specialists(issueComponent), now)
				if err != nil {
					gotErrors = true
					log.Printf("Error finding an assignee for issue %q: %v", issue.Key, err)
					return
				}
			}

			log.Printf("Assigning issue %q to %q", issue.Key, censorEmail(assignee.JiraName))

			if err := assign(jiraClient, issue, assignee); err != nil {
				gotErrors = true
				log.Print(err)
				return
			}

			if err := notify(SLACK_HOOK, slackClient, issue, assignee); err != nil {
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

	if TEAM_MEMBERS_DICT == "" {
		ex_usage = true
		log.Print("Required environment variable not found: TEAM_MEMBERS_DICT")
	}

	if TEAM_VACATION == "" {
		TEAM_VACATION = "[]"
	}

	if ex_usage {
		log.Print("Exiting.")
		os.Exit(64)
	}
}
