package build

import "github.com/wso2/integration-platform-tools/internal/clirpc/server"

func init() {
	server.RegisterHandler("build/getList", GetBuildList)
	server.RegisterHandler("build/create", CreateBuild)
	server.RegisterHandler("build/logs", GetBuildLogs)
	server.RegisterHandler("build/getLogsForType", GetBuildLogsForType)
	server.RegisterHandler("build/getAutoBuildStatus", GetAutoBuildStatus)
	server.RegisterHandler("build/enableAutoBuild", EnableAutoBuild)
	server.RegisterHandler("build/disableAutoBuild", DisableAutoBuild)
}
