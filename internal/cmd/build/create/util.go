package create

import (
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/utils"
	deploymentbuild "github.com/wso2/integration-platform-tools/pkg/api/deploymentBuild"
)

func triggerComponentBuild(orgId string, componentName string, projectHandle string, deploymentTrackId string, commitHash string) (deploymentBuild *deploymentbuild.BuildKind, err error) {
	deploymentBuildSpinner := utils.CreateSpinner(" Triggering component build...", "")
	deploymentBuildSpinner.Start()
	deploymentBuildRes, err := auth.DeploymentBuildClient.CreateDeploymentBuilds(orgId, componentName, projectHandle, deploymentTrackId, commitHash)
	deploymentBuildSpinner.Stop()
	if err != nil {
		return nil, err
	}
	return &deploymentBuildRes, nil
}
