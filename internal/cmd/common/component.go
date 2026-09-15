package common

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

func AddComponentFlag(cmdFlags *pflag.FlagSet, bindTo *string) {
	cmdFlags.StringVarP(bindTo, "component", "c", "", i18n.T("component name, ID, or handle"))
}

func AddDeploymentTrackFlag(cmdFlags *pflag.FlagSet, bindTo *string) {
	cmdFlags.StringVarP(bindTo, "deployment-track", "d", "", i18n.T("name of the deployment track (e.g., main)"))
}

func ResolveTargetComponent(
	selectedOrg *api.Organization,
	projectId string,
	componentFlag string) (*models.Component, error) {

	getComponentSpinner := utils.CreateSpinner(i18n.T(" Fetching component information..."), "")
	getComponentSpinner.Start()
	remoteComponents, err := auth.ComponentClient.GetComponents(selectedOrg.Handle, selectedOrg.ID, projectId)
	getComponentSpinner.Stop()
	if err != nil {
		return nil, err
	}

	if len(remoteComponents) == 0 {
		return nil, fmt.Errorf("%w - no components found in project", api.ErrFailedToResolveComp)
	}

	return ResolveTargetComponentFromCompList(selectedOrg, projectId, remoteComponents, componentFlag)
}

func ResolveTargetComponentFromCompList(
	selectedOrg *api.Organization,
	projectId string,
	remoteComponents []models.Component,
	componentFlag string) (*models.Component, error) {

	var selectedComponent *models.Component

	if componentFlag != "" {
		for _, componentItem := range remoteComponents {
			if componentItem.Name == componentFlag ||
				componentItem.Handler == componentFlag ||
				componentItem.Id == componentFlag ||
				componentItem.DisplayName == componentFlag {

				selectedComponent = &componentItem
				break
			}
		}
	} else {
		selectedComponentName, err := promptToSelectComponent(remoteComponents)
		if err != nil {
			if errors.Is(err, internal.ErrNonInteractive) {
				return nil, utils.CreateNonInteractiveError("component selection", "component")
			}
			return nil, fmt.Errorf("%w - %s", api.ErrFailedToResolveComp, err)
		}
		for _, item := range remoteComponents {
			if item.Name == selectedComponentName {
				selectedComponent = &item
				break
			}
		}
	}

	if selectedComponent == nil {
		return nil, fmt.Errorf("%w - no matching component found", api.ErrFailedToResolveComp)
	}

	return selectedComponent, nil
}

func promptToSelectComponent(components []models.Component) (string, error) {
	var selectedComponentName string
	var componentNames []string

	for _, item := range components {
		componentNames = append(componentNames, item.Name)
	}

	err := prompt.NewPromptSelectMessage[string](
		prompt.PromptSelectOpts[string]{
			Message: "Component:",
			Values:  componentNames,
		},
		&selectedComponentName,
	).Prompt()

	if err != nil {
		return "", err
	}

	return selectedComponentName, nil
}

func GetLatestCommit(componentId string, branch string, orgId string) (commitHash string, err error) {
	getCommitSpinner := utils.CreateSpinner(i18n.T(" Fetching commit history..."), "")
	getCommitSpinner.Start()
	pushedCommitHistory, err := auth.GitClient.GetCommitHistory(componentId, branch, orgId)
	getCommitSpinner.Stop()
	if err != nil {
		return "", err
	}

	// find latest pushed commit
	var latestPushedCommitHash string
	for _, commit := range pushedCommitHistory {
		if commit.IsLatest {
			latestPushedCommitHash = commit.Sha
			utils.PrintInfo(i18n.T("Using the latest commit: %s %s\n"), commit.Message, utils.CS.Gray(latestPushedCommitHash))
			break
		}
	}

	return latestPushedCommitHash, nil
}

func ResolveDeploymentTrack(deploymentTracks []models.DeploymentTrack, deploymentTrackFlag string) (*models.DeploymentTrack, error) {
	var selectedTrack *models.DeploymentTrack

	if deploymentTrackFlag != "" {
		for _, item := range deploymentTracks {
			if item.Branch == deploymentTrackFlag || item.ApiVersion == deploymentTrackFlag {
				selectedTrack = &item
				break
			}
		}
	} else if len(deploymentTracks) == 1 {
		selectedTrack = &deploymentTracks[0]
		if selectedTrack.Branch == "" {
			selectedTrack.Branch = selectedTrack.ApiVersion
		}
	} else {
		var selectedTrackName string
		var trackNames []string

		for _, item := range deploymentTracks {
			if item.Branch == "" {
				item.Branch = item.ApiVersion
			}
			trackNames = append(trackNames, item.Branch)
		}

		err := prompt.NewPromptSelectMessage(
			prompt.PromptSelectOpts[string]{
				Message: i18n.T("Deployment track:"),
				Values:  trackNames,
			},
			&selectedTrackName,
		).Prompt()

		if err != nil {
			if errors.Is(err, internal.ErrNonInteractive) {
				return nil, utils.CreateNonInteractiveError("deployment track selection", "deployment-track")
			}
			return nil, err
		}

		for _, item := range deploymentTracks {
			if item.Branch == selectedTrackName {
				selectedTrack = &item
				break
			}
		}
	}

	if selectedTrack == nil {
		return nil, errors.New(i18n.T(" invalid deployment track"))
	}

	return selectedTrack, nil
}

