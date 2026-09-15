package logs

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/devops"
	"github.com/wso2/integration-platform-tools/pkg/api/logs"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
	commonUtils "github.com/wso2/integration-platform-tools/pkg/util/common"
)

func resolveLogType(logType *string) error {
	if *logType != "" {
		if !commonUtils.StringExistsInSlice(*logType, LogTypes) {
			return fmt.Errorf("%s", heredoc.Docf(i18n.T(`
				Invalid log type
				Log type must be one of %s
			`), strings.Join(LogTypes, ", ")))
		}
	} else {
		err := prompt.NewPromptSelectMessage[string](
			prompt.PromptSelectOpts[string]{Message: i18n.T("Log type:"), Values: LogTypes},
			logType,
		).Prompt()

		if err != nil {
			return err
		}
	}
	return nil
}

func printArgoBuildLogs(
	params LogsParams,
	orgId string,
	component models.Component,
	selectedDataPlaneHost string,
	deploymentTrackId string,
	workflowName string,
) error {
	buildLog, err := common.GetArgoBuildLog(orgId, selectedDataPlaneHost, component, deploymentTrackId, workflowName)
	if err != nil {
		return err
	}

	decodedLog, err := base64.StdEncoding.DecodeString(buildLog.Data.Build.Log)
	if err != nil {
		return err
	}

	logString := string(decodedLog)
	fmt.Println(logString)

	return nil
}

func printComponentLogs(
	params LogsParams,
	orgId string,
	component models.Component,
	selectedEnv project.ProjectEnvironment,
	selectedDataPlaneHost string,
	isCilium bool,
	apiVersionList []string,
	versionIDList []string,
) error {
	reqBody := logs.GetComponentLogsReqBody{
		ComponentID:   component.Id,
		EndTime:       time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
		Limit:         params.limitFlag,
		SearchPhrase:  params.queryFlag,
		Sort:          "asc",
		SortingOrder:  "asc",
		StartTime:     time.Now().Add(-24 * time.Hour).Format("2006-01-02T15:04:05.999Z"),
		EnvironmentID: selectedEnv.ID,
		VersionList:   apiVersionList,
		VersionIDList: versionIDList,
	}

	logsSpinner := utils.CreateSpinner(
		fmt.Sprintf(
			i18n.T(" Fetching %s logs for component %s within %s environment"),
			params.typeFlag,
			component.Name,
			selectedEnv.Name,
		),
		"\n",
	)
	logsSpinner.Start()

	logs, err := auth.LogsClient.GetComponentLogs(
		reqBody,
		orgId,
		selectedDataPlaneHost,
		isCilium,
		selectedEnv.Name,
		params.typeFlag,
		false,
	)

	logsSpinner.Stop()

	if err != nil {
		return fmt.Errorf(i18n.T("failed to get component logs: %w"), err)
	}

	if len(logs) == 0 {
		if params.followFlag {
			fmt.Println(i18n.T("Watching for new logs..."))
		} else {
			fmt.Println(i18n.T("No recent logs found"))
		}
	}

	auth.LogsClient.PrintComponentLogs(logs)

	if params.followFlag {
		// Create a ticker that ticks every 5 seconds
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		// Create a channel to listen for SIGINT (Ctrl+C) signal
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

		for {
			select {
			case <-ticker.C:
				// Call the function every time the ticker ticks
				reqBody.StartTime = reqBody.EndTime
				reqBody.EndTime = time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
				go func() {
					logs, err := auth.LogsClient.GetComponentLogs(
						reqBody,
						orgId,
						selectedDataPlaneHost,
						isCilium,
						selectedEnv.Name,
						params.typeFlag,
						true,
					)
					if err != nil {
						utils.HandleErr(fmt.Errorf(i18n.T("failed to follow component logs: %w"), err))
					}
					auth.LogsClient.PrintComponentLogs(logs)
				}()
			case <-interrupt:
				// If SIGINT is received, stop the function and return
				return nil
			}
		}
	}
	return nil
}

