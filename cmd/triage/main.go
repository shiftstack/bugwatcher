package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	jira "github.com/andygrunwald/go-jira"
	"github.com/shiftstack/bugwatcher/cmd/triage/tasker"
	"github.com/shiftstack/bugwatcher/pkg/jiraclient"
	"github.com/shiftstack/bugwatcher/pkg/query"
	"github.com/shiftstack/bugwatcher/pkg/slack"
	"github.com/shiftstack/bugwatcher/pkg/team"
)

const queryUntriaged = query.ShiftStack + `AND (labels not in ("Triaged") OR labels is EMPTY) AND "Need Info From" is EMPTY`

var (
	SLACK_HOOK = os.Getenv("SLACK_HOOK")
	JIRA_EMAIL = os.Getenv("JIRA_EMAIL")
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

	jiraClient, err := jiraclient.NewWithToken(query.JiraBaseURL, JIRA_EMAIL, JIRA_TOKEN)
	if err != nil {
		log.Fatalf("error building a Jira client: %v", err)
	}

	var (
		found     int
		gotErrors bool
		wg        sync.WaitGroup
	)
	slackClient := slack.New()
	issuesByAssignee := new(tasker.Tasker)
	for issue := range query.SearchIssues(ctx, jiraClient, queryUntriaged) {
		wg.Add(1)
		found++
		go func(issue jira.Issue) {
			defer wg.Done()

			var assignee string
			if issue.Fields.Assignee == nil {
				assignee = "team"
			} else {
				assignee = issue.Fields.Assignee.AccountID
			}
			issuesByAssignee.Assign(assignee, issue)

		}(issue)
	}
	wg.Wait()

	for {
		assignee, issues, ok := issuesByAssignee.Pop()
		if !ok {
			break
		}

		var slackId string
		if person, ok := team.PersonByJiraAccountID(people, assignee); ok {
			slackId = person.Slack
		} else {
			log.Printf("failed to find slack ID for team member %s", assignee)
			slackId = team.TeamSlackId
		}

		if err := slackClient.Send(SLACK_HOOK, notification(issues, slackId)); err != nil {
			gotErrors = true
			log.Print(err)
			return
		}
	}

	if gotErrors {
		os.Exit(1)
	}
}

var helpFlag bool

func init() {
	flag.BoolVar(&helpFlag, "help", false, "Show this help message")
	flag.BoolVar(&helpFlag, "h", false, "Show this help message")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `triage - triage ShiftStack Jira bugs

Assigns untriaged ShiftStack Jira bugs to available team members and sends
Slack notifications for each assignment.

Environment variables:
  SLACK_HOOK        Slack webhook URL for notifications
  JIRA_EMAIL        Jira account email address
  JIRA_TOKEN        Jira API token
  PEOPLE            JSON-formatted team member information

Flags:
`)
		flag.PrintDefaults()
	}
	flag.Parse()

	if helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	ex_usage := false
	if SLACK_HOOK == "" {
		ex_usage = true
		log.Print("Required environment variable not found: SLACK_HOOK")
	}

	if JIRA_EMAIL == "" {
		ex_usage = true
		log.Print("Required environment variable not found: JIRA_EMAIL")
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
