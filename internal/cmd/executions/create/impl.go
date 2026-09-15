package create

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
)

func handleImpl(params *CreateExecutionParams) error {

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

	dTrack, err := common.ResolveDeploymentTrack(sComponent.DeploymentTracks, params.DeploymentTrack)
	if err != nil {
		return err
	}

	sEnv, err := common.GetProjectEnv(sOrg.UUID, sOrg.ID, sProject.ID, params.Env)
	if err != nil {
		return err
	}

	depInfo, err := common.GetComponentDeplayment(sOrg.ID, sOrg.Handle, sOrg.UUID, sComponent.Id, dTrack.Id, sEnv.ID)
	if err != nil {
		return err
	}

	spinner := utils.CreateSpinner("Queuing execution...", "")
	spinner.Start()
	_, err = auth.ComponentClient.TriggerExecution(
		sOrg.ID, sOrg.Handle, sProject.ID, sComponent.Id, depInfo.ReleaseId, []string{})
	spinner.Stop()

	if err != nil {
		return err
	}

	utils.PrintInfo("\nSuccessfully triggered execution.\n\n")

	utils.PrintInfo("%s", heredoc.Docf(`

			To view the list of triggered executions:
            %s

			To view details of the triggered execution:
            %s

	`,
		fmt.Sprintf("wso2-integration-platform list executions --project '%s' --component '%s' --env %s", sProject.Name, sComponent.Name, sEnv.Name),
		fmt.Sprintf("wso2-integration-platform describe execution --project '%s' --component %s --env %s\n", sProject.Name, sComponent.Name, sEnv.Name),
	))

	return nil
}
