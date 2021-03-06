package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	jira "github.com/andygrunwald/go-jira"
)

const jiraBaseURL = "https://issues.redhat.com/"

const query = `
	project = "OpenShift Bugs"
	AND (
		component in (
			"Installer/OpenShift on OpenStack",
			"Storage/OpenStack CSI Drivers",
			"Cloud Compute/OpenStack Provider",
			"Machine Config Operator/platform-openstack",
			"Networking/kuryr")
		OR (
			component in (
				"Installer",
				"Machine Config Operator",
				"Machine Config Operator/platform-none",
				"Cloud Compute/Cloud Controller Manager",
				"Cloud Compute/Cluster Autoscaler",
				"Cloud Compute/MachineHealthCheck",
				"Cloud Compute/Other Provider")
			AND (
				summary ~ "osp"
				OR summary ~ "openstack"
			)
		)
	)
	AND assignee is EMPTY
	AND (labels not in ("Triaged") OR labels is EMPTY)
	`

var (
	SLACK_HOOK   = os.Getenv("SLACK_HOOK")
	JIRA_TOKEN   = os.Getenv("JIRA_TOKEN")
	TEAM_MEMBERS = os.Getenv("TEAM_MEMBERS")
)

func main() {
	var team Team
	{
		err := json.Unmarshal([]byte(TEAM_MEMBERS), &team)
		if err != nil {
			log.Fatalf("error unmarshaling TEAM_MEMBERS: %v", err)
		}
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

	slackClient := &http.Client{}

	var gotErrors bool
	var wg sync.WaitGroup
	for issue := range searchIssues(context.Background(), jiraClient, query) {
		wg.Add(1)
		go func(issue jira.Issue) {
			defer wg.Done()

			// "It should be 1 component per bug" 🤞
			// The current JQL query filters in by component anyway.
			//
			// https://coreos.slack.com/archives/C02F4Q7EF5L/p1656519746123569
			issueComponent := issue.Fields.Components[0].Name
			assignee := team.NewAssignee(issueComponent)
			log.Printf("Issue %q has component %q and will be assigned to %q", issue.Key, issueComponent, assignee.JiraName)

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
	ex_usage := false
	if SLACK_HOOK == "" {
		ex_usage = true
		log.Print("Required environment variable not found: SLACK_HOOK")
	}

	if JIRA_TOKEN == "" {
		ex_usage = true
		log.Print("Required environment variable not found: JIRA_TOKEN")
	}

	if TEAM_MEMBERS == "" {
		ex_usage = true
		log.Print("Required environment variable not found: TEAM_MEMBERS")
	}

	if ex_usage {
		log.Print("Exiting.")
		os.Exit(64)
	}

	rand.Seed(time.Now().UnixNano())
}
