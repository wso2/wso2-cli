package project

import "github.com/wso2/integration-platform-tools/internal/clirpc/server"

func init() {
	server.RegisterHandler("project/getProjects", GetProjectsByOrgID)
	server.RegisterHandler("project/create", CreateProject)
	server.RegisterHandler("project/getEnvs", GetEnvs)
}
