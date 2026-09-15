package apim

import (
	"encoding/json"

	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/clirpc/server"
	"github.com/wso2/integration-platform-tools/pkg/api"
)

func GetEndpointTestKey(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		APIMID  string `json:"apimId"`
		OrgUuid string `json:"orgUuid"`
		OrgId   string `json:"orgId"`
		EnvName string `json:"envName"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	apiKeyResp, err := auth.ApimPublisherClient.GetApiKey(
		request.APIMID,
		request.OrgUuid,
		request.OrgId,
		request.EnvName,
	)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		ApiKey       string `json:"apiKey"`
		ValidityTime int    `json:"validityTime"`
	}{
		ApiKey:       apiKeyResp.Apikey,
		ValidityTime: apiKeyResp.ValidityTime,
	}), nil
}

func GetSwaggerSpec(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		ApimRevisionId string `json:"apimRevisionId"`
		OrgUuid        string `json:"orgUuid"`
		OrgId          string `json:"orgId"`
	}

	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	swaggerResp, err := auth.ApimPublisherClient.GetSwaggerSpec(
		request.ApimRevisionId,
		request.OrgUuid,
		request.OrgId,
	)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		Swagger any `json:"swagger"`
	}{
		Swagger: swaggerResp,
	}), nil
}
