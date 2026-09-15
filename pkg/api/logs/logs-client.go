package logs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
)

type LogsClient struct {
	sysApiPrefix string
	client       *api.IPHTTPClient
}

func NewLogsClient(sysApiPrefix string, tokenStore api.ReadOnlyTokenStore) *LogsClient {
	return &LogsClient{
		sysApiPrefix: sysApiPrefix,
		client:       api.NewIPHTTPClient(tokenStore),
	}
}

func (c *LogsClient) GetProjectLogs(
	reqBody GetProjectLogsReqBody,
	orgId string,
	externalVHost string,
	isCilium bool,
	envName string,
	isLive bool) ([]ProjectLogItem, error) {

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling request body: %w", err)
	}

	logBaseUrl := c.sysApiPrefix + "." + externalVHost

	if isCilium {
		logBaseUrl = logBaseUrl + "/cilium"
	}

	logsReqUrl := logBaseUrl + "/systemapis/choreologgingapi/0.2.0/logs/project/application"
	if isLive {
		logsReqUrl = logsReqUrl + "?live=true"
	}
	req, err := http.NewRequest("POST", logsReqUrl, bytes.NewBuffer([]byte(reqBodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching project logs: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	var response *LogEntryApiResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	logItems, _ := GenerateProjectLogItems(response, envName)

	return logItems, nil
}

func (c *LogsClient) GetBuildLogs(
	reqBody GetBuildLogsReqBody,
	orgId string,
	externalVHost string) (*component.DeploymentBuildStatusResponse, error) {

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling request body: %w", err)
	}

	logBaseUrl := c.sysApiPrefix + "." + externalVHost
	logsReqUrl := logBaseUrl + "/systemapis/choreologgingapi/0.2.0/logs/component/build"

	req, err := http.NewRequest("POST", logsReqUrl, bytes.NewBuffer([]byte(reqBodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching build logs: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	var response *component.DeploymentBuildStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}

	return response, nil
}

func (c *LogsClient) GetComponentLogs(
	reqBody GetComponentLogsReqBody,
	orgId string,
	externalVHost string,
	isCilium bool,
	envName string,
	logType string,
	isLive bool) ([]ComponentLogItem, error) {

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling request body: %w", err)
	}

	logBaseUrl := c.sysApiPrefix + "." + externalVHost

	if isCilium {
		logBaseUrl = logBaseUrl + "/cilium"
	}

	logsReqUrl := logBaseUrl + "/systemapis/choreologgingapi/0.2.0/logs/component/" + logType

	if isLive {
		logsReqUrl = logsReqUrl + "?live=true"
	}
	req, err := http.NewRequest("POST", logsReqUrl, bytes.NewBuffer([]byte(reqBodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching component logs: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	var response *LogEntryApiResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	logItems, _ := GenerateComponentLogItems(response, envName)

	return logItems, nil
}

func (c *LogsClient) GetExecutionLogs(
	orgId string,
	externalVHost string,
	isCilium bool,
	componentId string,
	dTrackId string,
	attemptId string,
	envId string,
	envName string,
) ([]ComponentLogItem, error) {

	logBaseUrl := c.sysApiPrefix + "." + externalVHost

	if isCilium {
		logBaseUrl = logBaseUrl + "/cilium"
	}

	logsReqUrl := fmt.Sprintf(
		"%s/%s/components/%s/deployment-tracks/%s/executions/%s/logs?environmentId=%s",
		logBaseUrl,
		"systemapis/choreologgingapi/0.2.0",
		componentId,
		dTrackId,
		attemptId,
		envId,
	)

	req, err := http.NewRequest("GET", logsReqUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching component logs: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	var response *LogEntryApiResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	logItems, _ := GenerateComponentLogItems(response, envName)

	return logItems, nil
}

func (c *LogsClient) GetExecutionById(
	orgId, baseUrl, executionId, releaseId string,
	isCilium bool,
) (exec *ExecutionListItemV2, err error) {

	reqUrl := ""
	if isCilium {
		reqUrl = fmt.Sprintf(
			"%s.%s/clilium/choreoobsapi/0.3.0/tasks/executions/%s?releaseId=%s",
			c.sysApiPrefix,
			baseUrl,
			executionId,
			releaseId,
		)
	} else {
		reqUrl = fmt.Sprintf(
			"%s.%s/systemapis/choreoobsapi/0.3.0/tasks/executions/%s?releaseId=%s",
			c.sysApiPrefix,
			baseUrl,
			executionId,
			releaseId,
		)
	}

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&exec); err != nil {
		return nil, err
	}

	return
}

func (c *LogsClient) GetExecutionAttempts(
	orgId, baseUrl, executionId, releaseId string,
	isCilium bool,
) (list []ExecutionAttempt, err error) {
	// choreoobsapi/0.3.0/tasks/executions/%s/attempts?releaseId=%s
	reqUrl := ""
	if isCilium {
		reqUrl = fmt.Sprintf(
			"%s.%s/clilium/choreoobsapi/0.3.0/tasks/executions/%s/attempts?releaseId=%s",
			c.sysApiPrefix,
			baseUrl,
			executionId,
			releaseId,
		)
	} else {
		reqUrl = fmt.Sprintf(
			"%s.%s/systemapis/choreoobsapi/0.3.0/tasks/executions/%s/attempts?releaseId=%s",
			c.sysApiPrefix,
			baseUrl,
			executionId,
			releaseId,
		)
	}

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}

	return
}

func (c *LogsClient) GetExecutionsListV2(
	orgId, baseUrl, releaseId string,
	isCilium bool,
	limit int,
) (list []ExecutionListItemV2, err error) {
	// choreoobsapi/0.3.0/tasks/executions?releaseId=%s&limit=%d&verbose=true
	reqUrl := ""
	if isCilium {
		reqUrl = fmt.Sprintf(
			"%s.%s/clilium/systemapis/choreoobsapi/0.3.0/tasks/executions?releaseId=%s&limit=%d&verbose=true",
			c.sysApiPrefix,
			baseUrl,
			releaseId,
			limit,
		)
	} else {
		reqUrl = fmt.Sprintf(
			"%s.%s/systemapis/choreoobsapi/0.3.0/tasks/executions?releaseId=%s&limit=%d&verbose=true",
			c.sysApiPrefix,
			baseUrl,
			releaseId,
			limit,
		)
	}

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}

	return
}

func (c *LogsClient) GetExecutionsList(
	orgId, baseUrl, componentId, deploymentTrackId, envId string,
	isCilium bool,
	offSet, limit int,
) (list []ExecutionListItem, err error) {

	reqUrl := ""

	if isCilium {
		reqUrl = fmt.Sprintf(
			"%s.%s/clilium/systemapis/choreologgingapi/0.2.0/components/%s/deployment-tracks/%s/executions?environmentId=%s&offset=%d&limit=%d",
			c.sysApiPrefix,
			baseUrl,
			componentId,
			deploymentTrackId,
			envId,
			offSet,
			limit,
		)
	} else {
		reqUrl = fmt.Sprintf(
			"%s.%s/systemapis/choreologgingapi/0.2.0/components/%s/deployment-tracks/%s/executions?environmentId=%s&offset=%d&limit=%d",
			c.sysApiPrefix,
			baseUrl,
			componentId,
			deploymentTrackId,
			envId,
			offSet,
			limit,
		)
	}

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	var result *ExecutionListResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	list, err = GenerateExecutionListResponse(&result.List)

	return
}

func GenerateExecutionListResponse(response *LogEntryApiResponse) (list []ExecutionListItem, err error) {
	// Create a map for column index and name
	columnIndex := make(map[string]int)
	for i, column := range response.Columns {
		columnIndex[column.Name] = i
	}

	// Iterate over rows
	for _, row := range response.Rows {
		if len(row) != len(response.Columns) {
			return nil, fmt.Errorf("row length doesn't match column length")
		}

		executionItem := &ExecutionListItem{}
		// executionItem.Runs = make([]ExecutionAttempt, 0)

		executionItem.ControllerName = row[columnIndex["ControllerName"]].(string)
		executionItem.ID = row[columnIndex["ID"]].(string)
		executionItem.Name = row[columnIndex["Name"]].(string)
		executionItem.RunId = row[columnIndex["RunId"]].(string)
		executionItem.Revision = row[columnIndex["Revision"]].(string)
		executionItem.Reason = row[columnIndex["Reason"]].(string)
		executionItem.IsActive = row[columnIndex["isActive"]].(bool)

		if t, err := time.Parse(time.RFC3339, row[columnIndex["StartTime"]].(string)); err != nil {
			return nil, err
		} else {
			executionItem.StartTime = t
		}

		if t, err := time.Parse(time.RFC3339, row[columnIndex["EndTime"]].(string)); err != nil {
			return nil, err
		} else {
			executionItem.EndTime = t
		}

		runsStr := row[columnIndex["Runs"]].(string)
		if err = json.NewDecoder(bytes.NewReader([]byte(runsStr))).Decode(&executionItem.Runs); err != nil {
			return
		}

		list = append(list, *executionItem)
	}

	return
}

func GenerateProjectLogItems(response *LogEntryApiResponse, envName string) ([]ProjectLogItem, error) {
	var logItems []ProjectLogItem

	// Create a map for column index and name
	columnIndex := make(map[string]int)
	for i, column := range response.Columns {
		columnIndex[column.Name] = i
	}

	// Iterate over rows
	for _, row := range response.Rows {
		if len(row) != len(response.Columns) {
			return nil, fmt.Errorf("row length doesn't match column length")
		}

		var logEntry string
		if row[columnIndex["LogEntry"]] != nil {
			logEntryStr := row[columnIndex["LogEntry"]].(string)
			logEntry = logEntryStr
		}

		logItem := ProjectLogItem{
			LogLevel:         row[columnIndex["LogLevel"]].(string),
			Environment:      envName,
			ComponentName:    row[columnIndex["ComponentName"]].(string),
			ComponentVersion: row[columnIndex["ComponentVersion"]].(string),
			LogEntry:         logEntry,
			TimeGenerated:    row[columnIndex["TimeGenerated"]].(string),
		}

		logItems = append(logItems, logItem)
	}

	return logItems, nil
}

func (c *LogsClient) PrintProjectLogs(logItemsList []ProjectLogItem) {
	if len(logItemsList) > 0 {
		fmt.Println()
	}
	for _, logItem := range logItemsList {
		LogEntry := ""
		if logItem.LogEntry != "" {
			LogEntry = logItem.LogEntry
		}

		logMessage := fmt.Sprintf(
			"[%s] %s - %s %s: %s",
			utils.CS.Bold(logItem.LogLevel),
			utils.CS.Green(logItem.TimeGenerated),
			utils.CS.Cyan(logItem.ComponentName),
			utils.CS.CyanBold(logItem.ComponentVersion),
			LogEntry,
		)
		fmt.Println(logMessage)
	}
}

func GenerateComponentLogItems(response *LogEntryApiResponse, envName string) ([]ComponentLogItem, error) {
	var logItems []ComponentLogItem

	// Create a map for column index and name
	columnIndex := make(map[string]int)
	for i, column := range response.Columns {
		columnIndex[column.Name] = i
	}

	// Iterate over rows
	for _, row := range response.Rows {
		if len(row) != len(response.Columns) {
			return nil, fmt.Errorf("row length doesn't match column length")
		}

		var logContext string
		if row[columnIndex["LogContext"]] != nil {
			logContextStr := row[columnIndex["LogContext"]].(string)
			logContext = logContextStr
		}

		var logEntry string
		if row[columnIndex["LogEntry"]] != nil {
			logEntryStr := row[columnIndex["LogEntry"]].(string)
			logEntry = logEntryStr
		}

		logItem := ComponentLogItem{
			LogLevel:         row[columnIndex["LogLevel"]].(string),
			Environment:      envName,
			LogContext:       logContext,
			ComponentVersion: row[columnIndex["ComponentVersion"]].(string),
			LogEntry:         logEntry,
			TimeGenerated:    row[columnIndex["TimeGenerated"]].(string),
		}

		logItems = append(logItems, logItem)
	}

	return logItems, nil
}

func (c *LogsClient) PrintComponentLogs(logItemsList []ComponentLogItem) {
	if len(logItemsList) > 0 {
		fmt.Println()
	}
	for _, logItem := range logItemsList {
		logContext := ""
		if logItem.LogContext != "" {
			logContext = logItem.LogContext
		}

		var logEntry string
		if logItem.LogEntry != "" {
			logEntry = logItem.LogEntry
		}

		logMessage := fmt.Sprintf(
			"[%s] %s - %s: %s",
			utils.CS.Bold(logItem.LogLevel),
			utils.CS.Green(logItem.TimeGenerated),
			logContext,
			logEntry,
		)
		fmt.Println(logMessage)
	}
}
