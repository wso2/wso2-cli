package component

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/clirpc/server"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/cmd/component/create"
	createComponent "github.com/wso2/integration-platform-tools/internal/cmd/component/create"
	"github.com/wso2/integration-platform-tools/internal/cmd/component/link"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/devops"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/api/platformgit"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
	"github.com/wso2/integration-platform-tools/pkg/util/constants"
	"github.com/wso2/integration-platform-tools/pkg/util/repo"
)

type ComponentKindExtended struct {
	ApiVersion       string                          `json:"apiVersion"`
	Kind             string                          `json:"kind"`
	Metadata         component.ComponentKindMetadata `json:"metadata"`
	Spec             component.ComponentKindSpec     `json:"spec"`
	DeploymentTracks []models.DeploymentTrack        `json:"deploymentTracks"`
	ApiVersions      []models.ApiVersion             `json:"apiVersions"`
	CreatedAt        string                          `json:"createdAt"`
}

func CreateComponentLink(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		ComponentDir    string `json:"componentDir"`
		OrgHandle       string `json:"orgHandle"`
		ProjectHandle   string `json:"projectHandle"`
		ComponentHandle string `json:"componentHandle"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	err := link.CreateComponentLink(request.ComponentDir, request.ProjectHandle, request.OrgHandle, request.ComponentHandle)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct{}{}), nil
}

func GetComponents(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId         string `json:"orgId"`
		OrgHandle     string `json:"orgHandle"`
		ProjectHandle string `json:"projectHandle"`
		ProjectId     string `json:"projectId"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	componentList, err := auth.ComponentClient.GetComponents(request.OrgHandle, request.OrgId, request.ProjectId)
	if err != nil {
		return server.Result{}, err
	}

	var compKindList []ComponentKindExtended
	for _, compItem := range componentList {
		newCompItem := ComponentKindExtended{
			Metadata: component.ComponentKindMetadata{
				Name:        compItem.Name,
				DisplayName: compItem.DisplayName,
				ID:          compItem.Id,
				Handler:     compItem.Handler,
			},
			Spec:             component.ComponentKindSpec{Type: compItem.DisplayType, SubType: compItem.ComponentSubType},
			DeploymentTracks: compItem.DeploymentTracks,
		}
		compKindList = append(compKindList, newCompItem)
	}

	// TODO: batch these into groups of 5

	// Create a channel to collect results
	resultChan := make(chan ComponentKindExtended, len(compKindList))

	// Use WaitGroup to ensure all goroutines complete
	var wg sync.WaitGroup

	for _, compItem := range compKindList {
		wg.Add(1)
		go func(item ComponentKindExtended) {
			defer wg.Done()

			// Enrich the component
			enrichComponentKind(&item, request.OrgId, request.ProjectId)

			// Send the result back to the channel
			resultChan <- item
		}(compItem)
	}

	// Close the channel once all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results from the channel
	var enrichedCompKindList []ComponentKindExtended
	for enrichedCompKind := range resultChan {
		enrichedCompKindList = append(enrichedCompKindList, enrichedCompKind)
	}

	// Sort based on createdAt
	sort.Slice(enrichedCompKindList, func(i, j int) bool {
		// Parse the CreatedAt field of both components
		timeI, errI := time.Parse(time.RFC3339, enrichedCompKindList[i].CreatedAt)
		timeJ, errJ := time.Parse(time.RFC3339, enrichedCompKindList[j].CreatedAt)
		// If there's an error parsing, default to placing unparseable entries at the end
		if errI != nil || errJ != nil {
			return false
		}
		// Compare time values, most recent first (descending order)
		return timeI.After(timeJ)
	})

	return server.CreateResult(struct {
		Components []ComponentKindExtended `json:"components"`
	}{
		Components: enrichedCompKindList,
	}), nil
}

func GetComponentItem(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId         string `json:"orgId"`
		ComponentName string `json:"componentName"`
		ProjectHandle string `json:"projectHandle"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	componentItem, err := auth.ComponentClient.GetComponentDeclarative(request.ProjectHandle, request.ComponentName, request.OrgId)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		Component component.ComponentKind `json:"component"`
	}{
		Component: componentItem,
	}), nil
}

func GetBuildPacks(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgUuid       string `json:"orgUuid"`
		OrgId         string `json:"orgId"`
		ComponentType string `json:"componentType"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	buildpacks, err := auth.DevopsClient.GetBuildPackOptions(request.OrgUuid, request.OrgId, request.ComponentType)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		BuildPacks []devops.BuildPack `json:"buildPacks"`
	}{
		BuildPacks: buildpacks,
	}), nil
}

