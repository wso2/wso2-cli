package subscription

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/wso2/integration-platform-tools/pkg/api"
)

type SubscriptionsClient struct {
	billingApiBaseUrl string
	client            *api.IPHTTPClient
}

func NewSUbscriptionsClient(billingApiBaseUrl string, tokenStore api.ReadOnlyTokenStore) *SubscriptionsClient {
	return &SubscriptionsClient{
		billingApiBaseUrl: billingApiBaseUrl,
		client:            api.NewIPHTTPClient(tokenStore),
	}
}

func (c *SubscriptionsClient) GetSubscriptions(orgId string, orgUuid string, cloudType string) (GetSubscriptionsResp, error) {
	params := url.Values{}
	if cloudType == "" {
		params.Add("cloudType", "choreo")
	} else {
		params.Add("cloudType", cloudType)
	}
	params.Add("origin", "choreo-console")
	queryString := params.Encode()
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/organizations/%s/subscriptions?%s", c.billingApiBaseUrl, orgUuid, queryString), nil)
	if err != nil {
		return GetSubscriptionsResp{}, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return GetSubscriptionsResp{}, fmt.Errorf("error while fetching subscriptions: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return GetSubscriptionsResp{}, fmt.Errorf("error while reading response: %w", err)
	}

	var response GetSubscriptionsResp

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return GetSubscriptionsResp{}, err
	}

	return response, nil
}
