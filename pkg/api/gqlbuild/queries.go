package gqlbuild

import "fmt"

func wrapQuery(query string) string {
	return `{"query": "` + query + `"}`
}

func GetBuildListQuery(orgHandler string, componentId string, versionId string) string {
	return wrapQuery(fmt.Sprintf(`
		query BuildsByVersion {
			buildsByVersion(orgHandler: \"%s\", build: { componentId: \"%s\", versionId: \"%s\" }) {
				id
				versionId
				buildId
				commitHash
				componentId
				createdDate
				commitMessage
			}
		}
	`, orgHandler, componentId, versionId))
}

func GetBuildStepLogQuery(componentId string, runId int, logType string) string {
	return wrapQuery(fmt.Sprintf(`
        query BuildLogs{
            buildLogs(componentId:\"%s\", runId:\"%d\"){
               %s 
            }
        }
	`, componentId, runId, logType))
}

func GetScanResultQuery(cmpId, branch string) string {
	return wrapQuery(fmt.Sprintf(`
        query ScanResult{
            scanResult(componentId:\"%s\", branch:\"%s\"){
                trivyScan,
                checkovScan
            }
        }
	`, cmpId, branch))
}

func GetDeploymentStatusByVersionQuery(versionId string, componentId string) string {
	return wrapQuery(fmt.Sprintf(`
		query DeploymentStatusByVersion {
			deploymentStatusByVersion(versionId: \"%s\", componentId: \"%s\") {
				id
                sha
                completed_at
                started_at
                name
                status
                conclusion
                isAutoDeploy
                failureReason
                sourceCommitId
			}
		}	
	`, versionId, componentId))
}

func GetCreateProxyBuild(cmpId, commitHash, apiId string) string {
	return wrapQuery(fmt.Sprintf(`mutation TriggerProxyBuild {
        triggerProxyBuild(componentID: \"%s\", commitHash: \"%s\", apiId: \"%s\")    
    }`, cmpId, commitHash, apiId))
}