func CreateComponent(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId            string `json:"orgId"`
		OrgUUID          string `json:"orgUUID"`
		ProjectId        string `json:"projectId"`
		ProjectHandle    string `json:"projectHandle"`
		Name             string `json:"name"`
		DisplayName      string `json:"displayName"`
		ComponentSubType string `json:"componentSubType"`
		Type             string `json:"type"`
		BuildPackLang    string `json:"buildPackLang"`
		ComponentDir     string `json:"componentDir"`
		RepoUrl          string `json:"repoUrl"`
		GitProvider      string `json:"gitProvider"`
		GitCredRef       string `json:"gitCredRef"`
		Branch           string `json:"branch"`
		Port             int    `json:"port"`
		OriginCloud      string `json:"originCloud"`
	}

	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	_, relativePath, err := common.GetComponentSubPath(request.ComponentDir)
	if err != nil {
		return server.Result{}, err
	}

	if runtime.GOOS == "windows" {
		relativePath = strings.ReplaceAll(relativePath, "\\", "/")
	}

	gitOrgName, gitRepoName, err := repo.ParseGitURL(request.RepoUrl)
	if err != nil {
		return server.Result{}, err
	}

	buildPackConfig := make(map[string]string)
	buildPackConfig[createComponent.PORT] = strconv.Itoa(request.Port)

	repoProvider := create.GIT_HUB
	if request.GitProvider == create.BIT_BUCKET.String() {
		repoProvider = create.BIT_BUCKET
	} else if request.GitProvider == create.BIT_BUCKET_SERVER.String() {
		repoProvider = create.BIT_BUCKET_SERVER
	} else if request.GitProvider == create.GIT_LAB_SERVER.String() {
		repoProvider = create.GIT_LAB_SERVER
	}

	createCompParam := createComponent.CreateComponentParams{
		ComponentName:    request.Name,
		DisplayName:      request.DisplayName,
		ComponentType:    request.Type,
		ComponentSubType: request.ComponentSubType,
		RepoProvider:     repoProvider,
		Repo:             gitRepoName,
		RepoOrg:          gitOrgName,
		RepoBranch:       request.Branch,
		Subpath:          relativePath,
		BuildPack:        request.BuildPackLang,
		BuildPackConfigs: buildPackConfig,
		GitCredRef:       request.GitCredRef,
		OriginCloud:      request.OriginCloud,
	}

	componentReqData, err := createComponent.GetComponentKindForCreate(&createCompParam, request.ProjectHandle, request.OrgId, request.OrgUUID, request.RepoUrl)
	if err != nil {
		return server.Result{}, err
	}

	createdComponent, err := auth.ComponentClient.CreateNewComponent(request.OrgId, request.ProjectHandle, *componentReqData)
	if err != nil {
		return server.Result{}, err
	}

	newCompItem := ComponentKindExtended{
		Metadata: component.ComponentKindMetadata{
			Name:        createdComponent.Metadata.Name,
			DisplayName: createdComponent.Metadata.DisplayName,
			ID:          createdComponent.Metadata.ID,
			Handler:     request.Name,
		},
		Spec: component.ComponentKindSpec{Type: createdComponent.Spec.Type},
	}

	compoInfo, _ := enrichComponentKind(&newCompItem, request.OrgId, request.ProjectId)
	newCompItem.DeploymentTracks = compoInfo.DeploymentTracks

	return server.CreateResult(struct {
		Component ComponentKindExtended `json:"component"`
	}{
		Component: newCompItem,
	}), nil
}

func DeleteComponent(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgHandler  string `json:"orgHandler"`
		OrgId       string `json:"orgId"`
		ProjectId   string `json:"projectId"`
		ComponentId string `json:"componentId"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	resp, err := auth.ComponentClient.DeleteComponent(request.OrgId, request.OrgHandler, request.ComponentId, request.ProjectId)
	if err != nil {
		return server.Result{}, err
	}

	if resp.Status == "error" {
		return server.CreateResult(struct{}{}), fmt.Errorf("%s", resp.Message)
	}

	return server.CreateResult(struct{}{}), nil
}

func GetDeploymentTracks(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgHandler      string `json:"orgHandler"`
		OrgId           string `json:"orgId"`
		ProjectId       string `json:"projectId"`
		ComponentHandle string `json:"componentId"`
	}

	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	component, err := auth.ComponentClient.GetComponentInfo(request.OrgId, request.OrgHandler, request.ComponentHandle)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		DeploymentTracks []models.DeploymentTrack `json:"deploymentTracks"`
	}{
		DeploymentTracks: component.DeploymentTracks,
	}), nil
}

func GetCommitHistory(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgHandler  string `json:"orgHandler"`
		OrgId       string `json:"orgId"`
		Branch      string `json:"branch"`
		ComponentId string `json:"componentId"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	pushedCommitHistory, err := auth.GitClient.GetCommitHistory(request.ComponentId, request.Branch, request.OrgId)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		Commits []platformgit.CommitHistory `json:"commits"`
	}{
		Commits: pushedCommitHistory,
	}), nil
}

