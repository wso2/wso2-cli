package userstore_mgt

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/wso2/integration-platform-tools/pkg/api"
)

type UserStoreManagementClient struct {
	baseApiURL string
	tokenStore api.ReadOnlyTokenStore
}

func NewUserStoreManagementClient(baseApiURL string, tokenStore api.ReadOnlyTokenStore) *UserStoreManagementClient {
	return &UserStoreManagementClient{
		baseApiURL: baseApiURL,
		tokenStore: tokenStore,
	}
}

func (c *UserStoreManagementClient) getHttpClient() *http.Client {
	client := &http.Client{}
	if os.Getenv("TRACE_ENABLED") == "true" {
		client.Transport = api.CreateLogger()(client.Transport)
	}
	return client
}

type User struct {
	Username  string
	Password  string
	Groups    string
	FirstName string
	LastName  string
	Email     string
}

type UserStoreAssociation struct {
	UserStoreID string `json:"userStoreId"`
}

func (c *UserStoreManagementClient) GetTestUsers(orgUuid, orgId, envId string) ([]User, error) {
	// 1. get user store association
	url := fmt.Sprintf("%s/user-stores/associations?orgId=%s&environmentId=%s", c.baseApiURL, orgUuid, envId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	token, err := c.tokenStore.GetTokenForOrg(orgId, true)
	if err != nil {
		return nil, fmt.Errorf("error getting token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	client := c.getHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to get user store associations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // No associations found, which is not an error
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected response status for getting associations: %s, body: %s", resp.Status, string(bodyBytes))
	}

	var associations []UserStoreAssociation
	if err := json.NewDecoder(resp.Body).Decode(&associations); err != nil {
		return nil, fmt.Errorf("error decoding user store associations: %w", err)
	}

	if len(associations) == 0 {
		return nil, nil // No user store associated
	}

	userStoreId := associations[0].UserStoreID

	// 2. download users csv
	downloadUrl := fmt.Sprintf("%s/user-stores/%s/download?orgId=%s", c.baseApiURL, userStoreId, orgUuid)
	req, err = http.NewRequest("GET", downloadUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to download users: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to download users: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected response status for downloading users: %s, body: %s", resp.Status, string(bodyBytes))
	}

	csvReader := csv.NewReader(resp.Body)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading csv body: %w", err)
	}

	if len(records) <= 1 {
		return []User{}, nil
	}

	var users []User
	// skip header row
	for _, record := range records[1:] {
		users = append(users, User{
			Username:  record[0],
			Password:  record[1],
			Groups:    record[2],
			FirstName: record[3],
			LastName:  record[4],
			Email:     record[5],
		})
	}

	return users, nil
}

func (c *UserStoreManagementClient) DeleteUserStore(orgUuid, orgId, userStoreId string) error {
	url := fmt.Sprintf("%s/user-stores/%s?orgId=%s", c.baseApiURL, userStoreId, orgUuid)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating delete request: %w", err)
	}

	token, err := c.tokenStore.GetTokenForOrg(orgId, true)
	if err != nil {
		return fmt.Errorf("error getting token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	client := c.getHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected response status for deleting user store: %s, body: %s", resp.Status, string(bodyBytes))
	}

	return nil
}

func (c *UserStoreManagementClient) AddTestUser(orgUuid, orgId, envId string, user User) error {
	// 1. Get existing user store associations to check if one exists
	url := fmt.Sprintf("%s/user-stores/associations?orgId=%s&environmentId=%s", c.baseApiURL, orgUuid, envId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	token, err := c.tokenStore.GetTokenForOrg(orgId, true)
	if err != nil {
		return fmt.Errorf("error getting token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	client := c.getHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request to get user store associations: %w", err)
	}
	defer resp.Body.Close()

	var existingUsers []User
	// If associations exist, delete the existing user store first
	if resp.StatusCode == http.StatusOK {
		var associations []UserStoreAssociation
		if err := json.NewDecoder(resp.Body).Decode(&associations); err != nil {
			return fmt.Errorf("error decoding user store associations: %w", err)
		}

		if len(associations) > 0 {
			userStoreId := associations[0].UserStoreID

			// Get existing users before deletion
			existingUsers, err = c.GetTestUsers(orgUuid, orgId, envId)
			if err != nil {
				return fmt.Errorf("failed to get existing test users before deletion: %w", err)
			}

			// Delete the existing user store
			err = c.DeleteUserStore(orgUuid, orgId, userStoreId)
			if err != nil {
				return fmt.Errorf("failed to delete existing user store: %w", err)
			}
		}
	}

	// Check if user already exists in the list
	for _, u := range existingUsers {
		if u.Username == user.Username {
			return fmt.Errorf("user with username '%s' already exists", user.Username)
		}
	}

	allUsers := append(existingUsers, user)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add userstore name
	formFieldName := "name"
	formFieldValue := fmt.Sprintf("userstore_%s_%s", orgUuid, envId)
	if err := writer.WriteField(formFieldName, formFieldValue); err != nil {
		return fmt.Errorf("error writing form field '%s': %w", formFieldName, err)
	}

	// Add file part
	part, err := writer.CreateFormFile("userstoreFile", "users.csv")
	if err != nil {
		return fmt.Errorf("error creating form file: %w", err)
	}

	// Write CSV header
	csvHeader := "username,password,groups,first_name,last_name,email\n"
	_, err = io.WriteString(part, csvHeader)
	if err != nil {
		return fmt.Errorf("error writing csv header: %w", err)
	}

	// Write CSV data
	for _, u := range allUsers {
		csvRow := fmt.Sprintf("\"%s\",\"%s\",\"[%s]\",\"%s\",\"%s\",\"%s\"\n",
			u.Username, u.Password, u.Groups, u.FirstName, u.LastName, u.Email)
		_, err = io.WriteString(part, csvRow)
		if err != nil {
			return fmt.Errorf("error writing csv row: %w", err)
		}
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("error closing multipart writer: %w", err)
	}

	createUrl := fmt.Sprintf("%s/user-stores?orgId=%s&associateToEnv=%s", c.baseApiURL, orgUuid, envId)
	createReq, err := http.NewRequest("POST", createUrl, body)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	createReq.Header.Set("Authorization", "Bearer "+token)
	createReq.Header.Set("Content-Type", writer.FormDataContentType())

	createResp, err := client.Do(createReq)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(createResp.Body)
		return fmt.Errorf("unexpected response status: %s, body: %s", createResp.Status, string(bodyBytes))
	}

	return nil
}
