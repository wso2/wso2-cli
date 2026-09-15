package build

import (
	"encoding/json"
	"strconv"
	"sync"

	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/clirpc/server"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/logs"
	"github.com/wso2/integration-platform-tools/pkg/api/platformgit"

	deploymentbuild "github.com/wso2/integration-platform-tools/pkg/api/deploymentBuild"
)

func GetBuildList(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	type BuildResponse struct {
		Builds []deploymentbuild.BuildKind `json:"builds"`
	}

	var request struct {
		OrgId         string `json:"orgId"`
		ComponentId   string `json:"componentId"`
		ComponentName string `json:"componentName"`
		DisplayType   string `json:"displayType"`
		Branch        string `json:"branch"`
		ApiVersionId  string `json:"apiVersionId"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	builds, err := getBuildKinds(request.OrgId, request.ComponentId, request.ComponentName, request.Branch, request.DisplayType, request.ApiVersionId)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(BuildResponse{Builds: builds}), nil
}

func CreateBuild(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId             string `json:"orgId"`
		ComponentName     string `json:"componentName"`
		DisplayType       string `json:"displayType"`
		ProjectHandle     string `json:"projectHandle"`
		DeploymentTrackId string `json:"deploymentTrackId"`
		CommitHash        string `json:"commitHash"`
		GitRepoUrl        string `json:"gitRepoUrl"`
		GitBranch         string `json:"gitBranch"`
		SubPath           string `json:"subPath"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	buildResponse, err := auth.DeploymentBuildClient.CreateDeploymentBuilds(request.OrgId, request.ComponentName, request.ProjectHandle, request.DeploymentTrackId, request.CommitHash)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		Build deploymentbuild.BuildKind `json:"build"`
	}{
		Build: buildResponse,
	}), nil
}

func GetBuildLogs(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgHandler        string `json:"orgHandler"`
		OrgId             string `json:"orgId"`
		OrgUuid           string `json:"orgUuid"`
		ProjectId         string `json:"projectId"`
		ComponentId       string `json:"componentId"`
		DeploymentTrackId string `json:"deploymentTrackId"`
		DisplayType       string `json:"displayType"`
		BuildId           int    `json:"buildId"`
		BuildRef          string `json:"buildRef"`
		ClusterId         string `json:"clusterId"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	isArgo := request.ClusterId != ""

	type BuildLogsResponse struct {
		Data component.DeploymentBuildStatusResponse `json:"data"`
	}

	if isArgo {
		dataPlanes, cloudDataPlanes, err := auth.DevopsClient.GetAllDataPlaneClusters(request.OrgId, request.OrgUuid)
		if err != nil {
			return server.Result{}, err
		}

		dataPlaneHost, err := common.GetDataPlaneHost(cloudDataPlanes, dataPlanes, request.ClusterId)
		if err != nil {
			return server.Result{}, err
		}

		reqBody := logs.GetBuildLogsReqBody{
			ComponentId:       request.ComponentId,
			DeploymentTrackId: request.DeploymentTrackId,
			WorkflowName:      request.BuildRef,
		}

		buildLogsDataRes, err := auth.LogsClient.GetBuildLogs(reqBody, request.OrgId, dataPlaneHost)
		if err != nil {
			return server.Result{}, err
		}
		return server.CreateResult(BuildLogsResponse{Data: *buildLogsDataRes}), nil
	} else {
		buildLogsDataRes, err := auth.ComponentClient.GetComponentBuildStatus(request.OrgHandler, request.ProjectId, request.ComponentId, request.BuildId, request.OrgId)
		if err != nil {
			return server.Result{}, err
		}
		return server.CreateResult(BuildLogsResponse{Data: *buildLogsDataRes}), nil
	}
}

func GetBuildLogsForType(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId       string `json:"orgId"`
		ComponentId string `json:"componentId"`
		LogType     string `json:"logType"`
		BuildId     int    `json:"buildId"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	buildLogsDataRes, err := auth.ComponentClient.GetBuildLogsForType(request.OrgId, request.ComponentId, request.BuildId, request.LogType)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		Data component.ProjectBuildLogsData `json:"data"`
	}{
		Data: buildLogsDataRes,
	}), nil
}

func GetAutoBuildStatus(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId       string `json:"orgId"`
		ComponentId string `json:"componentId"`
		VersionId   string `json:"versionId"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	resp, err := auth.ComponentClient.GetAutoBuildStatus(
		request.ComponentId,
		request.VersionId,
		request.OrgId,
	)

	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(resp.Data), nil
}

func EnableAutoBuild(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId       string `json:"orgId"`
		ComponentId string `json:"componentId"`
		VersionId   string `json:"versionId"`
		EnvId       string `json:"envId"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	resp, err := auth.ComponentClient.HandleEnableAutoBuild(
		request.ComponentId,
		request.VersionId,
		request.EnvId,
		request.OrgId,
	)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		Success bool `json:"success"`
	}{
		Success: resp.Success,
	}), nil
}

