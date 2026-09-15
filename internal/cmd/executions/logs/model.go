package logs

type ExecutionLogsOpts struct {
	Org             string
	Project         string
	Component       string
	DeploymentTrack string
	Env             string
	ExecutionId     string
	AttemptNumber   int
}
