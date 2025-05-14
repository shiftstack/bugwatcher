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
* `PEOPLE`: an address book. It is a YAML object in the form:

```yaml
- kerberos: user1
  github_handle: ghuser
  jira_name: jirauser
  slack_id: U012334
- kerberos: user2
  github_handle: ghuser2
  jira_name: jirauser2
  slack_id: U0122345
```

* `TEAM`: an object containing team members, referencing the `kerberos` property of PEOPLE. It is a YAML object in the form:

```yaml
user1:
  bug_triage: true
  leave:
  - start: 2024-11-21
    end: 2025-02-28
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

* `JIRA_TOKEN`: a [Jira API token](https://issues.redhat.com/secure/ViewProfile.jspa?selectedTab=com.atlassian.pats.pats-plugin:jira-user-personal-access-tokens) of an account that can access the OCPBUGS project
* `SLACK_HOOK`: a [Slack hook](https://api.slack.com/messaging/webhooks) URL
* `PEOPLE` and `TEAM` described [above][pretriage].

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
* `PEOPLE` and `TEAM` described [above][pretriage].
