#!/usr/bin/env python
# -*- coding: utf-8 -*-

import os
import argparse
import json
import sys
import bugzilla
import requests
import tenacity
from datetime import datetime


URL = "https://bugzilla.redhat.com"
SHIFTSTACK_QUERY = (
    "https://bugzilla.redhat.com/buglist.cgi?bug_status=ON_QA&bug_status=VERIF"
    "IED&f1=component&f10=OP&f11=component&f12=component&f13=component&f14=CP&"
    "f16=CP&f18=CP&f19=cf_doc_type&f2=OP&f3=rh_sub_components&f4=rh_sub_compon"
    "ents&f5=rh_sub_components&f6=rh_sub_components&f7=rh_sub_components&f8=OP"
    "&f9=short_desc&j10=OR&j2=OR&list_id=12471057&o1=notequals&o11=equals&o12="
    "equals&o13=equals&o19=equals&o3=equals&o4=equals&o5=equals&o6=equals&o7=e"
    "quals&o9=anywords&query_format=advanced&v1=Documentation&v11=Installer&v1"
    "2=Machine%20Config%20Operator&v13=Cloud%20Compute&v19=If%20docs%20needed%"
    "2C%20set%20a%20value&v3=OpenShift%20on%20OpenStack&v4=OpenStack%20CSI%20D"
    "rivers&v5=OpenStack%20Provider&v6=platform-openstack&v7=kuryr&v9=osp%20op"
    "enstack"
)


def bz_to_slack(team, bz_id):
    for member in team:
        if member['bz_id'] == bz_id:
            return member['slack_id']
    return 'openstack-dev-team'


@tenacity.retry(
    reraise=True,
    stop=tenacity.stop_after_attempt(2),
    wait=tenacity.wait_fixed(60)
)
def notify_slack(hook, recipient, bugs):
    msg = {'link_names': True,
           'text': (f'<@{recipient}> please check the doctext for these bugs: '
                    + " ".join({f'<{bug.weburl}|{bug.id}>' for bug in bugs}))}

    x = requests.post(hook, json=msg)

    if x.text != "ok":
        print(f'Error while notifying the assignment of {bug_url}: {x.text}')


@tenacity.retry(
    reraise=True,
    stop=tenacity.stop_after_attempt(10),
    wait=tenacity.wait_fixed(40)
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
    query["include_fields"] = ["id", "assigned_to", "weburl"]

    bugs = bzapi.query(query)

    print(f'Found {len(bugs)} bugs')

    bugs_by_assignee = {}
    for bug in bugs:
        slack_id = bz_to_slack(team, bug.assigned_to)
        bugs_by_assignee[slack_id] = bugs_by_assignee.get(slack_id, []) + [bug]

    return bugs_by_assignee


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

    bugs_by_assignee = fetch_bugs(bugzilla_api_key, team, slack_hook)

    for assignee in bugs_by_assignee:
        notify_slack(slack_hook, assignee, bugs_by_assignee[assignee])


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Finds ShiftStack bugs without doctext and notifies about them on Slack.')
    args = parser.parse_args()
    run()
