package deployment

import "github.com/wso2/integration-platform-tools/internal/clirpc/server"

func init() {
	server.RegisterHandler("connections/getMarketplaceItems", getMarketplaceItems)
	server.RegisterHandler("connections/getMarketplaceItemIdl", getMarketplaceItemIdl)
	server.RegisterHandler("connections/createComponentConnection", createComponentConnection)
	server.RegisterHandler("connections/getConnections", getConnections)
	server.RegisterHandler("connections/getConnectionItem", GetConnectionItem)
	server.RegisterHandler("connections/deleteConnection", deleteConnection)
	server.RegisterHandler("connections/getGuide", getConnectionGuide)
}
