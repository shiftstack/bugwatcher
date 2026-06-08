// This program checks that the variables are syntactically correct.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/shiftstack/bugwatcher/pkg/team"
)

var (
	PEOPLE = os.Getenv("PEOPLE")
)

func main() {
	people, err := team.Load(strings.NewReader(PEOPLE))
	if err != nil {
		log.Fatalf("Error loading team members: %v", err)
	}

	fmt.Printf("Found %d people\n", len(people))
}

var helpFlag bool

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)

	flag.BoolVar(&helpFlag, "help", false, "Show this help message")
	flag.BoolVar(&helpFlag, "h", false, "Show this help message")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `check - validate ShiftStack team member configuration

Parses the PEOPLE environment variable and reports the number of team members
found, exiting with an error if the data is syntactically incorrect.

Environment variables:
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
	if PEOPLE == "" {
		ex_usage = true
		log.Print("Required environment variable not found: PEOPLE")
	}

	if ex_usage {
		log.Print("Exiting.")
		os.Exit(64)
	}
}
