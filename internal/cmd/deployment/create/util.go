package create

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api"
	deployment "github.com/wso2/integration-platform-tools/pkg/api/deploymentBuild"
	"github.com/wso2/integration-platform-tools/pkg/api/devops"
	"github.com/wso2/integration-platform-tools/pkg/api/gqlbuild"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
	workflowmgt "github.com/wso2/integration-platform-tools/pkg/api/workflow-mgt"
)

func triggerComponentDeployments(
	orgId string,
	componentName string,
	componentType string,
	projectHandle string,
	deploymentTrackId string,
	envName string,
	buildRef string,
	params *DeployComponentParams,
) (deployments *deployment.DeploymentTrackResponse, err error) {
	getDeploymentSpinner := utils.CreateSpinner(i18n.T(" Triggering component deployment..."), "")

	deployOpts := deployment.DeploySpecOpts{}

	switch componentType {
	case "scheduled-task":
		deployOpts.ScheduleExp = params.cronDeployOpts.Expression
		deployOpts.ScheduleTZ = params.cronDeployOpts.TimeZone
	case "proxy":
		if params.proxyDeployOpts.SandboxEp != "" || params.proxyDeployOpts.TargetEp != "" {
			deployOpts.ProxyConf = &deployment.ProxyDeploymentConfig{
				Keys: &deployment.ProxyTargetEndpoint{
					SandboxEndpoint:    &params.proxyDeployOpts.SandboxEp,
					ProductionEndpoint: params.proxyDeployOpts.TargetEp,
				},
				// TODO: check if AccessMode also needs to be passed as param
			}
		}
	}

	getDeploymentSpinner.Start()
	var deploymentsRes deployment.DeploymentTrackResponse
	deploymentsRes, err = auth.DeploymentBuildClient.CreateDeployment(
		orgId,
		componentName,
		projectHandle,
		deploymentTrackId,
		envName,
		buildRef,
		deployOpts,
	)
	getDeploymentSpinner.Stop()
	if err != nil {
		return nil, err
	}
	return &deploymentsRes, nil
}

func GenerateEndpoints(comp models.Component, env project.ProjectEnvironment, deploymentTrackId string, commitHash, projectId string, orgHandle string, orgId string) error {
	if strings.HasSuffix(comp.DisplayType, "Service") {
		remoteComWithRepoData, err := common.GetComponentWithRepoData(orgId, comp.Handler, projectId)
		if err != nil {
			return err
		}

		matchingAppEnv, err := common.GetReleaseEnvForDeploymentTrack(remoteComWithRepoData, deploymentTrackId, env.ID)
		if err != nil {
			return err
		}

		generateEndpointsSpinner := utils.CreateSpinner(i18n.T(" Generating endpoints..."), "")
		generateEndpointsSpinner.Start()

		_, err = auth.ComponentClient.GenerateEndPoints(
			comp.Id,
			deploymentTrackId,
			matchingAppEnv.ReleaseId,
			commitHash,
			orgId,
		)

		generateEndpointsSpinner.Stop()

		if err != nil {
			return err
		}
	}

	return nil
}