func DisableAutoBuild(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId       string `json:"orgId"`
		ComponentId string `json:"componentId"`
		VersionId   string `json:"versionId"`
		EnvId       string `json:"envId"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	resp, err := auth.ComponentClient.HandleDisableAutoBuild(
		request.ComponentId,
		request.VersionId,
		request.EnvId,
		request.OrgId,
	)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		Success bool `json:"success"`
	}{
		Success: resp.Success,
	}), nil
}

func getBuildKinds(orgId, componentId, componentName, branch, displayType, versionId string) ([]deploymentbuild.BuildKind, error) {
	var wg sync.WaitGroup
	var deploymentTrackBuilds []component.ComponentDeploymentStatusForVersion
	var deploymentTrackImages []component.ComponentDeploymentTrackImages
	var commitHistory []platformgit.CommitHistory
	var errBuilds, errImages, errCommits error

	// Channels to handle concurrent requests
	buildsChan := make(chan []deploymentbuild.BuildKind, 1)
	imagesChan := make(chan []component.ComponentDeploymentTrackImages, 1)
	commitsChan := make(chan []platformgit.CommitHistory, 1)

	// Goroutine to fetch deploymentTrackBuilds
	wg.Add(1)
	go func() {
		defer wg.Done() // Ensure wg.Done is called no matter what happens
		deploymentTrackBuilds, errBuilds = auth.ComponentClient.GetDeploymentStatusByVersion(componentId, versionId, orgId)
		if errBuilds != nil {
			buildsChan <- nil
			return
		}

		var tempBuilds []deploymentbuild.BuildKind
		// Convert ComponentDeploymentStatusForVersion to BuildKind
		for _, deploymentTrackBuild := range deploymentTrackBuilds {
			buildSpec := deploymentbuild.BuildSpec{Revision: deploymentTrackBuild.SourceCommitId}
			status := deploymentbuild.BuildStatus{
				RunID:       deploymentTrackBuild.Id,
				Conclusion:  deploymentTrackBuild.Conclusion,
				Status:      deploymentTrackBuild.Status,
				StartedAt:   deploymentTrackBuild.StartedAt,
				CompletedAt: deploymentTrackBuild.CompletedAt,
				ClusterId:   deploymentTrackBuild.ClusterId,
				BuildRef:    deploymentTrackBuild.BuildRef,
			}
			tempBuilds = append(tempBuilds, deploymentbuild.BuildKind{
				Kind:     "Build",
				Status:   &status,
				Metadata: deploymentbuild.BuildMetadata{Name: strconv.Itoa(deploymentTrackBuild.Id), ComponentName: componentName},
				Spec:     buildSpec,
			})
		}
		buildsChan <- tempBuilds
		close(buildsChan) // Close the channel after writing
	}()

	// Goroutine to fetch deploymentTrackImages (only if necessary)
	if displayType != component.DisplayTypeGitProxy {
		wg.Add(1)
		go func() {
			defer wg.Done() // Ensure wg.Done is called no matter what happens
			deploymentTrackImages, errImages = auth.ComponentClient.GetDeploymentTrackImages(componentId, versionId, orgId)
			if errImages != nil {
				imagesChan <- nil
				return
			}
			imagesChan <- deploymentTrackImages
			close(imagesChan) // Close the channel after writing
		}()
	} else {
		close(imagesChan) // No need for goroutine, close it immediately if not fetching
	}

	// Goroutine to fetch commit history
	wg.Add(1)
	go func() {
		defer wg.Done() // Ensure wg.Done is called no matter what happens
		commitHistory, errCommits = auth.GitClient.GetCommitHistory(
			componentId,
			branch,
			orgId,
		)
		if errCommits != nil {
			commitsChan <- nil
			return
		}
		commitsChan <- commitHistory
		close(commitsChan) // Close the channel after writing
	}()

	// Wait for all goroutines to finish
	wg.Wait()

	// Collect results from the channels
	builds := <-buildsChan
	deploymentTrackImages = <-imagesChan
	commitHistory = <-commitsChan

	// Build maps for Git commit and images
	revisionToGitCommit := map[string]deploymentbuild.BuildStatusCommit{}
	for _, gitCommit := range commitHistory {
		// Convert platformgit.CommitHistory to deploymentbuild.BuildStatusCommit
		revisionToGitCommit[gitCommit.Sha] = deploymentbuild.BuildStatusCommit{
			Email:   gitCommit.Author.Email,
			Message: gitCommit.Message,
			Author:  gitCommit.Author.Name,
			Date:    gitCommit.Author.Date,
		}
	}

	revisionToBuildImages := make(map[string][]deploymentbuild.BuildStatusImage)
	for _, deploymentTrackImage := range deploymentTrackImages {
		existingImages, hasKey := revisionToBuildImages[deploymentTrackImage.CommitHash]
		item := deploymentbuild.BuildStatusImage{
			ID:        deploymentTrackImage.ImageId,
			CreatedAt: deploymentTrackImage.CreatedAt,
			UpdatedAt: deploymentTrackImage.UpdatedAt,
		}
		if hasKey {
			images := append(existingImages, item)
			revisionToBuildImages[deploymentTrackImage.CommitHash] = images
		} else {
			revisionToBuildImages[deploymentTrackImage.CommitHash] = []deploymentbuild.BuildStatusImage{item}
		}
	}

	// Final builds processing
	for i, build := range builds {
		buildSpec := build.Spec
		builds[i].Status.Images = revisionToBuildImages[buildSpec.Revision]
		builds[i].Status.GitCommit = revisionToGitCommit[buildSpec.Revision]
	}

	return builds, nil
}
