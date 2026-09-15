package gqlbuild

type BuildListResponse struct {
	Id            string `json:"id"`
	VersionId     string `json:"versionId"`
	BuildId       string `json:"buildId"`
	CommitHash    string `json:"commitHash"`
	ComponentId   string `json:"componentId"`
	CreatedDate   string `json:"createdDate"`
	CommitMessage string `json:"commitMessage"`
}

type DeploymentStatusByVersionResponse struct {
	Id             int    `json:"id"`
	Sha            string `json:"sha"`
	Completed_at   string `json:"completed_at"`
	Started_at     string `json:"started_at"`
	Name           string `json:"name"`
	Status         string `json:"status"`
	Conclusion     string `json:"conclusion"`
	IsAutoDeploy   bool   `json:"isAutoDeploy"`
	FailureReason  int    `json:"failureReason"`
	SourceCommitId string `json:"sourceCommitId"`
}

type ProxyBuildCreateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ProjectBuildLogsData struct {
	IntegrationProjectBuild string `json:"integrationProjectBuild"`
	LibraryTrivyReport      string `json:"libraryTrivyReport"`
	DropinsTrivyReport      string `json:"dropinsTrivyReport"`
	PostBuildCheckLogs      string `json:"postBuildCheckLogs"`
	MainSequenceValidation  string `json:"mainSequenceValidation"`
	MIVersionValidation     string `json:"mIVersionValidation"`
	ProxyBuildLogs          string `json:"proxyBuildLogs,omitempty"`
	GovernanceLogs          string `json:"governanceLogs,omitempty"`
	ConfigValidationLogs    string `json:"configValidationLogs,omitempty"`
}

type DeployScanResult struct {
	TrivyScan   string `json:"trivyScan"`
	CheckovScan string `json:"checkovScan"`
}
