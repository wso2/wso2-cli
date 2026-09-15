package repo

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/wso2/integration-platform-tools/pkg/api/platformgit"
)

// Get the root directory of the git repository
func GetGitRepoRoot(path string) (string, error) {
	// Traverse upwards to find the root of the git directory
	for {
		repo, err := git.PlainOpen(path)
		if err == nil {
			worktree, err := repo.Worktree()
			if err == nil {
				rootDir := worktree.Filesystem.Root()
				return rootDir, nil
			}
		}

		// If we've reached the root directory and still haven't found a git repository, return an error
		if path == filepath.Dir(path) {
			return "", err
		}

		// Traverse upwards
		path = filepath.Dir(path)
	}
}

func GetGitRepoRelative(repoRoot string, componentPath string) (string, error) {
	subPath, err := filepath.Rel(repoRoot, componentPath)
	if strings.HasPrefix(subPath, ".."+string(filepath.Separator)) {
		return "", errors.New("invalid component path")
	}
	if err != nil {
		return "", err
	}
	return subPath, nil
}

func ParseGitURL(url string) (string, string, error) {
	// Assuming the URL format is "https://github.com/org/repo.git" or "git@github.com:org/repo.git"
	var org, repoName string
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		parts := strings.Split(url, "/")
		if len(parts) < 2 {
			return "", "", errors.New("invalid Git URL")
		}
		org = parts[len(parts)-2]
		repoName = strings.TrimSuffix(parts[len(parts)-1], ".git")
	} else if strings.HasPrefix(url, "git@") {
		parts := strings.Split(url, ":")
		if len(parts) < 2 {
			return "", "", errors.New("invalid Git URL")
		}
		orgRepo := strings.Split(parts[1], "/")
		if len(orgRepo) < 2 {
			return "", "", errors.New("invalid Git URL")
		}
		org = orgRepo[0]
		repoName = strings.TrimSuffix(orgRepo[1], ".git")
	} else {
		return "", "", errors.New("invalid Git URL")
	}
	return org, repoName, nil
}

// Check if given repo url exists within repos/orgs authorized to be accessed by the platform
func IsGitRepoAuthorized(
	authorizedGitOrgs []platformgit.GithubOrganization,
	repoUrl string) (gitRepoName string, gitOrgName string, err error) {

	gitOrgName, gitRepoName, err = ParseGitURL(repoUrl)
	if err != nil {
		return
	}

	if err != nil {
		return
	}

	var gitOrg platformgit.GithubOrganization
	orgFound := false
	for _, org := range authorizedGitOrgs {
		if org.OrgName == gitOrgName {
			gitOrg = org
			orgFound = true
			break
		}
	}

	if !orgFound {
		err = fmt.Errorf("organization '%s' is not authorized by WSO2 Integration Platform", gitOrgName)
		return
	}

	var foundRepo bool
	for _, repo := range gitOrg.Repositories {
		if repo.Name == gitRepoName {
			foundRepo = true
			break
		}
	}
	if !foundRepo {
		err = fmt.Errorf("repository '%s' is not authorized by WSO2 Integration Platform", gitRepoName)
		return
	}

	return gitRepoName, gitOrgName, nil
}

func GetGitRemotesNames(dirPath string) ([]string, error) {
	if dirPath == "" {
		return nil, fmt.Errorf("repo path not found")
	}
	// Open the current repository
	repo, err := git.PlainOpen(dirPath)
	if err != nil {
		return nil, err
	}

	// Get all remotes
	remotes, err := repo.Remotes()
	if err != nil {
		return nil, err
	}

	remoteNames := []string{}

	for _, remote := range remotes {
		for _, item := range remote.Config().URLs {
			sanitizedUrl, err := removeCredentialsFromGitURL(item)
			if err != nil {
				return nil, err
			}
			if sanitizedUrl != "" {
				remoteNames = append(remoteNames, sanitizedUrl)
			}
		}
	}

	return remoteNames, nil
}
func removeCredentialsFromGitURL(gitURL string) (string, error) {
	if strings.HasPrefix(gitURL, "git@") {
		// ssh url
		parts := strings.Split(strings.TrimPrefix(gitURL, "git@"), ":")
		if len(parts) == 2 {
			host := parts[0]
			path := strings.TrimSuffix(parts[1], ".git")
			httpsURL := fmt.Sprintf("https://%s/%s", host, strings.TrimSuffix(path, ".git"))
			return httpsURL, nil
		}
	} else {
		// http/https url
		parsedURL, err := url.Parse(gitURL)
		if err != nil {
			return "", err
		}

		// Split the user info from the URL
		userInfo := parsedURL.User

		// If user info is present, remove it
		if userInfo != nil {
			// Remove only the user info, keep the hostname
			parsedURL.User = nil
		}

		// Convert the URL back to string
		redactedURL := parsedURL.String()

		return strings.TrimSuffix(redactedURL, ".git"), nil
	}
	return "", fmt.Errorf("failed to parse git repository url")
}

func GetDefaultBranch(branchNames []string) string {
	for _, b := range branchNames {
		if b == "main" {
			return "main"
		} else if b == "master" {
			return "master"
		}
	}
	if len(branchNames) > 0 {
		return branchNames[0]
	}
	return ""
}

func AddToGitIgnore(ditIgnoreDir string, ignoredFile string) error {
	gitignorePath := filepath.Join(ditIgnoreDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		// create new .gitignore file
		f, err := os.Create(gitignorePath)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := f.WriteString(ignoredFile + "\n"); err != nil {
			return err
		}
	} else {
		// check if path already exists in .gitignore
		f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)

		// check if the existing gitignore already contains ignoredFile
		var alreadyHasIgnored bool
		for scanner.Scan() {
			if scanner.Text() == ignoredFile {
				alreadyHasIgnored = true
				break
			}
		}

		if !alreadyHasIgnored {
			if _, err := f.WriteString("\n" + ignoredFile + "\n"); err != nil {
				return err
			}
		}
	}
	return nil
}
