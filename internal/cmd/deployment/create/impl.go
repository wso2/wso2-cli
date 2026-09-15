package create

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

func HandleDeployComponent(params *DeployComponentParams) error {
	orgId, projectId, err := common.ResolveContext(params.OrgFlag, params.ProjectFlag)

	if err == nil {
		if params.OrgFlag == "" {
			params.OrgFlag = orgId
		}

		if params.ProjectFlag == "" {
			params.ProjectFlag = projectId
		}
	}

	selectedOrg, err := common.ResolveTargetOrganization(params.OrgFlag)
	if err != nil {
		return err
	}

	project, err := common.ResolveTargetProject(selectedOrg, params.ProjectFlag)
	if err != nil {
		return err
	}

	remoteComponent, err := common.ResolveTargetComponent(selectedOrg, project.ID, params.ComponentFlag)
	if err != nil {
		return err
	}

	remoteComWithRepoData, err := common.GetComponentWithRepoData(selectedOrg.ID, remoteComponent.Handler, project.ID)
	if err != nil {
		return err
	}

	resp, err := auth.ComponentClient.GetComponentInitStatus(selectedOrg.ID, selectedOrg.Handle, project.ID, remoteComponent.Id)
	if err != nil {
		return err
	}

	if resp.Data.Status == "queued" || resp.Data.Status == "in_progress" {
		utils.PrintInfo("%s", i18n.T("Component initialization is in progress. Please run the command in a while...\n"))
		return nil
	}

	deploymentTrack, err := common.ResolveDeploymentTrack(remoteComWithRepoData.DeploymentTracks, params.DeploymentTrackFlag)
	if err != nil {
		return err
	}

	selectedEnv, err := common.GetProjectEnv(selectedOrg.UUID, selectedOrg.ID, project.ID, params.EnvFlag)
	if err != nil {
		return err
	}

	versionId := deploymentTrack.Id

	var runId int = 0

	if len(params.RunIdFlag) > 0 {
		runId, err = strconv.Atoi(params.RunIdFlag)
		if err != nil {
			return err
		}
	}

	if strings.HasPrefix(strings.ToLower(remoteComponent.DisplayType), "byoi") {
		err := handleByoiDeployment(*selectedOrg, project, remoteComWithRepoData, selectedEnv, deploymentTrack, versionId, params)
		if err != nil {
			return err
		}
	} else if remoteComponent.DisplayType == component.DisplayTypeGitProxy {
		err = handleProxyDeployment(selectedOrg, project, remoteComWithRepoData, selectedEnv, deploymentTrack, runId, versionId, params)
		if err != nil {
			return err
		}

	} else {
		err = handleDeployment(selectedOrg, project, remoteComWithRepoData, selectedEnv, deploymentTrack, runId, versionId, params)
		if err != nil {
			return err
		}
	}

	utils.PrintInfo("%s", heredoc.Docf(`

			Deployment for component '%s' has been successfully triggered
		`,
		remoteComponent.Name,
	))

	time.Sleep(1 * time.Second)

	utils.PrintInfo("%s", heredoc.Docf(`

			To view deployment status of the component deployed in %s environment :
				%s
		`,
		selectedEnv.Name,
		fmt.Sprintf(`$ wso2-integration-platform describe component "%s" --project="%s"`, remoteComponent.Name, project.Name),
	))

	if strings.HasSuffix(remoteComponent.DisplayType, "Service") {
		utils.PrintInfo("%s", heredoc.Docf(`

			To get the test key needed to invoke the component deployed in %s environment :
				%s
		`,
			selectedEnv.Name,
			fmt.Sprintf(
				`$ wso2-integration-platform create test-key --project="%s" --component="%s" --deployment-track="%s" --env="%s"`,
				project.Name,
				remoteComponent.Name,
				deploymentTrack.Branch,
				selectedEnv.Name,
			),
		))
	}

	utils.PrintInfo("%s", heredoc.Docf(`

			To view application logs of the component deployed in %s environment :
				%s
		`,
		selectedEnv.Name,
		fmt.Sprintf(
			`$ wso2-integration-platform logs application --project="%s" --component="%s" --deployment-track="%s" --env=%s`,
			project.Name,
			remoteComponent.Name,
			deploymentTrack.Branch,
			selectedEnv.Name,
		),
	))

	return nil
}

func handleByoiDeployment(selectedOrg api.Organization, project *models.Project, remoteComWithRepoData models.Component, selectedEnv *project.ProjectEnvironment, deploymentTrack *models.DeploymentTrack, versionId string, params *DeployComponentParams) error {
	matchingAppEnv, err := common.GetReleaseEnvForDeploymentTrack(remoteComWithRepoData, deploymentTrack.Id, selectedEnv.ID)
	if err != nil {
		return err
	}
	imageHistory, err := getImageHistory(selectedOrg.ID, selectedOrg.UUID, project.ID, remoteComWithRepoData.Id, versionId)
	if err != nil {
		return err
	}

	resolvedImage, err := resolveImage(params.byoiDeployOpts.ImageWithTag, imageHistory)
	if err != nil {
		return err
	}

	err = generateByoiEndpoint(selectedOrg, *project, remoteComWithRepoData, matchingAppEnv, params.byoiDeployOpts)
	if err != nil {
		return err
	}

	success, err := deployByoiComponent(selectedOrg.ID, remoteComWithRepoData.Id, matchingAppEnv.ReleaseId, resolvedImage.ImageNameWithTag)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("failed to deploy image")
	}

	return nil
}