func GetComponentEndpoints(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgHandler        string `json:"orgHandler"`
		OrgId             string `json:"orgId"`
		ComponentId       string `json:"componentId"`
		DeploymentTrackId string `json:"deploymentTrackId"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	endpoints, err := auth.ProjectClient.GetComponentEndpoints(request.ComponentId, request.DeploymentTrackId, request.OrgId)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		Endpoints []project.Endpoint `json:"endpoints"`
	}{
		Endpoints: *endpoints,
	}), nil
}

func GetComponentDeploymentStatus(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgHandler        string `json:"orgHandler"`
		OrgId             string `json:"orgId"`
		OrgUuid           string `json:"orgUuid"`
		ComponentId       string `json:"componentId"`
		DeploymentTrackId string `json:"deploymentTrackId"`
		EnvId             string `json:"envId"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	componentDeployment, err := auth.ComponentClient.GetComponentDeployment(
		request.OrgHandler,
		request.OrgUuid,
		request.OrgId,
		request.ComponentId,
		request.DeploymentTrackId,
		request.EnvId)

	type Response struct {
		Deployment component.ComponentDeployment `json:"deployment"`
	}

	if err != nil {
		return server.Result{}, nil
	}

	return server.CreateResult(Response{
		Deployment: *componentDeployment,
	}), nil
}

func CreateComponentConfig(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		ComponentDir string                       `json:"componentDir"`
		Type         string                       `json:"type"`
		Inbound      models.ComponentYamlEndpoint `json:"inbound"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	inboundConfigs := []models.ComponentYamlEndpoint{}
	if request.Type == component.ComponentTypeService {
		inboundConfigs = []models.ComponentYamlEndpoint{request.Inbound}
	}

	configPath, err := create.CreateComponentConfigFile(request.ComponentDir, request.Type, inboundConfigs, nil)
	if err != nil {
		return server.Result{}, nil
	}

	return server.CreateResult(struct {
		ConfigPath string `json:"configPath"`
	}{
		ConfigPath: configPath,
	}), nil
}

func updateCodeServer(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	userInfo, err := auth.GetCurrentUser()
	if err != nil {
		return server.Result{}, err
	}

	var request struct {
		OrgId            string `json:"orgId"`
		OrgUuid          string `json:"orgUuid"`
		OrgHandle        string `json:"orgHandle"`
		ProjectId        string `json:"projectId"`
		ComponentId      string `json:"componentId"`
		SourceCommitHash string `json:"sourceCommitHash"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	registryItem, err := getContainerRegistry(request.OrgId, request.OrgUuid)
	if err != nil {
		return server.Result{}, err
	}

	samples, err := auth.DevopsClient.GetSamples(request.OrgId, request.OrgUuid, request.ProjectId)
	if err != nil {
		return server.Result{}, err
	}

	var codeServerImage *devops.ContainerImage

	for _, item := range samples {
		if item.Name == "Code Server" {
			codeServerImage = &item
			break
		}
	}

	if codeServerImage == nil {
		return server.Result{}, fmt.Errorf("failed to find code server from samples")
	}

	err = auth.GraphQLClient.UpdateCodeServer(request.OrgId, userInfo.IDPId, request.OrgUuid, request.ProjectId, request.ComponentId, request.OrgHandle, codeServerImage.ImageUrl, registryItem.Id, request.SourceCommitHash)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct{}{}), nil
}

func getContainerRegistry(orgId, orgUuid string) (*devops.ContainerRegistry, error) {
	containerRegs, err := auth.DevopsClient.GetContainerRegistries(orgId, orgUuid)
	if err != nil {
		return nil, nil
	}

	var registryItem *devops.ContainerRegistry
	for _, item := range containerRegs {
		if item.Host == constants.WSO2IP_AZURECR {
			registryItem = &item
			break
		}
	}

	if registryItem == nil {
		err := auth.DevopsClient.RegisterNewContainerRegistry(orgId, orgUuid)
		if err != nil {
			return nil, nil
		}

		containerRegs, err := auth.DevopsClient.GetContainerRegistries(orgId, orgUuid)
		if err != nil {
			return nil, nil
		}

		for _, item := range containerRegs {
			if item.Host == constants.WSO2IP_AZURECR {
				registryItem = &item
				break
			}
		}

		if registryItem == nil {
			return nil, fmt.Errorf("failed to find registry")
		}
	}
	return registryItem, nil
}

