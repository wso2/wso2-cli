package list

import (
	"fmt"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

func handleImpl(params *ListExecutionParams) error {
	// An invalid --output falls back to a persisted OUTPUT_FORMAT (with a
	// warning); with no global set it is an error.
	outputFormat, err := common.ResolveOutputFormat(params.OutputFlag)
	if err != nil {
		return err
	}

	orgId, projectId, err := common.ResolveContext(params.Org, params.Project)

	if err == nil {
		if params.Org == "" {
			params.Org = orgId
		}

		if params.Project == "" {
			params.Project = projectId
		}
	}

	sOrg, err := common.ResolveTargetOrganization(params.Org)
	if err != nil {
		return err
	}

	sProject, err := common.ResolveTargetProject(sOrg, params.Project)
	if err != nil {
		return err
	}

	sComponent, err := common.ResolveTargetComponent(sOrg, sProject.ID, params.Component)
	if err != nil {
		return err
	}

	switch sComponent.DisplayType {
	case component.DisplayTypeManualTrigger, component.DisplayTypeScheduledTask, component.DisplayTypeBuildpackJob,
		component.DisplayTypeMiJob, component.DisplayTypeMiCronjob, component.DisplayTypeBuildpackCronJob,
		component.DisplayTypeByocCronjob, component.DisplayTypeByoiCronjob, component.DisplayTypeByocJob,
		component.DisplayTypeByoiJob:
	default:
		return fmt.Errorf("executions apply only to Automation integrations (scheduled tasks)")
	}

	cInitStat, err := auth.ComponentClient.GetComponentInitStatus(sOrg.ID, sOrg.Handle, sProject.ID, sComponent.Id)
	if err != nil {
		return err
	}

	if cInitStat.Data.Status == "queued" || cInitStat.Data.Status == "in_progress" {
		utils.PrintInfo("%s", i18n.T("Component initialization is in progress. Please run the command in a while...\n"))
		return nil
	}

	sEnv, err := common.GetProjectEnv(sOrg.UUID, sOrg.ID, sProject.ID, params.Env)
	if err != nil {
		return err
	}

	var latestVersion models.ApiVersion
	for _, versionItem := range sComponent.ApiVersions {
		if versionItem.Latest {
			latestVersion = versionItem
			break
		}
	}

	deploymentStatusSpinner := utils.CreateSpinner(
		fmt.Sprintf(i18n.T(" Fetching deployment status for the %s environment"), sEnv.Name),
		"")
	deploymentStatusSpinner.Start()
	componentDeployment, err := auth.ComponentClient.GetComponentDeployment(
		sOrg.Handle,
		sOrg.UUID,
		sOrg.ID,
		sComponent.Id,
		latestVersion.Id,
		sEnv.ID)
	deploymentStatusSpinner.Stop()

	if err != nil {
		return err
	}

	cloudDplanes, dplanes, err := common.GetDataPlaneInfo(sOrg.ID, sOrg.UUID)
	if err != nil {
		return err
	}

	dpHost, isCilium := common.ResolveDataPlane(*sEnv, cloudDplanes, dplanes)

	execSpinner := utils.CreateSpinner("Fetching executions...", "")
	execSpinner.Start()
	executions, err := auth.LogsClient.GetExecutionsListV2(
		sOrg.ID,
		dpHost,
		componentDeployment.ReleaseId,
		isCilium,
		params.Limit,
	)
	execSpinner.Stop()

	if err != nil {
		return err
	}

	if outputFormat.IsStructured() {
		return printExecutionsStructured(executions, outputFormat, *sComponent, *sProject)
	}

	printExecutionsTable(executions, *sComponent, *sProject)
	return nil
}
