package application

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/logs"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

const logType = "application"

func printApplicationLogs(opts *ApplicationLogsOpts) (err error) {
	orgId, projectId, err := common.ResolveContext(opts.org, opts.project)

	if err == nil {
		if opts.org == "" {
			opts.org = orgId
		}

		if opts.project == "" {
			opts.project = projectId
		}
	}

	org, err := common.ResolveTargetOrganization(opts.org)

	if err != nil {
		return
	}

	project, err := common.ResolveTargetProject(org, opts.project)

	if err != nil {
		return
	}

	cmp, err := common.ResolveTargetComponent(org, project.ID, opts.component)

	if err != nil {
		return
	}

	if cmp.DisplayType == component.DisplayTypeProxy {
		return errors.New(i18n.T("application logs is not applicable for proxy type components"))
	}

	dtrack, err := common.ResolveDeploymentTrack(cmp.DeploymentTracks, opts.deploymentTrack)

	if err != nil {
		return
	}

	env, err := common.GetProjectEnv(org.UUID, org.ID, project.ID, opts.env)

	if err != nil {
		return
	}

	cloudDataPlanes, dataPlanes, err := common.GetDataPlaneInfo(org.ID, org.UUID)
	if err != nil {
		return err
	}

	selectedDataPlaneHost, isCilium := common.ResolveDataPlane(*env, cloudDataPlanes, dataPlanes)

	err = printComponentLogs(opts, org.ID, *cmp, *env, selectedDataPlaneHost, isCilium, []string{dtrack.ApiVersion}, []string{dtrack.Id})
	if err != nil {
		return err
	}
	return nil
}

func printComponentLogs(
	opts *ApplicationLogsOpts,
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
		Limit:         opts.limitFlag,
		SearchPhrase:  opts.queryFlag,
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
			logType,
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
		logType,
		false,
	)

	logsSpinner.Stop()

	if err != nil {
		return fmt.Errorf(i18n.T("failed to get component logs: %w"), err)
	}

	if len(logs) == 0 {
		if opts.followFlag {
			fmt.Println(i18n.T("Watching for new logs..."))
		} else {
			fmt.Println(i18n.T("No recent logs found"))
		}
	}

	auth.LogsClient.PrintComponentLogs(logs)

	if opts.followFlag {
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
						logType,
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
