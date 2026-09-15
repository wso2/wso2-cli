package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type UserManagementClient struct {
	baseApiURL string
}

var ErrNoAccountFound = errors.New("no account found")

func NewUserManagementClient(baseApiURL string) *UserManagementClient {
	return &UserManagementClient{baseApiURL: baseApiURL}
}

func (c *UserManagementClient) getHttpClient() *http.Client {
	client := &http.Client{}
	if os.Getenv("TRACE_ENABLED") == "true" {
		client.Transport = CreateLogger()(client.Transport)
	}
	return client
}

func (c *UserManagementClient) ValidateUser(asgardioToken string) (*UserInfo, error) {
	client := c.getHttpClient()
	req, err := http.NewRequest("GET", c.baseApiURL+"/validate/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+asgardioToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while fetching user info: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		if resp.StatusCode == 401 {
			return nil, fmt.Errorf("unauthorized access")
		} else {
			bodyBytes, _ := io.ReadAll(resp.Body)
			stringReader := strings.NewReader(string(bodyBytes))
			var validateUserFailure ValidateUserFailure
			if err := json.NewDecoder(stringReader).Decode(&validateUserFailure); err == nil {
				if validateUserFailure.Code == 1003 {
					return nil, ErrNoAccountFound
				}
			}
			return nil, fmt.Errorf("unexpected response: %s: status code: %d", string(bodyBytes), resp.StatusCode)
		}
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	return &userInfo, nil
}

func (c *UserManagementClient) IsValidNewOrg(asgardioToken, newOrgName, userEmail string) (bool, error) {
	params := url.Values{}
	params.Add("orgName", newOrgName)
	params.Add("email", userEmail)
	queryString := params.Encode()

	client := c.getHttpClient()
	req, err := http.NewRequest("GET", c.baseApiURL+"/validate/orgname?"+queryString, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("Authorization", "Bearer "+asgardioToken)
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error while fetching user info: %w", err)
	}
	defer resp.Body.Close()
	type Response struct {
		IsValid            bool   `json:"isValid"`
		OrganizationHandle string `json:"organizationHandle"`
	}

	var response Response

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, err
	}

	return response.IsValid, nil
}

func (c *UserManagementClient) RegisterUser(asgardioToken, newOrgName string) (*UserInfo, error) {
	type ReqBody struct {
		Organization struct {
			Name string `json:"name"`
		} `json:"organization"`
		ServiceName   string `json:"serviceName"`
		TermsAccepted bool   `json:"termsAccepted"`
	}
	reqBody := ReqBody{
		Organization: struct {
			Name string "json:\"name\""
		}{Name: newOrgName},
		ServiceName:   "choreo",
		TermsAccepted: true,
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseApiURL+"/register-user", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error while encoding request body: %w", err)
	}

	client := c.getHttpClient()
	req.Header.Add("Authorization", "Bearer "+asgardioToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while registering user info: %w", err)
	}
	defer resp.Body.Close()

	var response UserInfo

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *UserManagementClient) GetBaseApiURL() string {
	return c.baseApiURL
}
