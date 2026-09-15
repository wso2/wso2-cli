package resume

import (
	"fmt"

	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

type RedeployCmdOpts struct {
	Org             string
	Project         string
	Component       string
	DeploymentTrack string
	Env             string
}

var cmdOpts RedeployCmdOpts

var RedeployCmd = &cobra.Command{
	Use:    "deployment [component-name] [flags]",
	Short:  "Resume a deployment",
	Long:   "Resume a suspended deployment",
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handleRedeploy(&cmdOpts, args); err != nil {
			utils.HandleErr(err)
		}
	},
}

func handleRedeploy(opts *RedeployCmdOpts, args []string) error {

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

	status, err := auth.DeploymentBuildClient.ResumeDeployment(
		org.ID,
		org.Handle,
		component.Id,
		componentDeployment.ReleaseId,
		component.DisplayType,
	)

	if status == "success" {
		fmt.Fprintln(
			utils.IO.Out,
			fmt.Sprintf(
				i18n.T("Successfully redeployed deployment of %s in %s environment."),
				component.DisplayName,
				env.Name,
			),
		)
	} else {
		fmt.Fprintln(
			utils.IO.ErrOut,
			fmt.Sprintf(
				i18n.T("Failed to redeploy deployment of %s in %s environment."),
				component.DisplayName,
				env.Name,
			),
		)
	}

	return nil
}

func init() {
	common.AddOrgFlag(RedeployCmd.Flags(), &cmdOpts.Org)
	common.AddProjectFlag(RedeployCmd.Flags(), &cmdOpts.Project)
	common.AddEnvFlag(RedeployCmd.Flags(), &cmdOpts.Env)
	common.AddDeploymentTrackFlag(RedeployCmd.Flags(), &cmdOpts.DeploymentTrack)
}
