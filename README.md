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

* `JIRA_EMAIL`: the email address associated with the Jira Cloud account
* `JIRA_TOKEN`: a [Jira API token](https://id.atlassian.com/manage-profile/security/api-tokens) of an account that can access the OCPBUGS project
* `JIRA_ACCOUNT_ID`: the Jira Cloud account ID of the service account (used for JQL queries)
* `SLACK_HOOK`: a [Slack hook](https://api.slack.com/messaging/webhooks) URL
* `PEOPLE`: an address book. It is a YAML object in the form:

```yaml
- kerberos: user1
  github_handle: ghuser
  jira_name: jirauser
  jira_account_id: "712020:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  slack_id: U012334
  bug_triage: true
  leave:
  - start: 2024-11-21
    end: 2025-02-28
- kerberos: user2
  github_handle: ghuser2
  jira_name: jirauser2
  jira_account_id: "712020:yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy"
  slack_id: U0122345
```

### Local testing

A local script will set the required environment variables for you if you
provide a [vault](https://vault.ci.openshift.org) token.

```shell
export VAULT_TOKEN=$vault_token
make run-pretriage
```

## triage

Usage:

```shell
./triage
```

Reminds assignees about the bugs assigned to them for triage.

Required environment variables:

* `JIRA_EMAIL`: the email address associated with the Jira Cloud account
* `JIRA_TOKEN`: a [Jira API token](https://id.atlassian.com/manage-profile/security/api-tokens) of an account that can access the OCPBUGS project
* `SLACK_HOOK`: a [Slack hook](https://api.slack.com/messaging/webhooks) URL
* `PEOPLE` described [above][pretriage].

## posttriage

Usage:

```shell
./posttriage
```

Resets the `Triaged` keyword on bugs that still need attention.

Required environment variables:

* `JIRA_EMAIL`: the email address associated with the Jira Cloud account
* `JIRA_TOKEN`: a [Jira API token](https://id.atlassian.com/manage-profile/security/api-tokens) of an account that can access the OCPBUGS project

## doctext

Usage:

```shell
./doctext
```

Finds resolved bugs lacking a doc text, and posts a reminder to Slack.

Required environment variables:

* `JIRA_EMAIL`: the email address associated with the Jira Cloud account
* `JIRA_TOKEN`: a [Jira API token](https://id.atlassian.com/manage-profile/security/api-tokens) of an account that can access the OCPBUGS project
* `SLACK_HOOK`: a [Slack hook](https://api.slack.com/messaging/webhooks) URL
* `PEOPLE` described [above][pretriage].
