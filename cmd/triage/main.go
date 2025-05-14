package main

import (
	"context"
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
	JIRA_TOKEN = os.Getenv("JIRA_TOKEN")
	PEOPLE     = os.Getenv("PEOPLE")
	TEAM       = os.Getenv("TEAM")
)

func main() {
	ctx := context.Background()

	var people []team.Person
	{
		var err error
		people, err = team.Load(strings.NewReader(PEOPLE), strings.NewReader(TEAM))
		if err != nil {
			log.Fatalf("error fetching team information: %v", err)
		}
	}

	jiraClient, err := jiraclient.NewWithToken(query.JiraBaseURL, JIRA_TOKEN)
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
				assignee = issue.Fields.Assignee.Name
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
		if person, ok := team.PersonByJiraName(people, assignee); ok {
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

	if TEAM == "" {
		ex_usage = true
		log.Print("Required environment variable not found: TEAM")
	}

	if ex_usage {
		log.Print("Exiting.")
		os.Exit(64)
	}
}
