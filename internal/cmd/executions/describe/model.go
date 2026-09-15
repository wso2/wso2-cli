package describe

type DescribeExecutionParams struct {
	Project         string
	Org             string
	Component       string
	Env             string
	DeploymentTrack string
	ExecutionId     string
	Output          string
}

// ExecutionDescribeOutput is the JSON shape for `wso2-integration-platform describe execution`.
// Kept separate from the on-the-wire logs.ExecutionListItemV2 so the CLI's
// machine-readable contract can stay stable while that evolves. The fields
// mirror what the text path prints, section for section.
type ExecutionDescribeOutput struct {
	Name      string                   `json:"name"`
	Status    string                   `json:"status"`
	Revision  string                   `json:"revision"`
	Duration  string                   `json:"duration"`
	StartTime string                   `json:"startTime"`
	Attempts  []ExecutionAttemptOutput `json:"attempts"`
}

type ExecutionAttemptOutput struct {
	Number   int    `json:"number"`
	ID       string `json:"id"`
	Status   string `json:"status"`
	Duration string `json:"duration"`
	Started  string `json:"started"`
}
