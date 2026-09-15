package platformservices

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wso2/integration-platform-tools/pkg/api"
)

type PlatformServicesClient struct {
	platformServicesApiBaseUrl string
	client                     *api.IPHTTPClient
}

func NewPlatformServicesClient(platformServicesApiBaseUrl string, tokenStore api.ReadOnlyTokenStore) *PlatformServicesClient {
	return &PlatformServicesClient{
		platformServicesApiBaseUrl: platformServicesApiBaseUrl,
		client:                     api.NewIPHTTPClient(tokenStore),
	}
}

func (c *PlatformServicesClient) GetDatabaseServicePlans(orgId, databaseType string) ([]ServicePlan, error) {
	req, err := http.NewRequest("GET", c.platformServicesApiBaseUrl+"/db-service-plans?type="+databaseType, nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}
	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while getting database service plans: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error while getting database service plans: received status code %d", resp.StatusCode)
	}
	var response []ServicePlan
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}
	return response, nil
}

func (c *PlatformServicesClient) CreateDatabaseServer(orgId string, createDatabaseServerReq *CreateDatabaseServerRequest) (*DatabaseServer, error) {
	jsonData, err := json.Marshal(createDatabaseServerReq)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling request: %w", err)
	}
	req, error := http.NewRequest("POST", c.platformServicesApiBaseUrl+"/db-servers", bytes.NewBuffer(jsonData))
	if error != nil {
		return nil, fmt.Errorf("error while creating request: %w", error)
	}
	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while creating database service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("error while creating database service: received status code %d", resp.StatusCode)
	}
	var response DatabaseServer
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}
	return &response, nil
}

func (c *PlatformServicesClient) GetDatabaseServer(orgId, databaseServerId string) (*DatabaseServer, error) {
	req, err := http.NewRequest("GET", c.platformServicesApiBaseUrl+"/db-servers/"+databaseServerId, nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}
	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while getting database service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error while getting database service: received status code %d", resp.StatusCode)
	}
	var response DatabaseServer
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}
	return &response, nil
}

func (c *PlatformServicesClient) GetDatabaseServers(orgId, orgUuid string) ([]DatabaseServer, error) {
	req, err := http.NewRequest("GET", c.platformServicesApiBaseUrl+"/db-servers?organization_id="+orgUuid, nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}
	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while getting database services: %w", err)
	}
	defer resp.Body.Close()
	var response []DatabaseServer
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}
	return response, nil
}

func (c *PlatformServicesClient) CreateDBInstance(orgId, databaseServerId string, createDBInstanceReq *CreateDatabaseRequest) (*Database, error) {
	jsonData, err := json.Marshal(createDBInstanceReq)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling request: %w", err)
	}
	req, err := http.NewRequest("POST", c.platformServicesApiBaseUrl+"/db-servers/"+databaseServerId+"/databases", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}
	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while creating database instance: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("error while creating database instance: received status code %d", resp.StatusCode)
	}
	var response Database
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}
	return &response, nil
}

func (c *PlatformServicesClient) GetDBInstances(orgId, databaseServerId string) ([]Database, error) {
	req, err := http.NewRequest("GET", c.platformServicesApiBaseUrl+"/db-servers/"+databaseServerId+"/databases", nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}
	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while getting database instances: %w", err)
	}
	defer resp.Body.Close()
	var response []Database
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}
	return response, nil
}

// ToDo: Improve this function to support user provided credentials as well.
func (c *PlatformServicesClient) CreateDBCredentials(orgId, databaseServerId string, createDBCredRequest *CreateDBCredRequest) (*DatabaseCredential, error) {
	jsonData, err := json.Marshal(createDBCredRequest)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling request: %w", err)
	}
	req, err := http.NewRequest("POST", c.platformServicesApiBaseUrl+"/db-servers/"+databaseServerId+"/credentials", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}
	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while creating database credentials: %w", err)
	}
	defer resp.Body.Close()
	var response DatabaseCredential
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}
	return &response, nil
}

func (c *PlatformServicesClient) GetDBCredentials(orgId, databaseServerId string) ([]DatabaseCredential, error) {
	req, err := http.NewRequest("GET", c.platformServicesApiBaseUrl+"/db-servers/"+databaseServerId+"/credentials", nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}
	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while getting database credentials: %w", err)
	}
	defer resp.Body.Close()
	var response []DatabaseCredential
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}
	return response, nil
}

func (c *PlatformServicesClient) PublishDatabase(orgId, databaseServerId, databaseName string, publishDatabaseRequest *PublishDatabaseRequest) (*PublishDatabaseResponse, error) {
	jsonData, err := json.Marshal(publishDatabaseRequest)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling request: %w", err)
	}
	req, err := http.NewRequest("PUT", c.platformServicesApiBaseUrl+"/db-servers/"+databaseServerId+"/databases/"+databaseName, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}
	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while publishing database: %w", err)
	}
	defer resp.Body.Close()
	var response PublishDatabaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}
	return &response, nil
}

func (c *PlatformServicesClient) GetDatabaseCredentials(orgId string, dbServerId string, dbName string) (*[]DatabaseCredential, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/db-servers/%s/credentials?db_name=%s", c.platformServicesApiBaseUrl, dbServerId, dbName), nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}
	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching database credentials: %w", err)
	}
	defer resp.Body.Close()

	var response []DatabaseCredential
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}

	return &response, nil
}
