package deployment

import "github.com/wso2/integration-platform-tools/internal/clirpc/server"

func init() {
	server.RegisterHandler("deployment/create", CreateDeployment)
	server.RegisterHandler("deployment/getProxyDeploymentInfo", GetProxyDeploymentInfo)
	server.RegisterHandler("deployment/checkWorkflowStatus", CheckWorkflowStatus)
	server.RegisterHandler("deployment/promoteProxy", PromoteProxy)
	server.RegisterHandler("deployment/requestPromoteApproval", RequestPromoteApproval)
	server.RegisterHandler("deployment/cancelApprovalRequest", CancelApprovalRequest)
}
