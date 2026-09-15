package auth

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/clirpc/server"
	"github.com/wso2/integration-platform-tools/internal/region"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

func GetUserInfo(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.UserNotFound
	}

	user, err := auth.GetCurrentUser()
	if err != nil {
		return server.Result{}, api.UserNotFound
	}

	return server.CreateResult(struct {
		IsLoggedIn bool         `json:"isLoggedIn"`
		UserInfo   api.UserInfo `json:"userInfo"`
	}{
		IsLoggedIn: true,
		UserInfo:   *user,
	}), nil
}

func SignOutUser(req *json.RawMessage) (server.Result, error) {
	err := auth.SignOut()
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct{}{}), nil
}

func GetCurrentRegion(req *json.RawMessage) (server.Result, error) {
	region := region.GetCurrentRegion()

	return server.CreateResult(struct {
		Region string `json:"region"`
	}{
		Region: region,
	}), nil
}

func GetSignInAuthUrl(req *json.RawMessage) (server.Result, error) {
	var request struct {
		CallbackUrl string `json:"callbackUrl"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	regionStrs := []string{}
	clientIds := []string{}
	regions := region.GetValidRegions()
	for _, regionStr := range regions {
		regionStrs = append(regionStrs, regionStr)
		regionConfig := region.GetConfigByRegion(regionStr)
		clientIds = append(clientIds, regionConfig.AsgardeoClientId)
	}

	config := region.GetRegionConfig()

	url, err := auth.GetAuthUrl(request.CallbackUrl, config.ConsoleUrls.LoginUrl, config.AsgardeoClientId, regionStrs, clientIds)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		LoginUrl string `json:"loginUrl"`
	}{
		LoginUrl: url,
	}), nil
}

func GetDevantSignInAuthUrl(req *json.RawMessage) (server.Result, error) {
	var request struct {
		CallbackUrl string `json:"callbackUrl"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	regionStrs := []string{}
	clientIds := []string{}
	regions := region.GetValidRegions()
	for _, regionStr := range regions {
		regionStrs = append(regionStrs, regionStr)
		regionConfig := region.GetConfigByRegion(regionStr)
		clientIds = append(clientIds, regionConfig.DevantConfig.AsgardeoClientId)
	}

	config := region.GetRegionConfig()

	url, err := auth.GetAuthUrl(request.CallbackUrl, config.DevantConfig.ConsoleUrls.LoginUrl, config.DevantConfig.AsgardeoClientId, regionStrs, clientIds)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		LoginUrl string `json:"loginUrl"`
	}{
		LoginUrl: url,
	}), nil
}

