package deployment

import (
	"encoding/json"

	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/clirpc/server"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/marketplace"
)

func getMarketplaceItems(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId   string                                `json:"orgId"`
		Request marketplace.MarketplaceGetServicesReq `json:"request"`
	}

	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	marketplaceResp, err := auth.MarketplaceClient.GetMarketplaceServices(request.OrgId, request.Request)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(marketplaceResp), nil
}

func getMarketplaceItemIdl(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId     string `json:"orgId"`
		ServiceId string `json:"serviceId"`
	}

	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	marketplaceResp, err := auth.MarketplaceClient.GetMarketplaceServiceIdl(request.OrgId, request.ServiceId)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(marketplaceResp), nil
}

func createComponentConnection(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId             string `json:"orgId"`
		OrgUuid           string `json:"orgUuid"`
		ProjectId         string `json:"projectId"`
		ComponentId       string `json:"componentId"`
		ComponentPath     string `json:"componentPath"`
		ComponentType     string `json:"componentType"`
		ServiceId         string `json:"serviceId"`
		ServiceVisibility string `json:"serviceVisibility"`
		ServiceSchemaId   string `json:"serviceSchemaId"`
		Name              string `json:"name"`
		GenerateCreds     bool   `json:"generateCreds"`
	}

	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	envs, err := auth.ProjectClient.GetProjectEnvironments(request.OrgUuid, request.OrgId, request.ProjectId)
	if err != nil {
		return server.Result{}, err
	}

	createdConn, err := auth.ConnectionsClient.CreateNewConnection(request.OrgId, request.OrgUuid, request.ProjectId, request.Name, request.ServiceId, request.ServiceSchemaId, request.ServiceVisibility, *envs, request.ComponentId, request.GenerateCreds)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(createdConn), nil
}

func deleteConnection(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId         string `json:"orgId"`
		ConnectionId  string `json:"connectionId"`
		ComponentPath string `json:"componentPath"`
	}

	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	err := auth.ConnectionsClient.DeleteConnection(request.OrgId, request.ConnectionId)
	if err != nil {
		return server.Result{}, err
	}

	return server.Result{}, nil
}

func getConnections(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId       string `json:"orgId"`
		ProjectId   string `json:"projectId"`
		ComponentId string `json:"componentId"`
	}

	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	connectionsListResp, err := auth.ConnectionsClient.GetConnectionList(request.OrgId, request.ProjectId, request.ComponentId)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(connectionsListResp), nil
}

func GetConnectionItem(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId             string `json:"orgId"`
		ConnectionGroupId string `json:"connectionGroupId"`
	}

	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	connectionsListResp, err := auth.ConnectionsClient.GetConnectionItem(request.OrgId, request.ConnectionGroupId)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(connectionsListResp), nil
}

func getConnectionGuide(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId                  string `json:"orgId"`
		OrgUuid                string `json:"orgUuid"`
		ServiceId              string `json:"serviceId"`
		ConfigGroupId          string `json:"configGroupId"`
		ConnectionSchemaId     string `json:"connectionSchemaId"`
		ConnectionName         string `json:"connectionName"`
		Audience               string `json:"audience"`
		IsSpa                  bool   `json:"isSpa"`
		IsProjectLvlConnection bool   `json:"isProjectLvlConnection"`
		BuildpackType          string `json:"buildpackType"`
		ConfigFileType         string `json:"configFileType"`
	}

	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	guideStr, err := auth.MarketplaceClient.GetMarketplaceGuide(request.OrgId, request.OrgUuid, request.ServiceId, request.ConfigGroupId, request.ConnectionName, request.ConnectionSchemaId, request.Audience, request.IsSpa, request.IsProjectLvlConnection, request.BuildpackType, request.ConfigFileType)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		Guide string `json:"guide"`
	}{
		Guide: guideStr,
	}), nil
}
