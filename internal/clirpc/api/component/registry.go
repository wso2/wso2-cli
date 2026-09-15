package component

import "github.com/wso2/integration-platform-tools/internal/clirpc/server"

func init() {
	server.RegisterHandler("component/createLink", CreateComponentLink)
	server.RegisterHandler("component/getList", GetComponents)
	server.RegisterHandler("component/getItem", GetComponentItem)
	server.RegisterHandler("component/getBuildPacks", GetBuildPacks)
	server.RegisterHandler("component/create", CreateComponent)
	server.RegisterHandler("component/delete", DeleteComponent)
	server.RegisterHandler("component/getDeploymentTracks", GetDeploymentTracks)
	server.RegisterHandler("component/getCommits", GetCommitHistory)
	server.RegisterHandler("component/getEndpoints", GetComponentEndpoints)
	server.RegisterHandler("component/getDeploymentStatus", GetComponentDeploymentStatus)
	server.RegisterHandler("component/createComponentConfig", CreateComponentConfig)
	server.RegisterHandler("component/updateCodeServer", updateCodeServer)
}