func printProjectLogs(
	params LogsParams,
	project models.Project,
	selectedEnv project.ProjectEnvironment,
	selectedDataPlaneHost string,
	isCilium bool,
) error {
	reqBody := logs.GetProjectLogsReqBody{
		ComponentIdList: []string{},
		EndTime:         time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
		Limit:           params.limitFlag,
		LogLevels:       []string{},
		ProjectID:       project.ID,
		SearchPhrase:    params.queryFlag,
		Sort:            "asc",
		SortingOrder:    "asc",
		StartTime:       time.Now().Add(-24 * time.Hour).Format("2006-01-02T15:04:05.999Z"),
		EnvironmentID:   selectedEnv.ID,
	}

	logsSpinner := utils.CreateSpinner(fmt.Sprintf(i18n.T(" Fetching project level logs for the %s environment"), selectedEnv.Name), "\n")
	logsSpinner.Start()
	logs, err := auth.LogsClient.GetProjectLogs(reqBody, strconv.Itoa(project.OrgID), selectedDataPlaneHost, isCilium, selectedEnv.Name, false)
	logsSpinner.Stop()

	if err != nil {
		return fmt.Errorf(i18n.T("failed to get project logs: %w"), err)
	}

	if len(logs) == 0 {
		if params.followFlag {
			fmt.Println(i18n.T("Watching for new logs..."))
		} else {
			fmt.Println(i18n.T("No recent logs found"))
		}
	}

	auth.LogsClient.PrintProjectLogs(logs)

	if params.followFlag {
		// Create a ticker that ticks every 5 seconds
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		// Create a channel to listen for SIGINT (Ctrl+C) signal
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

		for {
			select {
			case <-ticker.C:
				// Call the function every time the ticker ticks
				reqBody.StartTime = reqBody.EndTime
				reqBody.EndTime = time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
				go func() {
					logs, err := auth.LogsClient.GetProjectLogs(reqBody, strconv.Itoa(project.OrgID), selectedDataPlaneHost, isCilium, strings.ToLower(selectedEnv.Name), true)
					if err != nil {
						utils.HandleErr(fmt.Errorf(i18n.T("failed to follow project logs: %w"), err))
					}
					auth.LogsClient.PrintProjectLogs(logs)
				}()
			case <-interrupt:
				// If SIGINT is received, stop the function and return
				return nil
			}
		}
	}
	return nil
}

func printMiBuildLogs(orgId string, componentId string, runId int) error {
	buildLogData, err := common.GetMiComponentBuildLogs(orgId, componentId, runId)
	if err != nil {
		return err
	}

	decodedLog, err := base64.StdEncoding.DecodeString(buildLogData)
	if err != nil {
		return err
	}

	logString := string(decodedLog)
	fmt.Println(logString)

	return nil
}

func printBuildLogs(orgHandle, orgId, projectId, compType, componentId string, runId int) error {
	buildLogData, err := common.GetComponentBuildLogs(orgHandle, orgId, projectId, componentId, runId)
	if err != nil {
		return err
	}

	decodedLog := make([]byte, 0)

	if compType == component.DisplayTypeGitProxy {
		governaceLogs, err := common.GetBuildLogForPhase(orgId, componentId, runId, "governanceLogs")
		if err != nil {
			return err
		}

		oasLog, err := base64.StdEncoding.DecodeString(buildLogData.Data.OasValidation.Log)
		if err != nil {
			return err
		}

		updateApiLog, err := base64.StdEncoding.DecodeString(buildLogData.Data.UpdateApi.Log)
		if err != nil {
			return err
		}

		governanceAPILog, err := base64.StdEncoding.DecodeString(governaceLogs.GovernanceLogs)
		if err != nil {
			return err
		}

		decodedLog = append(decodedLog, []byte("\n\nAPI Definition Validation\n\n")...)
		decodedLog = append(decodedLog, oasLog...)
		decodedLog = append(decodedLog, []byte("\n\nAPI Update\n\n")...)
		decodedLog = append(decodedLog, updateApiLog...)
		decodedLog = append(decodedLog, []byte("\n\nAPI Governance\n\n")...)
		decodedLog = append(decodedLog, []byte(governanceAPILog)...)

	} else {
		decodedLog, err = base64.StdEncoding.DecodeString(buildLogData.Data.Build.Log)
		if err != nil {
			return err
		}
	}

	logString := string(decodedLog)
	fmt.Println(logString)

	return nil
}

func getDataPlaneInfo(orgId string, orgUuid string) (cloudDataPlanes []devops.GatewayInfo, dataPlanes []devops.DataPlaneItem, err error) {
	dataPlaneInfo := utils.CreateSpinner(i18n.T(" Fetching data plane information..."), "")
	dataPlaneInfo.Start()

	cloudDataPlanes, err = auth.DevopsClient.GetCloudPlaneClusters(orgUuid, orgId)
	if err != nil {
		return nil, nil, fmt.Errorf(i18n.T("failed to fetch cloud data plane information: %w"), err)
	}

	dataPlanes, err = auth.DevopsClient.GetDataPlaneClusters(orgId)
	if err != nil {
		return nil, nil, fmt.Errorf(i18n.T("failed to fetch data plane information: %w"), err)
	}

	dataPlaneInfo.Stop()

	return cloudDataPlanes, dataPlanes, nil
}

func resolveDataPlane(selectedEnv project.ProjectEnvironment, cloudDataPlanes []devops.GatewayInfo, dataPlanes []devops.DataPlaneItem) (selectedDataPlaneHost string, isCilium bool) {
	for _, dataPlaneItem := range cloudDataPlanes {
		if dataPlaneItem.ID == selectedEnv.DPID {
			selectedDataPlaneHost = dataPlaneItem.ExternalGatewayVirtualHost
			isCilium = dataPlaneItem.IsCilium
			break
		}
	}

	if selectedDataPlaneHost == "" {
		for _, dataPlaneItem := range dataPlanes {
			if dataPlaneItem.ID == selectedEnv.DPID {
				selectedDataPlaneHost = dataPlaneItem.ExternalGatewayVirtualHost
				break
			}
		}
	}

	return selectedDataPlaneHost, isCilium
}
