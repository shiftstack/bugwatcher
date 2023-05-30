package tasker

import (
	"sync"

	jira "github.com/andygrunwald/go-jira"
)

// Tasker contains the one-to-many relationship between assignees and their
// assigned issues.
type Tasker struct {
	sync.Mutex
	issues map[string][]jira.Issue
}

// Assign adds one issue to a specific assignee.
func (t *Tasker) Assign(assignee string, issue jira.Issue) {
	t.Lock()
	defer t.Unlock()

	if t.issues == nil {
		t.issues = make(map[string][]jira.Issue)
	}

	t.issues[assignee] = append(t.issues[assignee], issue)
}

// Pop returns one assignee and all their assigned issues, and removes them
// from the Tasker.
// The boolean value is false if the tasker is empty.
func (t *Tasker) Pop() (string, []jira.Issue, bool) {
	t.Lock()
	defer t.Unlock()

	for assignee, issues := range t.issues {
		delete(t.issues, assignee)
		return assignee, issues, true
	}
	return "", nil, false
}
