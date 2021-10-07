#!/usr/bin/env python
# -*- coding: utf-8 -*-

import os
import json
import sys
import secrets
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
    "ator&v12=Cloud%20Compute&v18=Triaged&v19=shiftstack-bugwatcher%40redhat.c"
    "om&v3=OpenShift%20on%20OpenStack&v4=OpenStack%20CSI%20Drivers&v5=OpenStac"
    "k%20Provider&v6=platform-openstack&v8=osp%20openstack"
)

BUGZILLA_API_KEY = os.getenv("BUGZILLA_API_KEY")
SLACK_HOOK = os.getenv("SLACK_HOOK")
TEAM_MEMBERS = json.loads(os.getenv("TEAM_MEMBERS"))


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
        assignee = secrets.choice(TEAM_MEMBERS)
        bzapi.update_bugs([bug.id], bzapi.build_update(
            assigned_to=assignee['bz_id']))
        notify_slack(SLACK_HOOK, assignee['slack_id'], bug.weburl)


if __name__ == '__main__':

    if TEAM_MEMBERS is None:
        sys.exit(
            ("Error: the JSON object describing the team is required. Set the "
             "TEAM_MEMBERS environment variable.")
        )

    if SLACK_HOOK is None:
        sys.exit(
            ("Error: Slack hook required. Set the SLACK_HOOK environment "
             "variable.")
        )

    fetch_bugs()
