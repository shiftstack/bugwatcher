package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"strings"
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
				"Installer / OpenShift on OpenStack",
				"Storage / OpenStack CSI Drivers",
				"Cloud Compute / OpenStack Provider",
				"Machine Config Operator / platform-openstack",
				"Networking / kuryr")
			OR (
				component in (
					"Installer",
					"Machine Config Operator",
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
			log.Fatalf("FATAL: error building a Jira client: %v", err)
		}
	}

	triageChecks := [...]triageCheck{
		priorityCheck,
	}

	var (
		found     int
		gotErrors bool
		wg        sync.WaitGroup
	)
	for issue := range searchIssues(ctx, jiraClient, queryTriaged) {
		wg.Add(1)
		found++
		go func(issue jira.Issue) {
			defer wg.Done()
			reasons := make([]string, 0, len(triageChecks))

			for _, check := range triageChecks {
				triaged, msg, err := check(issue)
				if err != nil {
					log.Printf("WARNING: Triage check failed: %v", err)
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

	rand.Seed(time.Now().UnixNano())
}
