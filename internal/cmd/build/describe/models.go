package describe

type BuildDescribeParams struct {
	ComponentFlag       string
	ProjectFlag         string
	OrgFlag             string
	DeploymentTrackFlag string
	RunIdFlag           int
	OutputFlag          string
}

// BuildDescribeOutput is the JSON shape for `wso2-integration-platform describe build`.
// Kept separate from deploymentbuild.BuildKind so the on-the-wire API type can
// evolve without breaking the CLI's machine-readable contract.
type BuildDescribeOutput struct {
	ID          int               `json:"id"`
	Status      string            `json:"status"`
	Conclusion  string            `json:"conclusion"`
	StartedAt   string            `json:"startedAt"`
	CompletedAt string            `json:"completedAt,omitempty"`
	Commit      BuildCommitOutput `json:"commit"`
	Steps       *BuildStepsOutput `json:"steps,omitempty"`
}

type BuildCommitOutput struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
}

// BuildStepsOutput is only populated for failed builds, where the CLI already
// fetches step-level data to render the per-phase tables in text mode.
type BuildStepsOutput struct {
	Initialization []BuildStepOutput `json:"initialization"`
	Build          []BuildStepOutput `json:"build"`
	Finalization   []BuildStepOutput `json:"finalization"`
}

type BuildStepOutput struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Conclusion  string `json:"conclusion"`
	StartedAt   string `json:"startedAt,omitempty"`
	CompletedAt string `json:"completedAt,omitempty"`
}
