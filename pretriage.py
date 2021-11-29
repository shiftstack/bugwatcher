#!/usr/bin/env python
# -*- coding: utf-8 -*-

import os
import json
import sys
import ntplib
import random
import bugzilla
import requests
import secrets
import tenacity


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

BUGZILLA_API_KEY = os.getenv("BUGZILLA_API_KEY")
SLACK_HOOK = os.getenv("SLACK_HOOK")
TEAM_MEMBERS = []
RANDOM = random.choice


def init_random():
    c = ntplib.NTPClient()
    try:
        seed = c.request('pool.ntp.org').tx_time
        random.seed(a=seed)
    except ntplib.NTPException as err:
        print(f'Failed to get time from NTP: {err}')
        print(f'Falling back to secrets.choice()')
        RANDOM = secrets.choice


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
def fetch_bugs():
    print('Fetching bugs...')
    bzapi = bugzilla.Bugzilla(URL, api_key=BUGZILLA_API_KEY)
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
        assignee = RANDOM(TEAM_MEMBERS)
        bzapi.update_bugs([bug.id], bzapi.build_update(
            assigned_to=assignee['bz_id']))
        notify_slack(SLACK_HOOK, assignee['slack_id'], bug.weburl)


def run():
    team = os.getenv("TEAM_MEMBERS")
    if team is None:
        sys.exit(
            ("Error: the JSON object describing the team is required. Set the "
             "TEAM_MEMBERS environment variable.")
        )
    TEAM_MEMBERS = json.loads(team)

    if SLACK_HOOK is None:
        sys.exit(
            ("Error: Slack hook required. Set the SLACK_HOOK environment "
             "variable.")
        )

    init_random()
    fetch_bugs()


if __name__ == '__main__':
    run()