func GetLatestDeployementTrack(deploymentTracks []models.DeploymentTrack) (*models.DeploymentTrack, error) {
	for _, item := range deploymentTracks {
		if item.Latest {
			return &item, nil
		}
	}
	if len(deploymentTracks) > 0 {
		return &deploymentTracks[0], nil
	}
	return nil, errors.New(i18n.T(" no deployment track found for the component"))
}

func GetComponentDeploymentStatusByVersion(
	componentId string,
	versionId string,
	orgId string) (componentDeploymentData []component.ComponentDeploymentStatusForVersion, err error) {

	componentVersionSpinner := utils.CreateSpinner(i18n.T(" Fetching component status for the selected deployment track..."), "")
	componentVersionSpinner.Start()
	componentDeploymentData, err = auth.ComponentClient.GetDeploymentStatusByVersion(componentId, versionId, orgId)
	componentVersionSpinner.Stop()
	if err != nil {
		return nil, fmt.Errorf(i18n.T("failed to fetch deployment status details: %w"), err)
	}
	return componentDeploymentData, err
}

func GetComponentBuildLogs(
	orgHandle string,
	orgId string,
	projectId string,
	componentId string,
	runId int) (*component.DeploymentBuildStatusResponse, error) {

	buildLogsSpinner := utils.CreateSpinner(i18n.T(" Fetching component build logs..."), "")
	buildLogsSpinner.Start()
	buildLogsDataRes, err := auth.ComponentClient.GetComponentBuildStatus(orgHandle, projectId, componentId, runId, orgId)
	buildLogsSpinner.Stop()
	if err != nil {
		return nil, fmt.Errorf(i18n.T("failed to fetch build logs: %w"), err)
	}
	return buildLogsDataRes, err
}

func GetBuildLogForPhase(
	orgId string,
	componentId string,
	runId int,
	logPhase string,
) (*component.ProjectBuildLogsData, error) {
	buildLogsSpinner := utils.CreateSpinner(i18n.T(" Fetching component build logs..."), "")
	buildLogsSpinner.Start()
	buildLogsDataRes, err := auth.ComponentClient.GetBuildLogsForType(orgId, componentId, runId, logPhase)
	buildLogsSpinner.Stop()
	if err != nil {
		return nil, fmt.Errorf(i18n.T("failed to fetch build logs: %w"), err)
	}
	return &buildLogsDataRes, err
}

func GetMiComponentBuildLogs(orgId string, componentId string, runId int) (string, error) {
	buildLogsSpinner := utils.CreateSpinner(i18n.T(" Fetching component MI build logs..."), "")
	buildLogsSpinner.Start()
	buildLogsDataRes, err := auth.ComponentClient.GetBuildLogsForType(orgId, componentId, runId, "integrationProjectBuild")
	buildLogsSpinner.Stop()
	if err != nil {
		return "", fmt.Errorf(i18n.T("failed to fetch build logs: %w"), err)
	}
	return buildLogsDataRes.IntegrationProjectBuild, err
}

func GetComponentWithRepoData(orgId string, componentHandle string, projectId string) (models.Component, error) {
	componentSpinner := utils.CreateSpinner(i18n.T(" Fetching additional component details..."), "")
	componentSpinner.Start()

	remoteComWithRepoData, err := auth.ComponentClient.GetComponentInfo(
		orgId,
		componentHandle,
		projectId,
	)
	componentSpinner.Stop()
	if err != nil {
		return models.Component{}, err
	}

	return *remoteComWithRepoData, nil
}

func GetReleaseEnvForDeploymentTrack(remoteCompWithRepoData models.Component, deploymentTrackId string, envId string) (models.AppEnvVersions, error) {
	var matchingApiVersion *models.ApiVersion
	for _, apiVersion := range remoteCompWithRepoData.ApiVersions {
		if apiVersion.Id == deploymentTrackId {
			matchingApiVersion = &apiVersion
			break
		}
	}

	if matchingApiVersion == nil {
		return models.AppEnvVersions{}, errors.New(i18n.T("no matching ApiVersion found"))
	}

	var matchingAppEnvVersion *models.AppEnvVersions

	for _, appEnvVersion := range matchingApiVersion.AppEnvVersions {
		if appEnvVersion.EnvironmentId == envId {
			matchingAppEnvVersion = &appEnvVersion
			break
		}
	}

	if matchingAppEnvVersion == nil {
		return models.AppEnvVersions{}, errors.New(i18n.T("no matching AppEnvVersion found"))
	}

	return *matchingAppEnvVersion, nil
}

func GetComponentsForProject(org *api.Organization, project *models.Project) (comps []models.Component, err error) {
	compSpinner := utils.CreateSpinner("Fetching components", "")
	compSpinner.Start()
	comps, err = auth.ComponentClient.GetComponents(org.Handle, org.ID, project.ID)
	compSpinner.Stop()

	return
}

