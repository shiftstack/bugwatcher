#!/usr/bin/env python3

import argparse
import json
import ntplib
import os
import random
import sys
from datetime import datetime

import bugzilla
import requests
import tenacity


URL = "https://bugzilla.redhat.com"
SHIFTSTACK_QUERY = (
    "https://bugzilla.redhat.com/buglist.cgi?bug_status=__open__&f1=component&"
    "f10=OP&f11=component&f12=component&f13=component&f14=CP&f16=CP&f18=CP&f19"
    "=keywords&f2=OP&f20=assigned_to&f3=rh_sub_components&f4=rh_sub_components"
    "&f5=rh_sub_components&f6=rh_sub_components&f7=rh_sub_components&f8=OP&f9="
    "short_desc&j10=OR&j2=OR&list_id=12471018&o1=notequals&o11=equals&o12=equa"
    "ls&o13=equals&o19=nowords&o20=equals&o3=equals&o4=equals&o5=equals&o6=equ"
    "als&o7=equals&o9=anywords&query_format=advanced&v1=Documentation&v11=Inst"
    "aller&v12=Machine%20Config%20Operator&v13=Cloud%20Compute&v19=Triaged&v20"
    "=shiftstack-bugwatcher%40bot.bugzilla.redhat.com&v3=OpenShift%20on%20Open"
    "Stack&v4=OpenStack%20CSI%20Drivers&v5=OpenStack%20Provider&v6=platform-op"
    "enstack&v7=kuryr&v9=osp%20openstack"
)


@tenacity.retry(
    reraise=True,
    stop=tenacity.stop_after_attempt(10),
    wait=tenacity.wait_fixed(40),
)
def random_seed():
    c = ntplib.NTPClient()
    return c.request('pool.ntp.org').tx_time


def notify_slack(hook, recipient, bug_url):
    msg = {
        'link_names': True,
        'text': (
            f'<@{recipient}> you have been assigned the triage of this '
            f'bug: {bug_url}'
        ),
    }

    x = requests.post(hook, json=msg)

    if x.text != "ok":
        print(f'Error while notifying the assignment of {bug_url}: {x.text}')


@tenacity.retry(
    reraise=True,
    stop=tenacity.stop_after_attempt(10),
    wait=tenacity.wait_fixed(40),
)
def fetch_bugs(bugzilla_api_key, team, slack_hook):
    print('Fetching bugs...')
    bzapi = bugzilla.Bugzilla(URL, api_key=bugzilla_api_key)
    if not bzapi.logged_in:
        sys.exit(
            "Error: You are not logged into Bugzilla. Get an API key here: "
            "https://bugzilla.redhat.com/userprefs.cgi?tab=apikey then set "
            "the BUGZILLA_API_KEY environment variable."
        )

    query = bzapi.url_to_query(SHIFTSTACK_QUERY)
    query["include_fields"] = ["id", "weburl", "component"]

    bugs = bzapi.query(query)

    print(f'Found {len(bugs)} bugs')

    for bug in bugs:
        print(f'Processing bug {bug.id}')
        specialists = [
            m for m in team if bug.component in m.get('components', [])
        ]
        if specialists:
            print(
                f'Found {len(specialists)} specialists for bug {bug.id} '
                f'(component: {bug.component}'
            )
            assignee = random.choice(specialists)
        else:
            print(
                f'Found no specialists for bug {bug.id} '
                f'(component: {bug.component}'
            )
            assignee = random.choice(team)

        bzapi.update_bugs(
            [bug.id], bzapi.build_update(assigned_to=assignee['bz_id'])
        )
        print(f'Assigned bug {bug.id}')
        notify_slack(slack_hook, assignee['slack_id'], bug.weburl)
        print(f'Notified assignee about bug {bug.id}')


def get_team_list():
    team = os.getenv("TEAM_MEMBERS_DICT")
    vacation = os.getenv("TEAM_VACATION")
    if team is None:
        sys.exit(
            "Error: the JSON object describing the team is required. Set the "
            "TEAM_MEMBERS_DICT environment variable."
        )
    team = json.loads(team)

    now = datetime.now()
    for vacation in json.loads(vacation):
        if datetime.fromisoformat(vacation['end']) > now:
            if datetime.fromisoformat(vacation['start']) < now:
                team.pop(vacation['kerberos'], None)

    return [team[member] for member in team]


def run():
    team = get_team_list()

    slack_hook = os.getenv("SLACK_HOOK")
    if slack_hook is None:
        sys.exit(
            "Error: Slack hook required. Set the SLACK_HOOK environment "
            "variable."
        )

    bugzilla_api_key = os.getenv("BUGZILLA_API_KEY")

    random.seed(a=random_seed())

    fetch_bugs(bugzilla_api_key, team, slack_hook)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        description=(
            'Finds untriaged, unassigned ShiftStack bugs and assigns them to '
            'a team member.'
        ),
    )
    args = parser.parse_args()
    run()
