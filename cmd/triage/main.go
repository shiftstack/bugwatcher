package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	jira "github.com/andygrunwald/go-jira"
	"github.com/shiftstack/bugwatcher/cmd/triage/tasker"
	"github.com/shiftstack/bugwatcher/pkg/jiraclient"
	"github.com/shiftstack/bugwatcher/pkg/query"
)

const queryUntriaged = query.ShiftStack + `AND (labels not in ("Triaged") OR labels is EMPTY) AND "Need Info From" is EMPTY`

var (
	SLACK_HOOK        = os.Getenv("SLACK_HOOK")
	JIRA_TOKEN        = os.Getenv("JIRA_TOKEN")
	TEAM_MEMBERS_DICT = os.Getenv("TEAM_MEMBERS_DICT")
)

func main() {
	ctx := context.Background()

	var team Team
	if err := team.Load(strings.NewReader(TEAM_MEMBERS_DICT)); err != nil {
		log.Fatalf("error unmarshaling TEAM_MEMBERS_DICT: %v", err)
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
	slackClient := &http.Client{}
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

		teamMember, ok := team[assignee]
		if !ok {
			log.Printf("failed to find slack ID for team member %s", assignee)
			teamMember = team["team"]
		}

		if err := notify(SLACK_HOOK, slackClient, issues, teamMember); err != nil {
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

	if TEAM_MEMBERS_DICT == "" {
		ex_usage = true
		log.Print("Required environment variable not found: TEAM_MEMBERS_DICT")
	}

	if ex_usage {
		log.Print("Exiting.")
		os.Exit(64)
	}
}
