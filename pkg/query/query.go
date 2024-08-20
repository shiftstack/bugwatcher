package query

const JiraBaseURL = "https://issues.redhat.com/"

const ShiftStack = `project = "OpenShift Bugs"
	AND (
		component in (
			"Installer / OpenShift on OpenStack",
			"Storage / OpenStack CSI Drivers",
			"Cloud Compute / OpenStack Provider",
			"Machine Config Operator / platform-openstack",
			"Networking / kuryr",
			"Test Framework / OpenStack"
		)
		OR (
			component in (
				"Installer",
				"Machine Config Operator",
				"Machine Config Operator / platform-none",
				"Cloud Compute / Cloud Controller Manager",
				"Cloud Compute / Cluster Autoscaler",
				"Cloud Compute / MachineHealthCheck",
				"Cloud Compute / Unknown")
			AND (
				summary ~ "osp"
				OR summary ~ "openstack"
			)
		)
	)
`
