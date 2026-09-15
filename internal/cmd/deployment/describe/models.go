package describe

type DescribeDeploymentParams struct {
	componentFlag       string
	projectFlag         string
	orgFlag             string
	envFlag             string
	deploymentTrackFlag string
	outputFlag          string
}

// The JSON output for `wso2-integration-platform describe deployment` describes a single
// environment's deployment, which is exactly the per-environment shape the
// component describe command already emits. Rather than redefine it, the impl
// reuses componentDescribe.EnvironmentDeploymentOutput so the two commands
// stay in lockstep.
