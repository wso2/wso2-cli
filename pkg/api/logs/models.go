package logs

import "time"

type GetProjectLogsReqBody struct {
	ComponentIdList []string `json:"componentIdList"`
	EndTime         string   `json:"endTime"`
	EnvironmentID   string   `json:"environmentId"`
	Limit           uint     `json:"limit"`
	LogLevels       []string `json:"logLevels"`
	ProjectID       string   `json:"projectId"`
	SearchPhrase    string   `json:"searchPhrase"`
	Sort            string   `json:"sort"`
	SortingOrder    string   `json:"sortingOrder"`
	StartTime       string   `json:"startTime"`
}

type GetComponentLogsReqBody struct {
	ComponentID   string   `json:"componentId"`
	EnvironmentID string   `json:"environmentId"`
	VersionList   []string `json:"versionList"`
	VersionIDList []string `json:"versionIdList"`
	LogType       string   `json:"logType"`
	Region        string   `json:"region"`
	SearchPhrase  string   `json:"searchPhrase"`
	StartTime     string   `json:"startTime"`
	EndTime       string   `json:"endTime"`
	Limit         uint     `json:"limit"`
	Sort          string   `json:"sort"`
	SortingOrder  string   `json:"sortingOrder"`
}

type GetBuildLogsReqBody struct {
	ComponentId       string `json:"componentId"`
	DeploymentTrackId string `json:"deploymentTrackId"`
	WorkflowName      string `json:"workflowName"`
}

type LogEntryApiResponse struct {
	Columns []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"columns"`
	Rows [][]interface{} `json:"rows"` // rows could be 2d array of any type
}

type ProjectLogItem struct {
	LogLevel         string `json:"logLevel"`
	Environment      string `json:"environment"`
	ComponentName    string `json:"componentName"`
	ComponentVersion string `json:"componentVersion"`
	LogEntry         string `json:"logEntry"`
	TimeGenerated    string `json:"timeGenerated"`
}

type ComponentLogItem struct {
	LogLevel         string `json:"logLevel"`
	Environment      string `json:"environment"`
	LogContext       string `json:"logContext"`
	ComponentVersion string `json:"componentVersion"`
	LogEntry         string `json:"logEntry"`
	TimeGenerated    string `json:"timeGenerated"`
}

type ExecutionListResponse struct {
	List LogEntryApiResponse `json:"list"`
}

type ExecutionListItemV2 struct {
	ID             string `json:"id"`
	StartTime      string `json:"startTime"`
	CompletionTime string `json:"completionTime"`
	RunID          string `json:"runId"`
	RevisionID     string `json:"revisionId"`
	FailedReason   string `json:"failedReason"`
	Status         string `json:"status"`
}

type ExecutionListItem struct {
	ControllerName string
	ID             string
	Name           string
	RunId          string
	Revision       string
	StartTime      time.Time
	EndTime        time.Time
	Runs           []ExecutionAttempt
	Reason         string
	IsActive       bool
}

type ExecutionAttempt struct {
	ID             string `json:"id"`
	StartTime      string `json:"startTime"`
	CompletionTime string `json:"completionTime"`
	Status         string `json:"status"`
}
