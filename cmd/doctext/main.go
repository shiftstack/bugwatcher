package main

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"

	jira "github.com/andygrunwald/go-jira"
	"github.com/shiftstack/bugwatcher/pkg/jiraclient"
	"github.com/shiftstack/bugwatcher/pkg/query"
	"github.com/shiftstack/bugwatcher/pkg/slack"
	"github.com/shiftstack/bugwatcher/pkg/team"
)

const queryTriaged = query.ShiftStack + `AND status in ("Release Pending", Verified, ON_QA) AND "Release Note Text" is EMPTY`

var (
	SLACK_HOOK = os.Getenv("SLACK_HOOK")
	JIRA_TOKEN = os.Getenv("JIRA_TOKEN")
	PEOPLE     = os.Getenv("PEOPLE")
)

func main() {
	ctx := context.Background()

	var people []team.Person
	{
		var err error
		people, err = team.Load(strings.NewReader(PEOPLE))
		if err != nil {
			log.Fatalf("error fetching team information: %v", err)
		}
	}

	jiraClient, err := jiraclient.NewWithToken(query.JiraBaseURL, JIRA_TOKEN)
	if err != nil {
		log.Fatalf("error building a Jira client: %v", err)
	}

	triageChecks := [...]triageCheck{
		docTextCheck,
	}

	var (
		found     int
		gotErrors bool
		wg        sync.WaitGroup
	)
	slackClient := slack.New()
	issuesNeedingAttention := make(map[string][]jira.Issue)
	for issue := range query.SearchIssues(ctx, jiraClient, queryTriaged) {
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
					assignee = ""
				} else {
					assignee = issue.Fields.Assignee.Name
				}
				issuesNeedingAttention[assignee] = append(issuesNeedingAttention[assignee], issue)

			}
		}(issue)
	}
	wg.Wait()

	for assigneeJiraName, issues := range issuesNeedingAttention {
		var slackId string
		if person, ok := team.PersonByJiraName(people, assigneeJiraName); ok {
			slackId = person.Slack
		} else {
			slackId = team.TeamSlackId
		}

		if err := slackClient.Send(SLACK_HOOK, notification(issues, slackId)); err != nil {
			gotErrors = true
			log.Print(err)
			return
		}
	}

	log.Printf("INFO: The query found %d bugs", found)

	if gotErrors {
		os.Exit(1)
	}
}

func init() {
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
