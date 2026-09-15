package repo

import (
	"encoding/json"

	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/clirpc/server"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/credentials"
	"github.com/wso2/integration-platform-tools/pkg/api/platformgit"
	"github.com/wso2/integration-platform-tools/pkg/util/repo"
)

func GetRepoBranches(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		RepoUrl string `json:"repoUrl"`
		OrgId   string `json:"orgId"`
		CredRef string `json:"credRef"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	gitOrgName, gitRepoName, err := repo.ParseGitURL(request.RepoUrl)
	if err != nil {
		return server.Result{}, err
	}

	branches, err := auth.GitClient.GetRepoBranches(gitOrgName, gitRepoName, request.OrgId, request.CredRef)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		Branches []string `json:"branches"`
	}{
		Branches: branches,
	}), nil
}

func IsGitRepoAuthorized(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		RepoUrl string `json:"repoUrl"`
		OrgId   string `json:"orgId"`
		CredRef string `json:"credRef"`
	}
	type response struct {
		RetrievedRepos bool `json:"retrievedRepos"`
		IsAccessible   bool `json:"isAccessible"`
	}

	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	gitOrgName, gitRepoName, err := repo.ParseGitURL(request.RepoUrl)
	if err != nil {
		return server.Result{}, err
	}

	// isPublicResp, _ := auth.GitClient.IsPublicRepo(gitOrgName, gitRepoName, request.OrgId)
	// if isPublicResp.IsPublicRepo {
	// 	return server.CreateResult(response{IsAccessible: true, RetrievedRepos: false}), nil
	// }

	authorizedGitOrgs, err := auth.GitClient.GetAuthorizedGithubRepos(request.OrgId, request.CredRef)
	if err != nil {
		// Initiate Git auth dance to retrieve git token, if repos are not retrievable
		return server.CreateResult(response{IsAccessible: false, RetrievedRepos: false}), err
	}

	for _, authorizedGitOrg := range authorizedGitOrgs {
		if authorizedGitOrg.OrgName == gitOrgName || authorizedGitOrg.OrgHandler == gitOrgName {
			for _, authorizedGitRepo := range authorizedGitOrg.Repositories {
				if authorizedGitRepo.Name == gitRepoName {
					// Initiate git installation flow if repo has not been given access
					return server.CreateResult(response{IsAccessible: true, RetrievedRepos: true}), nil
				}
			}
		}
	}

	return server.CreateResult(response{IsAccessible: false, RetrievedRepos: true}), nil
}

func GetAuthorizedGitOrgs(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId   string `json:"orgId"`
		CredRef string `json:"credRef"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	authorizedGitOrgs, err := auth.GitClient.GetAuthorizedGithubRepos(request.OrgId, request.CredRef)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		GitOrgs []platformgit.GithubOrganization `json:"gitOrgs"`
	}{
		GitOrgs: authorizedGitOrgs,
	}), nil
}

func ObtainGithubToken(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		Code  string `json:"code"`
		OrgId string `json:"orgId"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	err := auth.GitClient.ObtainAccessToken(request.Code, request.OrgId)
	if err != nil {
		return server.Result{}, err
	}

	return server.Result{}, nil
}

func getCredentials(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId   string `json:"orgId"`
		OrgUuid string `json:"orgUuid"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	creds, err := auth.CredentialClient.GetCommonCredentials(request.OrgId, request.OrgUuid)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		Credentials []credentials.CredentialEntry `json:"credentials"`
	}{
		Credentials: creds,
	}), nil
}

func getCredentialDetails(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId        string `json:"orgId"`
		OrgUuid      string `json:"orgUuid"`
		CredentialId string `json:"credentialId"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	tokenResp, err := auth.CredentialClient.GetCommonCredentialsDetails(request.OrgId, request.OrgUuid, request.CredentialId)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(tokenResp), nil
}

func getRepoMetadata(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		OrgId        string `json:"orgId"`
		GitOrgName   string `json:"gitOrgName"`
		GitRepoName  string `json:"gitRepoName"`
		Branch       string `json:"branch"`
		RelativePath string `json:"relativePath"`
		SecretRef    string `json:"secretRef"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	metadata, err := auth.ComponentClient.GetRepoMetadata(request.OrgId, request.GitOrgName, request.GitRepoName, request.Branch, request.RelativePath, request.SecretRef)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(struct {
		Metadata component.RepoMetadataResponse `json:"metadata"`
	}{
		Metadata: metadata,
	}), nil
}

func gitTokenForRepository(req *json.RawMessage) (server.Result, error) {
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		return server.Result{}, api.ErrNotLoggedIn
	}

	var request struct {
		GitOrg    string `json:"gitOrg"`
		GitRepo   string `json:"gitRepo"`
		OrgId     string `json:"orgId"`
		SecretRef string `json:"secretRef"`
	}
	if err := json.Unmarshal(*req, &request); err != nil {
		return server.Result{}, err
	}

	tokenResp, err := auth.ComponentClient.GitTokenForRepository(request.GitOrg, request.GitRepo, request.OrgId, request.SecretRef)
	if err != nil {
		return server.Result{}, err
	}

	return server.CreateResult(tokenResp), nil
}