func enrichComponentKind(compItem *ComponentKindExtended, orgId string, projectId string) (compInfo *models.Component, err error) {
	if compItem.Spec.Type == component.DisplayTypeGitProxy {
		// Fetching only repositories since API returning a 500 when other query fields are also included
		compInfo, err = auth.ComponentClient.GetComponentRepoInfo(orgId, compItem.Metadata.Handler, projectId)
		if err != nil {
			return
		}
	} else {
		compInfo, err = auth.ComponentClient.GetComponentInfo(orgId, compItem.Metadata.Handler, projectId)
		if err != nil {
			return
		}
	}

	if len(compInfo.DeploymentTracks) > 0 {
		compItem.DeploymentTracks = compInfo.DeploymentTracks
	}

	if len(compInfo.ApiVersions) > 0 {
		compItem.ApiVersions = compInfo.ApiVersions
	}

	compItem.CreatedAt = compInfo.CreatedAt

	subPath := ""
	if strings.HasPrefix(compInfo.DisplayType, "mi") {
		subPath = compInfo.Repository.AppSubPath
	} else if compInfo.Repository.ByocBuildConfig != nil {
		subPath = compInfo.Repository.ByocBuildConfig.DockerContext
	} else if compInfo.Repository.BuildPackConfig != nil && len(compInfo.Repository.BuildPackConfig) > 0 {
		subPath = compInfo.Repository.BuildPackConfig[0].BuildContext
	} else if compInfo.Repository.AppSubPath != "" {
		subPath = compInfo.Repository.AppSubPath
	}

	if compInfo.Repository.GitProvider == create.GIT_HUB.String() {
		compItem.Spec.Source = component.ComponentKindSource{
			Github: &component.GitProvider{
				Repository: fmt.Sprintf(`https://github.com/%s/%s`, compInfo.Repository.OrganizationApp, compInfo.Repository.NameApp),
				Path:       subPath,
			},
		}
	} else if compInfo.Repository.GitProvider == create.BIT_BUCKET.String() {
		compItem.Spec.Source = component.ComponentKindSource{
			Bitbucket: &component.GitProvider{
				Repository: fmt.Sprintf(`https://bitbucket.org/%s/%s`, compInfo.Repository.OrganizationApp, compInfo.Repository.NameApp),
				Path:       subPath,
			},
			SecretRef: compInfo.Repository.RepoCredRef,
		}
	} else if compInfo.Repository.GitProvider == "bitbucket-server" {
		compItem.Spec.Source = component.ComponentKindSource{
			Bitbucket: &component.GitProvider{
				Repository: fmt.Sprintf(`%s/projects/%s/%s`, compInfo.Repository.BitbucketServerUrl, compInfo.Repository.OrganizationApp, compInfo.Repository.NameApp),
				Path:       subPath,
			},
			SecretRef: compInfo.Repository.RepoCredRef,
		}
	} else if compInfo.Repository.GitProvider == create.GIT_LAB_SERVER.String() {
		compItem.Spec.Source = component.ComponentKindSource{
			Gitlab: &component.GitProvider{
				Repository: fmt.Sprintf(`%s/%s/%s`, compInfo.Repository.ServerUrl, compInfo.Repository.OrganizationApp, compInfo.Repository.NameApp),
				Path:       subPath,
			},
			SecretRef: compInfo.Repository.RepoCredRef,
		}
	}

	if compInfo.DisplayType == component.DisplayTypeByocWebAppDockerLess {
		compItem.Spec.Build = component.ComponentKindSpecBuild{
			Webapp: &component.ComponentKindBuildWebapp{
				BuildCommand: compInfo.Repository.ByocWebAppBuildConfig.BuildCommand,
				NodeVersion:  compInfo.Repository.ByocWebAppBuildConfig.PackageManagerVersion,
				OutputDir:    compInfo.Repository.ByocWebAppBuildConfig.OutputDirectory,
				Type:         compInfo.Repository.ByocWebAppBuildConfig.WebAppType,
			},
		}
	} else if strings.HasPrefix(compInfo.DisplayType, "byoc") {
		compItem.Spec.Build = component.ComponentKindSpecBuild{
			Docker: &component.ComponentKindBuildDocker{
				DockerFilePath:    compInfo.Repository.ByocBuildConfig.DockerfilePath,
				DockerContextPath: compInfo.Repository.ByocBuildConfig.DockerContext,
			},
		}
	} else if strings.HasPrefix(compInfo.DisplayType, "buildpack") {
		if len(compInfo.Repository.BuildPackConfig) > 0 {
			compItem.Spec.Build = component.ComponentKindSpecBuild{
				Buildpack: &component.ComponentKindBuildBuildpack{
					Language: compInfo.Repository.BuildPackConfig[0].Buildpack.Language,
					Version:  compInfo.Repository.BuildPackConfig[0].LanguageVersion,
				},
			}
		}
	} else if strings.HasPrefix(compInfo.DisplayType, "mi") {
		compItem.Spec.Build = component.ComponentKindSpecBuild{
			Buildpack: &component.ComponentKindBuildBuildpack{Language: "WSO2 MI", Version: "Auto Detected"},
		}
	}
	return
}
