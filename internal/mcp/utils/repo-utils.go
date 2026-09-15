package utils

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/util/repo"
)

var ValidGitUrlRegex = `(?:https?://|git@)(?:github\.com|gitlab\.com|bitbucket\.org)[:/]([^/]+)/([^/]+?)(?:\.git)?$`
var ValidComponentNameRegex = `^[a-z0-9-]{0,24}[a-z0-9]$`

func IsGHRepoValid(ctx context.Context, targetOrg *api.Organization, repoUrl string, branch string, relativePath string, isPublicRepo bool) (bool, error) {
	gitOrgName, gitRepoName, err := repo.ParseGitURL(repoUrl)
	if err != nil {
		return false, fmt.Errorf("failed to parse git URL: %w", err)
	}

	// The WSO2 Cloud GitHub App must have access to ALL repos — public and private alike —
	// because CICD needs it to create webhooks and workflow files regardless of visibility.
	MCPDebugf("IsGHRepoValid: checking GitHub App authorization for %s/%s", gitOrgName, gitRepoName)
	authorizedGitOrgs, err := GetGitClient(ctx).GetAuthorizedGithubRepos(targetOrg.ID, "")
	if err != nil {
		return false, fmt.Errorf("failed to get authorized github repos: %w", err)
	}
	MCPDebugf("IsGHRepoValid: fetched %d authorized GitHub orgs from platform", len(authorizedGitOrgs))

	isAuthorized := false
	for _, authorizedGitOrg := range authorizedGitOrgs {
		if gitOrgName != "" && (authorizedGitOrg.OrgName == gitOrgName || authorizedGitOrg.OrgHandler == gitOrgName) {
			for _, authorizedGitRepo := range authorizedGitOrg.Repositories {
				if authorizedGitRepo.Name == gitRepoName {
					isAuthorized = true
					break
				}
			}
		}
	}
	if !isAuthorized {
		MCPDebugf("IsGHRepoValid: repo %s/%s not found in authorized list — GitHub App not installed", gitOrgName, gitRepoName)
		return false, fmt.Errorf("%w: the WSO2 Cloud GitHub App does not have access to %s/%s — install it at %s",
			api.RepoAccessNeeded, gitOrgName, gitRepoName, auth.GetEnvConfig().GhApp.InstallUrl)
	}

	if !isPublicRepo {
		// Validate branch/subpath via platform. GetRepoMetadata requires GitHub App
		// credentials, so it is only viable after confirming App access above.
		metadata, err := GetComponentClient(ctx).GetRepoMetadata(targetOrg.ID, gitOrgName, gitRepoName, branch, relativePath, "")
		if err != nil {
			return false, fmt.Errorf("failed to get repo metadata for org: %s, repo: %s, branch: %s, path: %s, error: %w", gitOrgName, gitRepoName, branch, relativePath, err)
		}
		if metadata.IsSubPathEmpty || !metadata.IsSubPathValid {
			return false, fmt.Errorf("invalid subpath %s", relativePath)
		}
	}

	return true, nil
}

// HasComponentFile searches the entire directory tree for the required WSO2 Integration Platform manifest files.
// It directly checks for the full path, making it simple and robust.
func HasComponentOrEndpointsYaml(entries []component.PathEntry, componentDir string) bool {
	// Define the exact full paths we are looking for using cross-platform path joining
	targetComponentYaml := filepath.Join(componentDir, ".wso2", "component.yaml")
	targetEndpointsYaml := filepath.Join(componentDir, ".wso2", "endpoints.yaml")

	// search is a small, recursive function to traverse the tree
	var search func(items []component.PathEntry) bool
	search = func(items []component.PathEntry) bool {
		for _, item := range items {
			// If the item is a file and its path matches, we've found it.
			if item.Type == "blob" && (item.Path == targetComponentYaml || item.Path == targetEndpointsYaml) {
				return true
			}
			// If it's a directory, search its children.
			if len(item.Children) > 0 && search(item.Children) {
				return true
			}
		}
		return false // Not found in this branch
	}

	return search(entries)
}

// check if a repo is public
func IsGHRepoPublic(ctx context.Context, repoUrl string) (bool, error) {
	gitOrgName, gitRepoName, err := repo.ParseGitURL(repoUrl)
	if err != nil {
		return false, fmt.Errorf("failed to parse git URL: %w", err)
	}
	// Try to access the repo without authentication to check if it's public
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", gitOrgName, gitRepoName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Status 200 means repo exists and is public
	// Status 404 means repo either doesn't exist or is private
	return resp.StatusCode == http.StatusOK, nil
}
