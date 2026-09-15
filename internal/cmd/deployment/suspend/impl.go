package suspend

import (
	"fmt"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

func handleDeploymentSuspend(opts *SuspendCmdOpts, args []string) error {
	if len(args) > 0 {
		opts.Component = args[0]
	}

	orgId, projectId, err := common.ResolveContext(opts.Org, opts.Project)

	if err == nil {
		if opts.Org == "" {
			opts.Org = orgId
		}

		if opts.Project == "" {
			opts.Project = projectId
		}
	}

	org, err := common.ResolveTargetOrganization(opts.Org)
	if err != nil {
		return err
	}

	project, err := common.ResolveTargetProject(org, opts.Project)
	if err != nil {
		return err
	}

	component, err := common.ResolveTargetComponent(org, project.ID, opts.Component)
	if err != nil {
		return err
	}

	deploymentTrack, err := common.ResolveDeploymentTrack(component.DeploymentTracks, opts.DeploymentTrack)
	if err != nil {
		return err
	}

	env, err := common.GetProjectEnv(org.UUID, org.ID, project.ID, opts.Env)
	if err != nil {
		return err
	}

	componentDeployment, err := auth.ComponentClient.GetComponentDeployment(
		org.Handle,
		org.UUID,
		org.ID,
		component.Id,
		deploymentTrack.Id,
		env.ID)
	if err != nil {
		return err
	}

	status, err := auth.DeploymentBuildClient.SuspendDeployment(
		org.ID,
		org.Handle,
		component.Id,
		componentDeployment.ReleaseId,
		component.DisplayType)

	if status == "success" {
		fmt.Fprintln(
			utils.IO.Out,
			fmt.Sprintf(
				i18n.T("Successfully suspended deployment of %s in %s environment."),
				component.DisplayName,
				env.Name,
			),
		)
	} else {
		fmt.Fprintln(
			utils.IO.Out,
			fmt.Sprintf(i18n.T("Failed to suspended deployment of %s in %s environment."),
				component.DisplayName,
				env.Name,
			),
		)
	}

	return nil
}
