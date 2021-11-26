# Shiftstack Bugwatcher

## pretriage

Usage: `./pretriage.py`

Finds untriaged, unassigned Shiftstack bugs and assigns them to a team member.

Required environment variables:

* `BUGZILLA_API_KEY`: a [Bugzilla API key](https://bugzilla.redhat.com/userprefs.cgi?tab=apikey). 
* `SLACK_HOOK`: a Slack hook URL.
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

1. Run `make pretriage_query_url`
1. Paste the resulting URL in your browser address bar
1. Click on the button "Edit Search" at the bottom of the bug list

## posttriage

Usage: `./posttriage.py`

Resets the `Triaged` keyword on bugs that still need attention.

Required environment variables:

* `BUGZILLA_API_KEY`: a [Bugzilla API key](https://bugzilla.redhat.com/userprefs.cgi?tab=apikey). 

### Development

To validate a Bugzilla query:

1. Run `make posttriage_query_url`
1. Paste the resulting URL in your browser address bar
1. Click on the button "Edit Search" at the bottom of the bug list
