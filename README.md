# Shiftstack Bugwatcher

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
