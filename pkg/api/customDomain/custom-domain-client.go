package customDomain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wso2/integration-platform-tools/pkg/api"
	workflowmgt "github.com/wso2/integration-platform-tools/pkg/api/workflow-mgt"
)

type CustomDomainClient struct {
	customDomainApiBaseUrl string
	urlMappingApiBaseUrl   string
	Client                 *api.IPHTTPClient
}

func NewCustomDomainClient(customDomainApiBaseUrl string, urlMappingApiBaseUrl string, tokenStore api.ReadOnlyTokenStore) *CustomDomainClient {
	return &CustomDomainClient{
		customDomainApiBaseUrl: customDomainApiBaseUrl,
		urlMappingApiBaseUrl:   urlMappingApiBaseUrl,
		Client:                 api.NewIPHTTPClient(tokenStore),
	}
}

func (c *CustomDomainClient) GetCustomDomains(orgId string) ([]CustomDomain, error) {
	url := c.customDomainApiBaseUrl
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return []CustomDomain{}, fmt.Errorf("error while creating request: %w", err)
	}
	fmt.Println("Custom Domain URL: ", url)
	res, err := c.Client.Do(req, orgId)

	if err != nil {
		fmt.Println("Error while executing request: ", err)
		return []CustomDomain{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	var response []CustomDomain
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return []CustomDomain{}, fmt.Errorf("error while decoding response: %w", err)
	}

	return response, nil
}

func (c *CustomDomainClient) ValidateCustomDomainForRegistration(orgId string, domainName string, componentType string, envId string) (validity CustomDomainValidateResp, err error) {
	url := c.customDomainApiBaseUrl + "/validate"
	reqBody := CustomDomainValidateReq{
		DomainName:    domainName,
		ComponentType: componentType,
		EnvID:         envId,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return CustomDomainValidateResp{}, fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return CustomDomainValidateResp{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.Client.Do(req, orgId)
	if err != nil {
		return CustomDomainValidateResp{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()
	var response CustomDomainValidateResp

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return CustomDomainValidateResp{}, fmt.Errorf("error while decoding response: %w", err)
	}

	return response, nil
}

func (c *CustomDomainClient) RegisterCustomDomain(orgId string, domainName string, componentType string, envId string, tlsProviderName string, enableAutoApply bool) (CustomDomain, error) {
	url := c.customDomainApiBaseUrl
	reqBody := CustomDomainCreateReq{
		DomainName: domainName,
		Environment: struct {
			ID string `json:"id"`
		}{
			ID: envId,
		},
		ComponentType: componentType,
		TLSProvider: struct {
			Name string `json:"name"`
		}{
			Name: tlsProviderName,
		},
		EnableAutoApply: enableAutoApply,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return CustomDomain{}, fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return CustomDomain{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.Client.Do(req, orgId)
	if err != nil {
		return CustomDomain{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	var response CustomDomain
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return CustomDomain{}, fmt.Errorf("error while decoding response: %w", err)
	}

	return response, nil
}

func (c *CustomDomainClient) CreateUrlMapping(orgId string, domainId string, componentId string, apiId string, defaultDomain string, customPath string, defaultPath string, workflowMgtEnabled bool, metaData *workflowmgt.RequestUrlCustomizationMetadata) (CreateUrlMappingRes, error) {
	url := c.urlMappingApiBaseUrl

	reqBody := CreateUrlMappingReq{
		DomainID:      domainId,
		ComponentID:   componentId,
		ApiID:         apiId,
		DefaultDomain: defaultDomain,
		CustomPath:    customPath,
		DefaultPath:   defaultPath,
	}

	if !workflowMgtEnabled {
		url += "/deploy"
	} else {
		reqBody.MetaData = metaData
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return CreateUrlMappingRes{}, fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return CreateUrlMappingRes{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.Client.Do(req, orgId)
	if err != nil {
		return CreateUrlMappingRes{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	var response CreateUrlMappingRes
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return CreateUrlMappingRes{}, fmt.Errorf("error while decoding response: %w", err)
	}

	return response, nil
}
