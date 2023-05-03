package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	jira "github.com/andygrunwald/go-jira"
)

const (
	jiraBaseURL = "https://issues.redhat.com/"

	baseQuery = `
	project = "OpenShift Bugs"
	AND (
		component in (
			"Installer / OpenShift on OpenStack",
			"Storage / OpenStack CSI Drivers",
			"Cloud Compute / OpenStack Provider",
			"Machine Config Operator / platform-openstack",
			"Networking / kuryr",
			"Test Framework / OpenStack"
		)
		OR (
			component in (
				"Installer",
				"Machine Config Operator",
				"Machine Config Operator / platform-none",
				"Cloud Compute / Cloud Controller Manager",
				"Cloud Compute / Cluster Autoscaler",
				"Cloud Compute / MachineHealthCheck",
				"Cloud Compute / Other Provider")
			AND (
				summary ~ "osp"
				OR summary ~ "openstack"
			)
		)
	)
	`

	queryUntriaged = baseQuery + `
	AND (labels not in ("Triaged") OR labels is EMPTY)
	`
)

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

	var jiraClient *jira.Client
	{
		var err error
		jiraClient, err = jira.NewClient(
			(&jira.BearerAuthTransport{Token: JIRA_TOKEN}).Client(),
			jiraBaseURL,
		)
		if err != nil {
			log.Fatalf("error building a Jira client: %v", err)
		}
	}

	var (
		found     int
		gotErrors bool
		wg        sync.WaitGroup
	)
	slackClient := &http.Client{}
	issuesByAssignee := make(map[string][]jira.Issue)
	for issue := range searchIssues(ctx, jiraClient, queryUntriaged) {
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
			issuesByAssignee[assignee] = append(issuesByAssignee[assignee], issue)

		}(issue)
	}
	wg.Wait()

	for assignee, issues := range issuesByAssignee {
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
