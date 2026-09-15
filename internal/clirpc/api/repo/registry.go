package repo

import "github.com/wso2/integration-platform-tools/internal/clirpc/server"

func init() {
	server.RegisterHandler("repo/getBranches", GetRepoBranches)
	server.RegisterHandler("repo/isRepoAuthorized", IsGitRepoAuthorized)
	server.RegisterHandler("repo/getAuthorizedGitOrgs", GetAuthorizedGitOrgs)
	server.RegisterHandler("repo/obtainGithubToken", ObtainGithubToken)
	server.RegisterHandler("repo/getCredentials", getCredentials)
	server.RegisterHandler("repo/getCredentialDetails", getCredentialDetails)
	server.RegisterHandler("repo/getRepoMetadata", getRepoMetadata)
	server.RegisterHandler("repo/gitTokenForRepository", gitTokenForRepository)
}
