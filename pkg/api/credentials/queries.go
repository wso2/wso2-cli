package credentials

import "fmt"

func wrapQuery(query string) string {
	return `{"query": "` + query + `"}`
}

func getCommonCredentialQuery(orgUUID string) string {
	return wrapQuery(fmt.Sprintf(
		`query { 
            commonCredentials( orgUuid: \"%s\") {
                id, name, createdAt, organizationUuid, type, referenceToken
            }
        }`,
		orgUUID,
	))
}

func getCommonCredentialDetailsQuery(orgUuid, credentialId string) string {
	return wrapQuery(fmt.Sprintf(`query {
      commonCredential(
        orgUuid: \"%s\",
        credentialId: \"%s\",
      ) {
        id,
        name,
        type,
        createdAt,
        organizationUuid,
        referenceToken,
        serverUrl
      }
    }`,
		orgUuid,
		credentialId))
}
