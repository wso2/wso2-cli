package deploymentbuild

import "fmt"

func wrapQuery(query string) string {
	return `{"query": "` + query + `"}`
}

func getProxyDeploymentQuery(orgHandler, orgUuid, cmpId, versionId, envId string) string {
	return wrapQuery(fmt.Sprintf(`
        query {
            proxyDeployment(
                orgHandler: \"%s\"
                orgUuid:\"%s\"
                componentId: \"%s\"
                versionId: \"%s\"
                environmentId: \"%s\"
            ) {
                apiId,
                environment {
                    choreoEnv,
                    name,
                    id
                },
                lifecycleStatus,
                version,
                invokeUrl,
                endpoint,
                sandboxEndpoint,
                apiRevision {
                    id,
                    displayName,
                    createdTime
                },
                build {
                    id
                    baseRevisionId
                    deployedRevisionId
                },
                deployedTime,
                successDeployedTime
            }
        }`, orgHandler, orgUuid, cmpId, versionId, envId))
}

func getComponentInsightsQuery(fFrom, fTo, orgId, envs, apiId string) string {
	return fmt.Sprintf(`
        {
            "operationName": "componentInsights",
            "variables": {
                "filter": {
                    "from": "%s",
                    "to": "%s"
                },
                "dataFilter": {
                    "orgId": "%s",
                    "environmentIds": [
                        %s
                    ],
                    "tenant": "carbon.super"
                },
                "apiId": "%s"
            },
            "query": "query componentInsights($dataFilter: DataFilter!, $filter: TimeFilter!, $apiId: ID!) {\n  getTotalTrafficByAPI(filter: $filter, dataFilter: $dataFilter, apiId: $apiId)\n  getOverallLatencyByAPI(filter: $filter, dataFilter: $dataFilter, apiId: $apiId) {\n    response\n    __typename\n  }\n  getTotalErrorsByAPI(filter: $filter, dataFilter: $dataFilter, apiId: $apiId) {\n    proxy\n    __typename\n  }\n}\n"
        }
    `, fFrom, fTo, orgId, envs, apiId)
}

func GetSuspendDeploymentQuery(orgHandler, componentId, releaseId, componentType string) string {
	return wrapQuery(fmt.Sprintf(`mutation {
      stopDeployment(
        orgHandler: \"%s\",
        componentId: \"%s\",
        releaseId: \"%s\",
        type: \"%s\"
      )
    }`, orgHandler, componentId, releaseId, componentType))
}

func GetResumeDeploymentQuery(orgHandler, componentId, releaseId, cmpType string) string {
	return wrapQuery(fmt.Sprintf(`mutation {
      redeployDeployment(
        orgHandler: \"%s\",
        componentId: \"%s\",
        releaseId: \"%s\",
        type: \"%s\"
      )
    }`, orgHandler, componentId, releaseId, cmpType))
}
