#!/usr/bin/env python
# -*- coding: utf-8 -*-

import os
import json
import sys
import random
import bugzilla
import requests
from datetime import datetime


URL = "https://bugzilla.redhat.com"
SHIFTSTACK_QUERY = (
    "https://bugzilla.redhat.com/buglist.cgi?bug_status=__open__&f1=component&"
    "f10=component&f11=component&f13=CP&f15=CP&f17=CP&f18=keywords&f19=assigne"
    "d_to&f2=OP&f3=rh_sub_components&f4=rh_sub_components&f5=rh_sub_components"
    "&f6=OP&f7=short_desc&f8=OP&f9=component&j2=OR&j8=OR&list_id=11921616&o1=n"
    "otequals&o10=equals&o11=equals&o18=nowords&o19=equals&o3=equals&o4=equals"
    "&o5=equals&o7=anywords&o9=equals&query_format=advanced&v1=Documentation&v"
    "10=Machine%20Config%20Operator&v11=Cloud%20Compute&v18=Triaged&v19=eduen%"
    "40redhat.com&v3=OpenShift%20on%20OpenStack&v4=OpenStack%20CSI%20drivers&v"
    "5=OpenStack%20Provider&v7=osp%20openstack&v9=Installer"
)

BUGZILLA_API_KEY = os.getenv("BUGZILLA_API_KEY")
SLACK_HOOK = os.getenv("SLACK_HOOK")
TEAM_MEMBERS = json.loads(os.getenv("TEAM_MEMBERS"))


def notify_slack(hook, recipient, bug_url):
    print(recipient)
    msg = {'link_names': True,
           'text': (f'<@{recipient}> you have been assigned'
                    f'the triage of this bug: {bug_url}')}

    x = requests.post(hook, json=msg)

    if x.text != "ok":
        print(f'Error while notifying the assignment of {bug_url}: {x.text}')


if __name__ == '__main__':
    random.seed(datetime.now())

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
        assignee = random.choice(TEAM_MEMBERS)
        bzapi.update_bugs([bug.id], bzapi.build_update(
            assigned_to=assignee['bz_id']))
        notify_slack(SLACK_HOOK, assignee['slack_id'], bug.weburl)
