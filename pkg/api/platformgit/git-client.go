package platformgit

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wso2/integration-platform-tools/pkg/api"
)

type GHAppConfig struct {
	AppUrl      string
	InstallUrl  string
	AuthUrl     string
	ClientId    string
	RedirectUrl string
}

type GitClient struct {
	projectApiUrl       string
	componentUtilApiUrl string
	ghAppConfig         *GHAppConfig
	client              *api.IPHTTPClient
}

func NewGitClient(
	projectApiUrl string,
	componentUtilApiUrl string,
	ghAppConfig *GHAppConfig,
	tokenStore api.ReadOnlyTokenStore,
) *GitClient {
	return &GitClient{
		projectApiUrl:       projectApiUrl,
		componentUtilApiUrl: componentUtilApiUrl,
		ghAppConfig:         ghAppConfig,
		client:              api.NewIPHTTPClient(tokenStore),
	}
}

func (c *GitClient) GetOauthURL(callbackUri string) (string, error) {
	state, err := c.getOauthState(callbackUri)
	if err != nil {
		return "", err
	}
	authUrl := c.ghAppConfig.AuthUrl
	redirectUrl := c.ghAppConfig.RedirectUrl
	clientId := c.ghAppConfig.ClientId
	return fmt.Sprintf("%s?redirect_uri=%s&client_id=%s&state=%s", authUrl, redirectUrl, clientId, state), nil
}

func (c *GitClient) getOauthState(callbackUri string) (string, error) {
	state := map[string]string{
		"origin":      "vscode.wso2ip.ext",
		"callbackUri": callbackUri,
	}
	stateJson, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(stateJson), nil
}

func (c *GitClient) ObtainAccessToken(code string, orgId string) error {
	query := ObtainUserTokenMutation(code)

	req, err := http.NewRequest("POST", c.projectApiUrl, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return fmt.Errorf("error while fetching access token: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error while reading response: %w", err)
	}
	type Response struct {
		Data struct {
			ObtainUserToken struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
			} `json:"obtainUserToken"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	if !response.Data.ObtainUserToken.Success {
		return fmt.Errorf("error while obtaining access token: %s", response.Data.ObtainUserToken.Message)
	}

	return nil

}

func (c *GitClient) GetAuthorizedGithubRepos(orgId, credRef string) (githubOrgs []GithubOrganization, err error) {
	query := GetAuthorizedGithubReposQuery(credRef)

	req, err := http.NewRequest("POST", c.projectApiUrl, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching authorized git repos: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}
	type Response struct {
		Data struct {
			UserRepos []GithubOrganization `json:"userRepos"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	githubOrgs = response.Data.UserRepos
	return githubOrgs, nil
}

func (c *GitClient) GetRepoBranches(gitOrgName, gitRepoName, orgId, credRef string) (branches []string, err error) {
	query := GetRepoBranchesQuery(gitOrgName, gitRepoName, credRef)

	req, err := http.NewRequest("POST", c.projectApiUrl, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching git repo branches: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}
	type Response struct {
		Data struct {
			RepoBranchList []struct {
				Name string `json:"name"`
			} `json:"repoBranchList"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	var branchNames []string
	for _, branch := range response.Data.RepoBranchList {
		branchNames = append(branchNames, branch.Name)
	}

	return branchNames, nil
}

func (c *GitClient) GetGithubInstallUrl() string {
	return c.ghAppConfig.InstallUrl
}

func (c *GitClient) GetCommitHistory(
	componentId string,
	branch string,
	orgId string,
) (commitHistory []CommitHistory, err error) {

	query := GetCommitHistory(componentId, branch)

	req, err := http.NewRequest("POST", c.projectApiUrl, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching commit history: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}
	type Response struct {
		Data struct {
			CommitHistory []CommitHistory `json:"commitHistory"`
		} `json:"data"`
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	commitHistory = response.Data.CommitHistory
	return commitHistory, nil
}

func (c *GitClient) IsPublicRepo(
	gitOrgName string,
	gitRepoName string,
	orgId string,
) (IsPublicRepoResponse, error) {
	url := fmt.Sprintf(
		"%s/repositories/%s/%s/visibility-level",
		c.componentUtilApiUrl,
		gitOrgName,
		gitRepoName,
	)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return IsPublicRepoResponse{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := c.client.Do(req, orgId)

	if err != nil {
		return IsPublicRepoResponse{}, fmt.Errorf("error while executing request: %w", err)
	}

	defer res.Body.Close()

	type Response struct {
		Success bool                 `json:"success"`
		Data    IsPublicRepoResponse `json:"data"`
		Message string               `json:"message"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return IsPublicRepoResponse{}, fmt.Errorf("error while decoding response: %w", err)
	}

	return response.Data, nil
}