func GetComponentsWithSystemCompsForProject(org *api.Organization, project *models.Project) (comps []models.Component, err error) {
	compSpinner := utils.CreateSpinner("Fetching components", "")
	compSpinner.Start()
	comps, err = auth.ComponentClient.GetAllComponents(org.Handle, org.ID, project.ID, true)
	compSpinner.Stop()

	return
}

// check the project directory structure for .wso2/component-config.yaml file before building and deploying
func IsSafeToBuildAndDeploy(
	proj *models.Project,
	comp *models.Component,
	org *api.Organization,
	dt *models.DeploymentTrack,
) (safe bool, err error) {

	compInfo, err := GetComponentWithRepoData(org.ID, comp.Handler, proj.ID)

	if err != nil {
		return
	}

	subPath := ""

	if compInfo.Repository.ByocBuildConfig != nil {
		subPath = compInfo.Repository.ByocBuildConfig.DockerContext
	} else if compInfo.Repository.BuildPackConfig != nil && len(compInfo.Repository.BuildPackConfig) > 0 {
		for _, bp := range compInfo.Repository.BuildPackConfig {
			if dt.Id == bp.VersionId {
				subPath = bp.BuildContext
				break
			}
		}
	} else if compInfo.Repository.AppSubPath != "" {
		subPath = compInfo.Repository.AppSubPath
	} else {
		// if none of the above are set, assume the root of the repo
		err = errors.New(i18n.T("unable to determine the build context"))
		return
	}

	gitProvider, gitOrg, gitRepo, gitBranch, gitServerUrl := "", "", "", "", ""
	if proj.Repository != "" && proj.Branch != "" && proj.GitOrganization != "" {
		// if mono-repo
		gitProvider, gitOrg, gitRepo, gitBranch = proj.GitProvider, proj.GitOrganization, proj.Repository, proj.Branch
	} else {
		// if multi-repo
		gitProvider, gitOrg, gitRepo, gitBranch, gitServerUrl = compInfo.Repository.GitProvider, compInfo.Repository.OrganizationApp, compInfo.Repository.NameApp, compInfo.Repository.BranchApp, compInfo.Repository.ServerUrl
		if compInfo.Repository.BitbucketServerUrl != "" {
			gitServerUrl = compInfo.Repository.BitbucketServerUrl
		}
	}

	usr, err := auth.GetCurrentUser()
	if err != nil {
		return
	}

	repoUrl := GenRepoUrl(gitProvider, gitOrg, gitRepo, gitServerUrl)

	if !strings.Contains(repoUrl, "github.com/") {
		return true, api.FuncionalityNotSupported
	}

	if org.Owner.IDPId == usr.IDPId {
		_, _, err = HandleGitRepoAuthorization(GenRepoUrl(gitProvider, gitOrg, gitRepo, gitServerUrl), org.ID, org.UUID, "")
		if err != nil {
			return
		}
	}

	isPublicRepoResp, _ := auth.GitClient.IsPublicRepo(gitOrg, gitRepo, org.ID)

	repoStructure, err := GetRepoStructure(gitOrg, gitRepo, gitBranch, isPublicRepoResp.IsPublicRepo, org.ID)

	if err != nil {
		// skip checking directory has component yaml if fail to find the repo details
		// todo:instead of using SubPathHasEndpointsYaml, we should use componentDeployment gql call
		safe = true
		err = nil
		return
	}

	safe = SubPathHasEndpointsYaml(subPath, &repoStructure)
	return
}

func SubPathHasEndpointsYaml(subPath string, pathEntries *[]component.PathEntry) (has bool) {
	if strings.HasPrefix(subPath, "./") {
		subPath = strings.TrimPrefix(subPath, "./")
	}

	if strings.HasPrefix(subPath, "/") {
		subPath = strings.TrimPrefix(subPath, "/")
	}

	for _, entry := range *pathEntries {
		if entry.Type == component.PathBlobType &&
			(filepath.Join(entry.Path) == filepath.Join(subPath, ".wso2/endpoints.yaml") ||
				filepath.Join(entry.Path) == filepath.Join(subPath, ".wso2/component-config.yaml") ||
				filepath.Join(entry.Path) == filepath.Join(subPath, ".wso2/component.yaml")) {
			has = true
			break
		}

		has = SubPathHasEndpointsYaml(subPath, &entry.Children)

		if has {
			return
		}
	}

	return
}

func GetComponentDeplayment(
	orgId, orgHandler, orgUUID, componentId, versionId, envId string) (*component.ComponentDeployment, error) {

	getCompDeploymentSpinner := utils.CreateSpinner(i18n.T(" Fetching component deployment information..."), "")
	getCompDeploymentSpinner.Start()
	deploymentInfo, err := auth.ComponentClient.GetComponentDeployment(
		orgHandler,
		orgUUID,
		orgId,
		componentId,
		versionId,
		envId,
	)
	getCompDeploymentSpinner.Stop()

	return deploymentInfo, err
}
