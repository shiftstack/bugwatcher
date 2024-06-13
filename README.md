# Shiftstack Bugwatcher

A collection of scripts to assist us in triaging our bugs.

```shell
make build
```

## pretriage

Usage:

```shell
./pretriage
```

Finds untriaged, unassigned Shiftstack bugs and assigns them to a team member.

Required environment variables:

* `JIRA_TOKEN`: a [Jira API token](https://issues.redhat.com/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens) of an account that can access the OCPBUGS project
* `SLACK_HOOK`: a [Slack hook](https://api.slack.com/messaging/webhooks) URL
* `TEAM_MEMBERS_DICT` is a JSON object in the form:

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

Optional environment variable: `TEAM_VACATION` in the form:

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

## triage

Usage:

```shell
./triage
```

Reminds assignees about the bugs assigned to them for triage.

Required environment variables:

* `JIRA_TOKEN`: a [Jira API token](https://issues.redhat.com/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens) of an account that can access the OCPBUGS project
* `SLACK_HOOK`: a [Slack hook](https://api.slack.com/messaging/webhooks) URL
* `TEAM_MEMBERS_DICT` is a JSON object in the form described [above][pretriage].

## posttriage

Usage:

```shell
./posttriage
```

Resets the `Triaged` keyword on bugs that still need attention.

Required environment variables:

* `JIRA_TOKEN`: a [Jira API token](https://issues.redhat.com/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens) of an account that can access the OCPBUGS project

## doctext

Usage:

```shell
./doctext
```

Finds resolved bugs lacking a doc text, and posts a reminder to Slack.

Required environment variables:

* `JIRA_TOKEN`: a [Jira API token](https://issues.redhat.com/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens) of an account that can access the OCPBUGS project
* `SLACK_HOOK`: a [Slack hook](https://api.slack.com/messaging/webhooks) URL
* `TEAM_MEMBERS_DICT` is a JSON object in the form:
