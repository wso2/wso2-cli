package marketplace

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/util/common"
)

type MarketplaceClient struct {
	marketplaceApiBaseUrl string
	client                *api.IPHTTPClient
}

func NewMarketplaceClient(marketplaceApiBaseUrl string, tokenStore api.ReadOnlyTokenStore) *MarketplaceClient {
	return &MarketplaceClient{
		marketplaceApiBaseUrl: marketplaceApiBaseUrl,
		client:                api.NewIPHTTPClient(tokenStore),
	}
}

func (c *MarketplaceClient) GetMarketplaceServices(orgId string, reqBody MarketplaceGetServicesReq) (MarketplaceResp, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/services?%s", c.marketplaceApiBaseUrl, common.ToQueryParams(reqBody)), nil)
	if err != nil {
		return MarketplaceResp{}, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return MarketplaceResp{}, fmt.Errorf("error while fetching marketplace services: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return MarketplaceResp{}, fmt.Errorf("error while reading response: %w", err)
	}

	var response MarketplaceResp

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return MarketplaceResp{}, err
	}

	return response, nil
}

func (c *MarketplaceClient) GetMarketplaceServiceDetails(orgId string, serviceId string, visibility string) (MarketplaceService, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/services/%s?visibility=%s", c.marketplaceApiBaseUrl, serviceId, strings.ToUpper(visibility)), nil)
	if err != nil {
		return MarketplaceService{}, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return MarketplaceService{}, fmt.Errorf("error while fetching marketplace service details: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return MarketplaceService{}, fmt.Errorf("error while reading response: %w", err)
	}

	var response MarketplaceService

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return MarketplaceService{}, err
	}

	return response, nil
}

func (c *MarketplaceClient) GetMarketplaceServiceIdl(orgId string, serviceId string) (MarketplaceIdlResp, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/services/%s/idl", c.marketplaceApiBaseUrl, serviceId), nil)
	if err != nil {
		return MarketplaceIdlResp{}, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return MarketplaceIdlResp{}, fmt.Errorf("error while fetching marketplace service idl: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return MarketplaceIdlResp{}, fmt.Errorf("error while reading response: %w", err)
	}

	var response MarketplaceIdlResp

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return MarketplaceIdlResp{}, err
	}

	return response, nil
}

func (c *MarketplaceClient) GetMarketplaceGuide(orgId string, orgUuid string, serviceId string, configGroupId string, connectionName string, connectionSchemaId string, audience string, isSpa bool, isProjectLvlConnection bool, buildpackType string, configFileType string) (string, error) {
	componentType := "SERVICE"
	if isSpa {
		componentType = "WEB_APP_SPA"
	}

	reqBody := ConnectionGuideReq{
		OrgId:                  orgUuid,
		ConnectionName:         connectionName,
		ConfigGroupId:          configGroupId,
		Audience:               audience,
		ComponentType:          componentType,
		IsProjectLvlConnection: isProjectLvlConnection,
		BuildpackType:          buildpackType,
		ConfigFileType:         configFileType,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/services/%s/connection-schemas/%s/how-to-use?%s", c.marketplaceApiBaseUrl, serviceId, connectionSchemaId, common.ToQueryParams(reqBody)), nil)
	if err != nil {
		return "", fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return "", fmt.Errorf("error while fetching marketplace services: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("error while reading response: %w", err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

func (c *MarketplaceClient) GetMarketplaceDatabases(orgId string) (*[]MarketplaceDatabaseResource, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/databases", c.marketplaceApiBaseUrl), nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching marketplace databases: %w", err)
	}
	defer resp.Body.Close()

	var response MarketplaceDatabasesResp
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}

	return &response.Data, nil
}

// GetDatabases is a method that fetches the list of databases from the marketplace API.
func (c *MarketplaceClient) GetDatabases(orgId string) (ListDatabasesResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/databases", c.marketplaceApiBaseUrl), nil)
	if err != nil {
		return ListDatabasesResponse{}, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return ListDatabasesResponse{}, fmt.Errorf("error while fetching databases: %w", err)
	}
	defer resp.Body.Close()

	var response ListDatabasesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return ListDatabasesResponse{}, err
	}

	return response, nil
}

func (c *MarketplaceClient) GetDatabaseConnectionGuide(orgId string, orgUuid string, serviceId string, connectionSchemaId string, connectionName string, buildpackType string) (string, error) {
	values := url.Values{}
	values.Add("audience", "mcp")
	values.Add("orgId", orgUuid)
	values.Add("configFileType", "component_v11")
	values.Add("connectionName", connectionName)
	values.Add("buildpackType", buildpackType)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/databases/%s/connection-schemas/%s/how-to-use?%s", c.marketplaceApiBaseUrl, serviceId, connectionSchemaId, values.Encode()), nil)
	if err != nil {
		return "", fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return "", fmt.Errorf("error while fetching database connection guide: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

func (c *MarketplaceClient) GetServiceConnectionHowToUse(orgId, orgUuid, serviceId, schemaId, groupUuid, componentType, buildpackType, configFileType, connectionName, audience string, isProjectLvlConnection bool) (string, error) {
	reqBody := ConnectionGuideReq{
		ConfigGroupId:          groupUuid,
		Audience:               audience,
		OrgId:                  orgUuid,
		ComponentType:          componentType,
		IsProjectLvlConnection: isProjectLvlConnection,
		ConfigFileType:         configFileType,
		ConnectionName:         connectionName,
		BuildpackType:          buildpackType,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/services/%s/connection-schemas/%s/how-to-use?%s", c.marketplaceApiBaseUrl, serviceId, schemaId, common.ToQueryParams(reqBody)), nil)
	if err != nil {
		return "", fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return "", fmt.Errorf("error while fetching service connection guide: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error while reading response: %w", err)
	}
	return string(body), nil
}

func (c *MarketplaceClient) GetThirdPartyServiceOpenApi(orgId string, resourceId string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/services/%s/idl", c.marketplaceApiBaseUrl, resourceId), nil)
	if err != nil {
		return "", fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return "", fmt.Errorf("error while fetching third party service openapi: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("error while fetching third party service openapi: %w", err)
	}

	// json unmarshal the body
	var body map[string]interface{}
	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		return "", fmt.Errorf("error while unmarshalling response: %w", err)
	}

	// body.content is sometimes url encoded, so we need to decode it
	content, err := url.QueryUnescape(body["content"].(string))
	if err == nil {
		body["content"] = content
	}

	// json marshal the body
	bodyBytes, err = json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("error while marshalling response: %w", err)
	}

	return string(bodyBytes), nil
}
