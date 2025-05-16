// This program checks that the variables are syntactically correct.
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/shiftstack/bugwatcher/pkg/team"
)

var (
	PEOPLE = os.Getenv("PEOPLE")
	TEAM   = os.Getenv("TEAM")
)

func main() {
	people, err := team.Load(strings.NewReader(PEOPLE), strings.NewReader(TEAM))
	if err != nil {
		log.Fatalf("Error loading team members: %v", err)
	}

	fmt.Printf("Found %d people\n", len(people))

	{
		var count int
		for i := range people {
			if people[i].TeamMember {
				count++
			}
		}
		fmt.Printf("Found %d team members\n", count)
	}
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)

	ex_usage := false
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
