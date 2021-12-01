#!/usr/bin/env python
# -*- coding: utf-8 -*-

import os
import argparse
import json
import sys
import ntplib
import random
import bugzilla
import requests
import tenacity
from datetime import datetime


URL = "https://bugzilla.redhat.com"
SHIFTSTACK_QUERY = (
    "https://bugzilla.redhat.com/buglist.cgi?bug_status=__open__&f1=component&"
    "f10=component&f11=component&f12=component&f13=CP&f15=CP&f17=CP&f18=keywor"
    "ds&f19=assigned_to&f2=OP&f3=rh_sub_components&f4=rh_sub_components&f5=rh_"
    "sub_components&f6=rh_sub_components&f7=OP&f8=short_desc&f9=OP&j2=OR&j9=OR"
    "&list_id=12150483&o1=notequals&o10=equals&o11=equals&o12=equals&o18=nowor"
    "ds&o19=equals&o3=equals&o4=equals&o5=equals&o6=equals&o8=anywords&query_f"
    "ormat=advanced&v1=Documentation&v10=Installer&v11=Machine%20Config%20Oper"
    "ator&v12=Cloud%20Compute&v18=Triaged&v19=shiftstack-bugwatcher%40bot.bugz"
    "illa.redhat.com&v3=OpenShift%20on%20OpenStack&v4=OpenStack%20CSI%20Driver"
    "s&v5=OpenStack%20Provider&v6=platform-openstack&v8=osp%20openstack"
)


@tenacity.retry(
    reraise=True,
    stop=tenacity.stop_after_attempt(10),
    wait=tenacity.wait_fixed(5)
)
def random_seed():
    c = ntplib.NTPClient()
    return c.request('pool.ntp.org').tx_time


def notify_slack(hook, recipient, bug_url):
    msg = {'link_names': True,
           'text': (f'<@{recipient}> you have been assigned '
                    f'the triage of this bug: {bug_url}')}

    x = requests.post(hook, json=msg)

    if x.text != "ok":
        print(f'Error while notifying the assignment of {bug_url}: {x.text}')


@tenacity.retry(
    reraise=True,
    stop=tenacity.stop_after_attempt(10),
    wait=tenacity.wait_fixed(5)
)
def fetch_bugs(bugzilla_api_key, team, slack_hook):
    print('Fetching bugs...')
    bzapi = bugzilla.Bugzilla(URL, api_key=bugzilla_api_key)
    if not bzapi.logged_in:
        sys.exit(
            ("Error: You are not logged into Bugzilla. Get an API key here: "
             "https://bugzilla.redhat.com/userprefs.cgi?tab=apikey then set "
             "the BUGZILLA_API_KEY environment variable.")
        )

    query = bzapi.url_to_query(SHIFTSTACK_QUERY)
    query["include_fields"] = ["id", "weburl"]

    bugs = bzapi.query(query)

    print(f'Found {len(bugs)} bugs')

    for bug in bugs:
        assignee = random.choice(team)
        bzapi.update_bugs([bug.id], bzapi.build_update(
            assigned_to=assignee['bz_id']))
        notify_slack(slack_hook, assignee['slack_id'], bug.weburl)


def run():
    team = os.getenv("TEAM_MEMBERS")
    if team is None:
        sys.exit(
            ("Error: the JSON object describing the team is required. Set the "
             "TEAM_MEMBERS environment variable.")
        )
    team = json.loads(team)

    slack_hook = os.getenv("SLACK_HOOK")
    if slack_hook is None:
        sys.exit(
            ("Error: Slack hook required. Set the SLACK_HOOK environment "
             "variable.")
        )

    bugzilla_api_key = os.getenv("BUGZILLA_API_KEY")

    random.seed(a=random_seed())

    fetch_bugs(bugzilla_api_key, team, slack_hook)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Finds untriaged, unassigned ShiftStack bugs and assigns them to a team member.')
    args = parser.parse_args()
    run()