func SignInWithAuthCode(req *json.RawMessage) (server.Result, error) {
	var request struct {
		AuthCode string `json:"authCode"`
		OrgId    string `json:"orgId"`
		Region   string `json:"region"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	if request.Region != "" {
		if err := region.SetCurrentRegion(request.Region); err != nil {
			return server.Result{}, err
		}
		auth.ReInit()
	}

	config := region.GetRegionConfig()

	userInfo, isNewUser, err := auth.SignInWithAuthCode(request.AuthCode, request.OrgId, config.ConsoleUrls.RedirectUrl, config.AsgardeoClientId)
	if err != nil {
		return server.Result{}, err
	}

	return initOrgWhenSignin(userInfo, isNewUser)
}

func SignInDevantWithAuthCode(req *json.RawMessage) (server.Result, error) {
	var request struct {
		AuthCode string `json:"authCode"`
		OrgId    string `json:"orgId"`
		Region   string `json:"region"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	if request.Region != "" {
		if err := region.SetCurrentRegion(request.Region); err != nil {
			return server.Result{}, err
		}
		auth.ReInit()
	}

	config := region.GetRegionConfig()

	userInfo, isNewUser, err := auth.SignInWithAuthCode(request.AuthCode, request.OrgId, config.DevantConfig.ConsoleUrls.RedirectUrl, config.DevantConfig.AsgardeoClientId)
	if err != nil {
		return server.Result{}, err
	}

	return initOrgWhenSignin(userInfo, isNewUser)
}

func initOrgWhenSignin(userInfo *api.UserInfo, isNewUser bool) (server.Result, error) {
	selectedOrg, err := auth.GetSelectedOrganization()
	if err == nil && selectedOrg != nil {
		deploymentPipelineData, err := auth.DevopsClient.GetDeploymentPipeline(selectedOrg.UUID, selectedOrg.ID)
		if err == nil && len(deploymentPipelineData) == 0 {
			auth.DevopsClient.InitOrgRegion(selectedOrg.ID, selectedOrg.UUID)
		}
	}

	if isNewUser {
		// create initial project
		orgIdInt, _ := strconv.Atoi(selectedOrg.ID)
		auth.ProjectClient.CreateProject(project.GetProjectMutationRequest{
			Name:       "Default",
			Region:     region.GetCurrentRegion(),
			Version:    "1.0.0",
			OrgID:      orgIdInt,
			OrgHandler: selectedOrg.Handle,
		}, selectedOrg.ID)
	}

	return server.CreateResult(struct {
		UserInfo api.UserInfo `json:"userInfo"`
	}{
		UserInfo: *userInfo,
	}), nil
}

func GetCurrentOrg(req *json.RawMessage) (server.Result, error) {
	selectedOrg, err := auth.GetSelectedOrganization()
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(selectedOrg), nil
}

func ChangeOrg(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.UserNotFound
	}
	var request struct {
		OrgId string `json:"orgId"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	userInfo, err := auth.GetCurrentUser()
	if err != nil {
		return server.Result{}, err
	}

	var targetOrg *api.Organization

	for _, org := range userInfo.Organizations {
		if org.ID == request.OrgId {
			targetOrg = &org
			break
		}
	}

	if targetOrg == nil {
		return server.Result{}, fmt.Errorf("organization with id %s not found", request.OrgId)
	}

	err = auth.SetSelectedOrg(targetOrg, userInfo.Organizations)
	if err != nil {
		return server.Result{}, err
	}

	selectedOrg, err := auth.GetSelectedOrganization()
	if err == nil && selectedOrg != nil {
		deploymentPipelineData, err := auth.DevopsClient.GetDeploymentPipeline(selectedOrg.UUID, selectedOrg.ID)
		if err == nil && len(deploymentPipelineData) == 0 {
			auth.DevopsClient.InitOrgRegion(selectedOrg.ID, selectedOrg.UUID)
		}
	}

	return server.CreateResult(struct{}{}), nil
}

func GetSubscriptions(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.UserNotFound
	}
	var request struct {
		OrgId     string `json:"orgId"`
		CloudType string `json:"cloudType"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	userInfo, err := auth.GetCurrentUser()
	if err != nil {
		return server.Result{}, err
	}

	var targetOrg *api.Organization

	for _, org := range userInfo.Organizations {
		if org.ID == request.OrgId {
			targetOrg = &org
			break
		}
	}

	if targetOrg == nil {
		return server.Result{}, fmt.Errorf("organization with id %s not found", request.OrgId)
	}

	subResp, err := auth.SubscriptionsClient.GetSubscriptions(targetOrg.ID, targetOrg.UUID, request.CloudType)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(subResp), nil
}

func GetStsToken(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.UserNotFound
	}

	token, err := auth.GetStsToken()
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		Token string `json:"token"`
	}{
		Token: token,
	}), nil
}

func getConfigs(req *json.RawMessage) (server.Result, error) {
	config := region.GetRegionConfig()

	type GhAppConfig struct {
		InstallUrl string `json:"installUrl"`
		AuthUrl    string `json:"authUrl"`
		ClientId   string `json:"clientId"`
	}

	return server.CreateResult(struct {
		BillingConsoleUrl string      `json:"billingConsoleUrl"`
		ConsoleUrl        string      `json:"consoleUrl"`
		DevantConsoleUrl  string      `json:"devantConsoleUrl"`
		GhApp             GhAppConfig `json:"ghApp"`
	}{
		BillingConsoleUrl: config.BillingConsoleBaseUrl,
		ConsoleUrl:        config.ConsoleUrls.BaseUrl,
		DevantConsoleUrl:  config.DevantConfig.ConsoleUrls.BaseUrl,
		GhApp: GhAppConfig{
			InstallUrl: config.GhApp.InstallUrl,
			AuthUrl:    config.GhApp.AuthUrl,
			ClientId:   config.GhApp.ClientId,
		},
	}), nil
}