func setDeployEnvConfigs(
	selectedOrg api.Organization,
	project models.Project,
	remoteComWithRepoData models.Component,
	deploymentTrackId string,
	selectedEnv project.ProjectEnvironment,
	envVars []common.KeyValOpt) error {

	matchingAppEnv, err := common.GetReleaseEnvForDeploymentTrack(remoteComWithRepoData, deploymentTrackId, selectedEnv.ID)
	if err != nil {
		return err
	}

	configsMounts, err := common.GetConfigMounts(selectedOrg.ID, selectedOrg.UUID, remoteComWithRepoData.Id, matchingAppEnv.ReleaseId, project.ID)
	if err != nil {
		return err
	}

	configs, err := common.GetConfigsOfComponent(selectedOrg.ID, selectedOrg.UUID, selectedEnv.ID, project.ID, matchingAppEnv.ReleaseId, remoteComWithRepoData.Id, configsMounts)
	if err != nil {
		return err
	}

	configName := fmt.Sprintf("%s-%s", deployConfigMapBaseName, remoteComWithRepoData.Id)

	selectedConfig, err := common.ResolveTargetConfig(configs, configName)
	if err == nil {
		// delete config-map if it exists
		mountItem, err := common.GetMountForConfig(selectedConfig, configsMounts)
		if err != nil {
			return err
		}

		err = common.DeleteConfigMount(selectedOrg.ID, selectedOrg.UUID, remoteComWithRepoData.Id, matchingAppEnv.ReleaseId, mountItem.ContainerID, mountItem.ID, project.ID)
		if err != nil {
			return err
		}
	}

	if len(envVars) > 0 {
		err = common.CreateConfig(common.CreateConfigParams{
			ConfigName:    configName,
			ConfigType:    common.ConfigTypeConfigMap,
			MountType:     common.MountTypeEnv,
			EnvVars:       envVars,
			FileMountPath: "",
			OrgId:         selectedOrg.ID,
			OrgUuid:       selectedOrg.UUID,
			ProjectId:     project.ID,
			EnvId:         selectedEnv.ID,
			AppEnvId:      matchingAppEnv.ReleaseId,
			ComponentId:   remoteComWithRepoData.Id,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func resolveProxyBuild(builds []gqlbuild.BuildListResponse, buildId string) (build gqlbuild.BuildListResponse, err error) {
	if len(builds) == 0 {
		err = fmt.Errorf("No succeeded builds found for the selected component")
		return
	}

	for _, val := range builds {
		if val.BuildId == buildId {
			build = val
			return
		}
	}

	var promptList []huh.Option[string]
	var promptSelection string
	for _, item := range builds {
		promptList = append(
			promptList,
			huh.Option[string]{
				Key: fmt.Sprintf(
					"%s - %s",
					item.BuildId,
					fmt.Sprintf("%s ago", utils.GetRelativeTime(item.CreatedDate)),
				),
				Value: item.BuildId,
			},
		)
	}

	promptList = reverseArray(promptList)

	err = prompt.NewPromptSelectMessage(
		prompt.PromptSelectOpts[string]{
			Message: "Build ID:",
			Options: promptList,
		},
		&promptSelection,
	).Prompt()

	if err != nil {
		return
	}

	for _, val := range builds {
		if val.BuildId == promptSelection {
			build = val
			return
		}
	}

	return
}

func reverseArray[T comparable](arr []T) []T {
	if len(arr) == 0 {
		return arr
	}

	return append(reverseArray(arr[1:]), arr[0])
}

func resolveEnv(envName string, envs []project.ProjectEnvironment) (env *project.ProjectEnvironment, err error) {
	opts := make([]huh.Option[*project.ProjectEnvironment], len(envs))

	for i, v := range envs {
		opts[i] = huh.NewOption(v.Name, &v)
		if envName == v.Name {
			return &v, nil
		}
	}

	if env == nil {
		err = prompt.NewPromptSelectMessage[*project.ProjectEnvironment](
			prompt.PromptSelectOpts[*project.ProjectEnvironment]{
				Options: opts,
				Message: i18n.T("Environment: "),
			},
			&env,
		).Prompt()

		if err != nil {
			utils.PrintInfo(i18n.T("%s Failed to select environment\n"), utils.CS.Red("!"))
			return nil, err
		}
	}

	return
}

func getBuildList(orgId, cmpId, versionId string) ([]gqlbuild.BuildListResponse, error) {
	buildListSpinner := utils.CreateSpinner(i18n.T(" Fetching builds..."), "")
	buildListSpinner.Start()
	buildList, err := auth.GQLBuildClient.GetBuildList(orgId, cmpId, versionId)
	buildListSpinner.Stop()

	if err != nil {
		return nil, err
	}

	return buildList, nil
}

func getImageHistory(orgId, orgIdUuid, projectId, compId, versionId string) ([]devops.ImageHistoryResponse, error) {
	imageHistorySpinner := utils.CreateSpinner(i18n.T(" Fetching component images..."), "")
	imageHistorySpinner.Start()
	imageHistory, err := auth.DevopsClient.GetImageHistory(orgId, orgIdUuid, projectId, compId, versionId)

	imageHistorySpinner.Stop()

	if err != nil {
		return nil, err
	}

	return imageHistory, nil
}

func deployByoiComponent(orgId, compId, envReleaseId, imageNameWithTag string) (bool, error) {
	imageHistorySpinner := utils.CreateSpinner(i18n.T(" Deploying image based component..."), "")
	imageHistorySpinner.Start()
	success, err := auth.ComponentClient.DeployByoiComponent(orgId, compId, envReleaseId, imageNameWithTag)

	imageHistorySpinner.Stop()

	if err != nil {
		return false, err
	}

	return success, nil
}

func genComponentDeployUrl(orgHandle, projectId, componentHandle string) string {
	envConfigs := auth.GetEnvConfig()

	return fmt.Sprintf(
		"%s/organizations/%s/projects/%s/components/%s/deploy",
		envConfigs.ConsoleUrls.BaseUrl,
		orgHandle,
		projectId,
		componentHandle,
	)
}

func verifyWorkflowStatus(
	org api.Organization,
	env project.ProjectEnvironment,
	projectId string,
	cmpHandle string,
	buildId string,
) (err error) {
	workflowStatusSpinner := utils.CreateSpinner(i18n.T(" Fetching builds..."), "")
	workflowStatusSpinner.Start()
	wfStatus, err := auth.WorkflowMgtClient.CheckWorkflowStatus(
		org.ID,
		buildId,
		env.ID,
	)
	workflowStatusSpinner.Stop()
	if err != nil {
		return err
	}
	switch wfStatus.Status {
	case workflowmgt.WFS_PENDING:
		// wait for approval
		return errors.New("Workflow aproval is pending, please run the command once approved")
	case workflowmgt.WFS_ENABLED, workflowmgt.WFS_NOT_FOUND, workflowmgt.WFS_REJECTED,
		workflowmgt.WFS_TIMEOUT, workflowmgt.WFS_CANCELLED:
		// instruct user to create new workflow
		return errors.New(fmt.Sprintf(
			"Workflows are enabled for this component please create an workflow approval request at:\n%s",
			genComponentDeployUrl(org.Handle, projectId, cmpHandle),
		))
	case workflowmgt.WFS_DISABLED, workflowmgt.WFS_APPROVED:
		// proceed nothing to do
	}
	return
}

func generateByoiEndpoint(org api.Organization, project models.Project, remoteComponent models.Component, matchingAppEnv models.AppEnvVersions, byoiDeployOpts ByoiDeployConfigurations) error {
	if byoiDeployOpts.EndpointsFile != "" && strings.HasSuffix(strings.ToLower(remoteComponent.DisplayType), "service") {
		endpointsContent, err := os.ReadFile(byoiDeployOpts.EndpointsFile)
		if err != nil {
			return fmt.Errorf("failed to read endpoints file: %w", err)
		}

		createEndpointsReq := devops.CreateByoiEndpointsRequest{
			Main: base64.StdEncoding.EncodeToString([]byte(string(endpointsContent))),
		}

		if len(byoiDeployOpts.ApiSchemaFile) > 0 {
			for _, schemaFilePath := range byoiDeployOpts.ApiSchemaFile {
				schemaContent, err := os.ReadFile(schemaFilePath)
				if err != nil {
					return fmt.Errorf("failed to read schema file: %w", err)
				}
				createEndpointsReq.ApiSchemas = append(createEndpointsReq.ApiSchemas, devops.CreateApiSchemaRequest{
					Filename: filepath.Base(schemaFilePath),
					Content:  base64.StdEncoding.EncodeToString([]byte(string(schemaContent))),
				})
			}
		}

		createEndppointsSpinner := utils.CreateSpinner(i18n.T(" Creating endpoints for the component..."), "")
		createEndppointsSpinner.Start()
		err = auth.DevopsClient.CreateByoiEndpoints(org.ID, org.UUID, remoteComponent.Id, matchingAppEnv.ReleaseId, project.ID, createEndpointsReq)
		createEndppointsSpinner.Stop()
		if err != nil {
			return err
		}
	}
	return nil
}

func resolveImage(imageName string, imageHistoryItems []devops.ImageHistoryResponse) (item *devops.ImageHistoryResponse, err error) {
	opts := make([]huh.Option[*devops.ImageHistoryResponse], len(imageHistoryItems))

	uniqueImages := make(map[string]struct{})
	var filteredImageHistoryItems []devops.ImageHistoryResponse

	for _, v := range imageHistoryItems {
		if _, exists := uniqueImages[v.ImageNameWithTag]; !exists {
			uniqueImages[v.ImageNameWithTag] = struct{}{}
			filteredImageHistoryItems = append(filteredImageHistoryItems, v)
		}
	}
	imageHistoryItems = filteredImageHistoryItems

	if len(imageHistoryItems) == 0 {
		err = fmt.Errorf("no images to deploy")
		return
	}

	if len(imageHistoryItems) == 1 {
		item = &imageHistoryItems[0]
		return
	}

	for i, v := range imageHistoryItems {
		opts[i] = huh.NewOption(v.ImageNameWithTag, &v)
		if imageName == v.ImageNameWithTag {
			return &v, nil
		}
	}

	if item == nil {
		err = prompt.NewPromptSelectMessage(
			prompt.PromptSelectOpts[*devops.ImageHistoryResponse]{
				Options: opts,
				Message: i18n.T("Image: "),
			},
			&item,
		).Prompt()

		if err != nil {
			utils.PrintInfo(i18n.T("%s Failed to select image\n"), utils.CS.Red("!"))
			return nil, err
		}
	}

	return
}
