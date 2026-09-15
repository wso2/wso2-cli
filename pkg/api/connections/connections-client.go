package connections

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

type ConnectionsClient struct {
	connectionsApiBaseUrl string
	client                *api.IPHTTPClient
}

func NewConnectionsClient(devopsApiBaseUrl string, tokenStore api.ReadOnlyTokenStore) *ConnectionsClient {
	return &ConnectionsClient{
		connectionsApiBaseUrl: devopsApiBaseUrl,
		client:                api.NewIPHTTPClient(tokenStore),
	}
}

func (c *ConnectionsClient) GetConnectionItem(orgId string, connectionId string) (Connection, error) {
	req, err := http.NewRequest("GET", c.connectionsApiBaseUrl+"/configurations/service-configs/connections/"+connectionId, nil)
	if err != nil {
		return Connection{}, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return Connection{}, fmt.Errorf("error while fetching connection item: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return Connection{}, fmt.Errorf("error while reading response: %w", err)
	}

	var response Connection

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return Connection{}, err
	}

	return response, nil
}

func (c *ConnectionsClient) GetConnectionList(orgId string, projectId string, componentId string) ([]Connection, error) {
	connectionScope := "scope=PROJECT"
	if componentId != "" {
		connectionScope = "componentId=" + componentId
	}
	req, err := http.NewRequest("GET", c.connectionsApiBaseUrl+"/configurations/service-configs/connections?projectId="+projectId+"&"+connectionScope, nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching connections: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	var response []Connection

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response, nil
}

func (c *ConnectionsClient) CreateNewConnection(
	orgId string,
	orgUuid string,
	projectId string,
	name string,
	serviceId string,
	schemaReference string,
	serviceVisibility string,
	envs []project.ProjectEnvironment,
	componentId string,
	generateCreds bool,
) (Connection, error) {

	var envItems []ConnectionReqEnv

	for _, env := range envs {
		envItems = append(envItems, ConnectionReqEnv{
			IsCritical: env.Critical,
			ID:         env.TemplateId,
		})
	}

	orgIDInteger, err := strconv.Atoi(orgId)
	if err != nil {
		return Connection{}, fmt.Errorf("error while converting org id: %w", err)
	}

	reqBody := ConnectionReqPayload{
		Name:                        name,
		Description:                 "",
		ServiceID:                   serviceId,
		SchemaReference:             schemaReference,
		RequestingServiceVisibility: serviceVisibility,
		OrgIDInteger:                orgIDInteger,
		Environments:                envItems,
	}

	if componentId != "" {
		reqBody.Visibilities = []ConnectionVisibility{{
			ProjectUUID:      projectId,
			OrganizationUUID: orgUuid,
			ComponentUuid:    componentId,
		}}
	} else {
		reqBody.Visibilities = []ConnectionVisibility{{
			ProjectUUID:      projectId,
			OrganizationUUID: orgUuid,
		}}
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return Connection{}, fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.connectionsApiBaseUrl+"/configurations/service-configs/choreo-connections?generateCreds="+strconv.FormatBool(generateCreds), bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return Connection{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)
	if err != nil {
		return Connection{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	var response Connection

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return Connection{}, err
	}

	return response, nil
}

func (c *ConnectionsClient) DeleteConnection(orgId string, connectionId string) error {
	req, err := http.NewRequest("DELETE", c.connectionsApiBaseUrl+"/configurations/service-configs/choreo-connections/"+connectionId, nil)
	if err != nil {
		return fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return fmt.Errorf("error while fetching connections: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error while reading response: %w", err)
	}

	return nil
}

func (c *ConnectionsClient) CreateDatabaseConnection(orgId string, connection CreateDatabaseConnectionReq) (*Connection, error) {
	reqBodyBytes, err := json.Marshal(connection)
	if err != nil {
		return nil, fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.connectionsApiBaseUrl+"/configurations/service-configs/choreo-database-connections", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	var response Connection

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *ConnectionsClient) CreateThirdPartyConnection(orgId string, connection CreateThirdPartyConnectionReq, wellKnownService bool) (*Connection, error) {
	reqBodyBytes, err := json.Marshal(connection)
	if err != nil {
		return nil, fmt.Errorf("error while encoding request body: %w", err)
	}

	url := c.connectionsApiBaseUrl + "/configurations/service-configs/third-party-connections"
	if wellKnownService {
		url += "?wellKnownService=true"
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	var response Connection

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
