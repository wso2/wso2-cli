package integration

import (
	"github.com/wso2/integration-platform-tools/pkg/api/component"
)

// Integration subtypes exposed by the MCP server. They are the platform's
// integration types, defined once in pkg/api/component and shared with the CLI
// so both surfaces offer the same six and map them the same way.
const (
	SubtypeAutomation       = component.IntegrationTypeAutomation
	SubtypeAPI              = component.IntegrationTypeAPI
	SubtypeAIAgent          = component.IntegrationTypeAIAgent
	SubtypeMCPServer        = component.IntegrationTypeMCPServer
	SubtypeEventIntegration = component.IntegrationTypeEventIntegration
	SubtypeFileIntegration  = component.IntegrationTypeFileIntegration
)

// Subtypes is the exhaustive, ordered list of integration subtypes the MCP
// server can create. No other platform component type is reachable through it.
var Subtypes = component.IntegrationTypes

var subtypeMappings = component.IntegrationTypeMappings

// ResolveSubtype maps an integration subtype to the underlying platform
// component type and component subtype to send on create. The match is exact:
// an agent passing an unrecognized or mis-cased value is told what the six are
// rather than having one guessed for it.
func ResolveSubtype(subtype string) (componentType string, componentSubType string, err error) {
	return component.ResolveIntegrationType(subtype)
}
