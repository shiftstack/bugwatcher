# Shiftstack Bugwatcher

## pretriage

Usage:
* for Bugzilla: `./pretriage.py`
* for Jira: `go build ./cmd/pretriage && ./pretriage`

Finds untriaged, unassigned Shiftstack bugs and assigns them to a team member.

Required environment variables:

* `BUGZILLA_API_KEY`: a [Bugzilla API key](https://bugzilla.redhat.com/userprefs.cgi?tab=apikey)
* `JIRA_TOKEN`: a [Jira API token](https://issues.redhat.com/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens) of an account that can access the OCPBUGS project
* `SLACK_HOOK`: a [Slack hook](https://api.slack.com/messaging/webhooks) URL
* `TEAM_MEMBERS` is a JSON object in the form:

```json
[
  {
    "slack_id": "UG65473AM",
    "bz_id": "user1@example.com",
    "components": ["component1"],
    "jira_name": "user1",
    "jira_components": ["component1/sub-component1"]
  },
  {
    "slack_id": "UGF8B93HA",
    "bz_id": "user2@example.com",
    "components": [],
    "jira_name": "user2",
    "jira_components": []
  }
]
```

### Development

To validate the Bugzilla query:

1. Run `make query_url_pretriage`
1. Paste the resulting URL in your browser address bar
1. Click on the button "Edit Search" at the bottom of the bug list

## posttriage

Usage:
* for Bugzilla: `./posttriage.py`
* for Jira: `go build ./cmd/posttriage && ./posttriage`

Resets the `Triaged` keyword on bugs that still need attention.

Required environment variables:

* `BUGZILLA_API_KEY`: a [Bugzilla API key](https://bugzilla.redhat.com/userprefs.cgi?tab=apikey). 
* `JIRA_TOKEN`: a [Jira API token](https://issues.redhat.com/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens) of an account that can access the OCPBUGS project

### Development

To validate the Bugzilla query:

1. Run `make query_url_posttriage`
1. Paste the resulting URL in your browser address bar
1. Click on the button "Edit Search" at the bottom of the bug list

## doctext

Usage: `./doctext.py`

Finds OCP 4.10 resolved bugs lacking a doc text, and posts a reminder to Slack.

Required environment variables:

* `BUGZILLA_API_KEY`: a [Bugzilla API key](https://bugzilla.redhat.com/userprefs.cgi?tab=apikey)
* `SLACK_HOOK`: a [Slack hook](https://api.slack.com/messaging/webhooks) URL
* `TEAM_MEMBERS` is a JSON object in the form:

```json
[
  {
    "slack_id": "UG65473AM",
    "bz_id": "user1@example.com"
  },
  {
    "slack_id": "UGF8B93HA",
    "bz_id": "user2@example.com"
  }
]
```

### Development

To validate a Bugzilla query:

1. Run `make query_url_doctext`
1. Paste the resulting URL in your browser address bar
1. Click on the button "Edit Search" at the bottom of the bug list
