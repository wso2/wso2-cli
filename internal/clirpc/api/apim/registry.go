package apim

import "github.com/wso2/integration-platform-tools/internal/clirpc/server"

func init() {
	server.RegisterHandler("apim/getTestKey", GetEndpointTestKey)
	server.RegisterHandler("apim/getSwaggerSpec", GetSwaggerSpec)
}
