#!/usr/bin/env python
# -*- coding: utf-8 -*-

import os
import sys
import argparse
import bugzilla


BUGZILLA_API_KEY = os.getenv("BUGZILLA_API_KEY")

URL = "https://bugzilla.redhat.com"
SHIFTSTACK_QUERY = (
    "https://bugzilla.redhat.com/buglist.cgi?bug_status=NEW&bug_status=ASSIGNE"
    "D&bug_status=POST&bug_status=MODIFIED&bug_status=ON_DEV&bug_status=ON_QA&"
    "bug_status=VERIFIED&f1=component&f10=component&f11=component&f12=componen"
    "t&f13=CP&f15=CP&f17=CP&f18=keywords&f2=OP&f3=rh_sub_components&f4=rh_sub_"
    "components&f5=rh_sub_components&f6=rh_sub_components&f7=OP&f8=short_desc&"
    "f9=OP&j2=OR&j9=OR&list_id=12283070&o1=notequals&o10=equals&o11=equals&o12"
    "=equals&o18=substring&o3=equals&o4=equals&o5=equals&o6=equals&o8=anywords"
    "&order=Bug%20Number&query_format=advanced&v1=Documentation&v10=Installer&"
    "v11=Machine%20Config%20Operator&v12=Cloud%20Compute&v18=Triaged&v3=OpenSh"
    "ift%20on%20OpenStack&v4=OpenStack%20CSI%20Drivers&v5=OpenStack%20Provider"
    "&v6=platform-openstack&v8=osp%20openstack"
)
QE_TEST_COVERAGE_FLAG = "qe_test_coverage"
TRIAGED_KEYWORD = "Triaged"


def flag_status(flags, flag_name):
    """get_flag returns the status of the first flag found with that name"""
    for flag in flags:
        if flag["name"] == flag_name:
            return flag["status"]


def run():
    print('Fetching bugs...')
    bzapi = bugzilla.Bugzilla(URL, api_key=BUGZILLA_API_KEY)
    if not bzapi.logged_in:
        sys.exit(
            ("Error: You are not logged into Bugzilla. Get an API key here: "
             "https://bugzilla.redhat.com/userprefs.cgi?tab=apikey then set "
             "the BUGZILLA_API_KEY environment variable.")
            )

    query = bzapi.url_to_query(SHIFTSTACK_QUERY)
    query["limit"] = 1000
    query["offset"] = 0
    query["include_fields"] = [
        "id", "keywords", "severity", "priority", "target_release", "flags"]

    bugs = bzapi.query(query)
    print(f'Found {len(bugs)} bugs')
    for bug in bugs:
        missing_severity = bug.severity == "unspecified"
        missing_priority = bug.priority == "unspecified"
        missing_qetest = flag_status(bug.flags, QE_TEST_COVERAGE_FLAG) is None

        reasons = []

        if missing_severity:
            reasons.append("* the severity assessment is missing")
        if missing_priority:
            reasons.append("* the priority assessment is missing")
        if missing_qetest:
            reasons.append(f'* the QE automation assessment (flag {QE_TEST_COVERAGE_FLAG}) is missing')

        if missing_severity or missing_priority or missing_qetest:
            reasons = "\n".join(reasons)
            update = bzapi.build_update(
                comment=f'Removing the {TRIAGED_KEYWORD} keyword because:\n{reasons}',
                keywords_remove=TRIAGED_KEYWORD,
            )
            bzapi.update_bugs([bug.id], update)
            print(f'Updated bug {bug.id}')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Resets the "Triaged" keyword on bugs that still need attention.')
    args = parser.parse_args()
    run()
