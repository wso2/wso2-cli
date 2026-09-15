package platformgit

import "fmt"

// wrapper for gql query
func wrapQuery(query string) string {
	return `{"query": "` + query + `"}`
}

func GetAuthorizedGithubReposQuery(credRef string) string {
	return wrapQuery(fmt.Sprintf(`query {
		userRepos(secretRef: \"%s\") {
			orgName
			orgHandler
			repositories {
				name
			}
		}
	}`, credRef))
}

func ObtainUserTokenMutation(authCode string) string {
	query := fmt.Sprintf(`mutation {
		obtainUserToken(authorizationCode: \"%s\") {
			success
			message
		}
	}`, authCode)

	return wrapQuery(query)
}

func GetRepoBranchesQuery(repositoryOrganization, repositoryName, credRef string) string {
	query := fmt.Sprintf(`query {
		repoBranchList(repositoryOrganization: \"%s\", repositoryName: \"%s\", secretRef: \"%s\") {
			name
		}
	}`,
		repositoryOrganization,
		repositoryName,
		credRef,
	)

	return wrapQuery(query)
}

func GetCommitHistory(componentId string, branch string) string {
	query := fmt.Sprintf(`query {
		commitHistory(
			componentId: \"%s\"
			branch: \"%s\"
		) {
			message
			sha
			isLatest
			author {
                    name,
                    date,
                    email,
                    avatarUrl
            }
		}
	}`,
		componentId,
		branch,
	)

	return wrapQuery(query)
}
