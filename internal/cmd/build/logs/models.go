package logs

type BuildLogFlags struct {
	org             string
	project         string
	component       string
	deploymentTrack string
	runId           int
	step            string
}

type CheckCovSummary struct {
	Passed         int    `json:"passed"`
	Failed         int    `json:"failed"`
	Skipped        int    `json:"skipped"`
	ParsingErrors  int    `json:"parsing_errors"`
	ResourceCount  int    `json:"resource_count"`
	CheckovVersion string `json:"checkov_version"`
}

type CheckCovResults struct {
	FailedChecks []CheckCovFailedCheck `json:"failed_checks"`
}

type CheckCovFailedCheck struct {
	CheckID       string `json:"check_id"`
	CheckName     string `json:"check_name"`
	RepoFilePath  string `json:"repo_file_path"`
	FileLineRange []int  `json:"file_line_range"`
}

type CheckCovResponse struct {
	Summary CheckCovSummary `json:"summary"`
	Results CheckCovResults `json:"results"`
}