func handleDeployment(selectedOrg *api.Organization, project *models.Project, remoteComWithRepoData models.Component, selectedEnv *project.ProjectEnvironment, deploymentTrack *models.DeploymentTrack, runId int, versionId string, params *DeployComponentParams) error {
	selectedBuild, err := common.ResolveSucceededBuild(selectedOrg, project.Handler, versionId, remoteComWithRepoData.Name, runId)
	if err != nil {
		return err
	}
	err = setDeployEnvConfigs(*selectedOrg, *project, remoteComWithRepoData, versionId, *selectedEnv, params.EnvVars)
	if err != nil {
		return err
	}

	err = GenerateEndpoints(remoteComWithRepoData, *selectedEnv, deploymentTrack.Id, selectedBuild.Spec.Revision, project.ID, selectedOrg.Handle, selectedOrg.ID)
	if err != nil {
		return err
	}

	if len(selectedBuild.Status.Images) == 0 {
		return fmt.Errorf("build does not contain any images")
	}

	if len(selectedEnv.PromoteFrom) > 0 {

		depInfo, err := auth.ComponentClient.GetComponentDeployment(
			selectedOrg.Handle,
			selectedOrg.UUID,
			selectedOrg.ID,
			remoteComWithRepoData.Id,
			deploymentTrack.Id,
			selectedEnv.PromoteFrom[0],
		)

		if err != nil {
			if errors.Is(err, api.ErrNotFound) {
				err = fmt.Errorf("%s", heredoc.Docf(`
						Deployment Pipeline Error: No deployment detected in the upstream environment.
						Please ensure that the required deployment is successfully completed in the previous environment before proceeding.
					`))
			}
			return err
		}

		if rid, err := strconv.Atoi(depInfo.Build.RunId); err != nil {
			return err
		} else if rid != selectedBuild.Status.RunID {
			err = fmt.Errorf("%s", heredoc.Docf(`
					Deployment Pipeline Error: The build must be deployed in the upstream environment before continuing.
					Please complete the deployment in the prior environment and try again.
				`))
			return err
		}

		if selectedEnv.Critical {
			if err := verifyWorkflowStatus(*selectedOrg, *selectedEnv, project.ID, remoteComWithRepoData.Handler, depInfo.Build.BuildId); err != nil {
				return err
			}
		}
	}

	_, err = triggerComponentDeployments(
		selectedOrg.ID,
		remoteComWithRepoData.Name,
		component.GetTypeForDisplayType(remoteComWithRepoData.DisplayType),
		project.Handler,
		deploymentTrack.Id,
		selectedEnv.Name,
		selectedBuild.Status.Images[0].ID,
		params,
	)

	return nil
}

func handleProxyDeployment(selectedOrg *api.Organization, project *models.Project, remoteComWithRepoData models.Component, selectedEnv *project.ProjectEnvironment, deploymentTrack *models.DeploymentTrack, runId int, versionId string, params *DeployComponentParams) error {
	apiId := ""
	if remoteComWithRepoData.DisplayType == component.DisplayTypeGitProxy {
		for _, v := range remoteComWithRepoData.ApiVersions {
			if v.Latest {
				versionId = v.VersionId
				apiId = v.Id
			}
		}
	}

	selectedBuild, err := common.ResolveSucceededBuild(selectedOrg, project.Handler, versionId, remoteComWithRepoData.Name, runId)
	if err != nil {
		return err
	}
	if len(selectedEnv.PromoteFrom) > 0 {
		depInfo, err := auth.DeploymentBuildClient.GetProxyDeploymentInfo(
			selectedOrg.ID,
			selectedOrg.Handle,
			selectedOrg.UUID,
			remoteComWithRepoData.Id,
			apiId,
			selectedEnv.PromoteFrom[0],
		)

		if err != nil {
			if errors.Is(err, api.ErrNotFound) {
				err = fmt.Errorf("Deployment pipeline violation. No deployment found in the upstream environment.")
			}
			return err
		}

		if selectedEnv.Critical {
			if err := verifyWorkflowStatus(*selectedOrg, *selectedEnv, project.ID, remoteComWithRepoData.Handler, depInfo.Build.ID); err != nil {
				return err
			}

			if params.proxyDeployOpts.TargetEp != "" || params.proxyDeployOpts.SandboxEp != "" {
				err = auth.ApimPublisherClient.UpdateApiProxyKeys(
					selectedOrg.ID,
					apiId,
					selectedEnv.APIMEnvID,
					selectedOrg.UUID,
					params.proxyDeployOpts.TargetEp,
					params.proxyDeployOpts.SandboxEp,
				)

				if err != nil {
					return err
				}

			}

			_, err := auth.DeploymentBuildClient.PromoteProxyComponent(
				selectedOrg.ID,
				remoteComWithRepoData.Id,
				apiId,
				selectedEnv.PromoteFrom[0],
				selectedEnv.ID,
				depInfo.Build.ID,
			)

			if err != nil {
				return err
			}

			utils.PrintInfo("%s", heredoc.Docf(`

                    Deployment for component '%s' has been successfully triggered
                `,
				remoteComWithRepoData.Name,
			))

			return nil
		}
	}

	_, err = triggerComponentDeployments(
		selectedOrg.ID,
		remoteComWithRepoData.Name,
		component.GetTypeForDisplayType(remoteComWithRepoData.DisplayType),
		project.Handler,
		versionId,
		selectedEnv.Name,
		fmt.Sprintf("%d", selectedBuild.Status.RunID),
		params,
	)
	return nil
}
