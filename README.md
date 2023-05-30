# Shiftstack Bugwatcher

## pretriage

**Usage:**

```shell
go build ./cmd/pretriage && ./pretriage`
```

Finds untriaged, unassigned Shiftstack bugs and assigns them to a team member.

**Required environment variables:**

* `JIRA_TOKEN`: a [Jira API token](https://issues.redhat.com/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens) of an account that can access the OCPBUGS project
* `SLACK_HOOK`: a [Slack hook](https://api.slack.com/messaging/webhooks) URL
* `TEAM_MEMBERS_DICT`: a JSON object in the form:

```json
{
  "kerberos_id1": {
    "slack_id": "UG65473AM",
    "bz_id": "user1@example.com",
    "components": ["component1"],
    "jira_name": "user1",
    "jira_components": ["component1/sub-component1"]
  },
  "kerberos_id2": {
    "slack_id": "UGF8B93HA",
    "bz_id": "user2@example.com",
    "components": [],
    "jira_name": "user2",
    "jira_components": []
  }
}
```

**Optional environment variables:**

* `TEAM_VACATION`: a JSON object in the form:

```json
[
  {
    "kerberos": "jdoe",
    "start": "2022-01-01",
    "end": "2022-01-15"
  },
  {
    "kerberos": "jdoe",
    "start": "2022-06-12",
    "end": "2022-06-15"
  }
]
```

## posttriage

**Usage:**

```shell
go build ./cmd/posttriage && ./posttriage
```

Resets the `Triaged` keyword on bugs that still need attention.

**Required environment variables:**

* `JIRA_TOKEN`: a [Jira API token](https://issues.redhat.com/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens) of an account that can access the OCPBUGS project

## doctext

**Usage:**

```shell
go build ./cmd/doctext && ./doctext
```

Finds resolved bugs lacking a doc text, and posts a reminder to Slack.

**Required environment variables:**

* `JIRA_TOKEN`: a [Jira API token](https://issues.redhat.com/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens) of an account that can access the OCPBUGS project
* `SLACK_HOOK`: a [Slack hook](https://api.slack.com/messaging/webhooks) URL (optional and ignored if `BUGWATCHER_DEBUG` set)
* `TEAM_MEMBERS_DICT`: a JSON object in the form described previously (optional and ignored if `BUGWATCHER_DEBUG` set)

**Optional environment variables:**

* `BUGWATCHER_DEBUG`: enable debug mode, where found bugs are logged to output instead of Slack
