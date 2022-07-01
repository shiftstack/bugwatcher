package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	jira "github.com/andygrunwald/go-jira"
)

const (
	jiraBaseURL = "https://issues.redhat.com/"

	baseQuery = `
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
	`

	queryTriaged = baseQuery + `AND labels = "Triaged"`
)

var JIRA_TOKEN = os.Getenv("JIRA_TOKEN")

func main() {
	ctx := context.Background()

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

	var gotErrors bool
	var wg sync.WaitGroup
	missingPriorityComment := "Removing the Triaged label because the Priority assessment is missing."
	for issue := range searchIssues(ctx, jiraClient, queryTriaged) {
		wg.Add(1)
		go func(issue jira.Issue) {
			defer wg.Done()
			if issue.Fields.Priority == nil || issue.Fields.Priority.Name == "Undefined" {
				log.Printf("Untriaging issue %q because of missing Priority", issue.Key)
				if err := untriage(ctx, jiraClient, issue, missingPriorityComment); err != nil {
					gotErrors = true
					log.Print(err)
					return
				}
			}
		}(issue)
	}
	wg.Wait()

	if gotErrors {
		os.Exit(1)
	}
}

func init() {
	if JIRA_TOKEN == "" {
		log.Print("Required environment variable not found: JIRA_TOKEN")
		log.Print("Exiting.")
		os.Exit(64)
	}

	rand.Seed(time.Now().UnixNano())
}
