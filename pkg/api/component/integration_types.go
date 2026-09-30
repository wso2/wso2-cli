package component

import (
	"fmt"
	"strings"
)

// The integration types a user picks from. These are the product's own
// vocabulary: what someone creates is an Automation, an API, an AI Agent, an
// MCP Server, an Event Integration or a File Integration.
//
// On the wire each one is a component type plus, where several integration
// types share a component type, a component subtype. An AI Agent and an MCP
// Server are both `service` components; a File Integration is an
// `eventHandler`. Nothing outside this file should have to know that: callers
// resolve a type through ResolveIntegrationType and pass the result on.
const (
	IntegrationTypeAutomation       = "Automation"
	IntegrationTypeAPI              = "API"
	IntegrationTypeAIAgent          = "AI Agent"
	IntegrationTypeMCPServer        = "MCP Server"
	IntegrationTypeEventIntegration = "Event Integration"
	IntegrationTypeFileIntegration  = "File Integration"
)

// IntegrationTypes is the exhaustive, ordered list of integration types that
// can be created. No other platform component type is reachable through it.
var IntegrationTypes = []string{
	IntegrationTypeAutomation,
	IntegrationTypeAPI,
	IntegrationTypeAIAgent,
	IntegrationTypeMCPServer,
	IntegrationTypeEventIntegration,
	IntegrationTypeFileIntegration,
}

// IntegrationTypeMapping is the platform representation of one integration
// type. ComponentSubType is empty when the component type alone identifies it.
type IntegrationTypeMapping struct {
	ComponentType    string
	ComponentSubType string
}

var IntegrationTypeMappings = map[string]IntegrationTypeMapping{
	IntegrationTypeAutomation:       {ComponentTypeScheduledTask, ""},
	IntegrationTypeAPI:              {ComponentTypeService, ""},
	IntegrationTypeAIAgent:          {ComponentTypeService, ComponentSubTypeAiAgent},
	IntegrationTypeMCPServer:        {ComponentTypeService, ComponentSubTypeMCP},
	IntegrationTypeEventIntegration: {ComponentTypeEventHandler, ""},
	IntegrationTypeFileIntegration:  {ComponentTypeEventHandler, ComponentSubTypeFileIntegration},
}

// legacyComponentTypeSpellings maps the platform component types the CLI's
// --type flag used to take directly onto the integration type that now covers
// them. Existing scripts keep working; the subtype-bearing types (AI Agent,
// MCP Server, File Integration) were never reachable this way.
var legacyComponentTypeSpellings = map[string]string{
	ComponentTypeService:       IntegrationTypeAPI,
	ComponentTypeScheduledTask: IntegrationTypeAutomation,
	ComponentTypeEventHandler:  IntegrationTypeEventIntegration,
}

// ResolveIntegrationType maps an integration type to the component type and
// component subtype to send on create. The match is exact: this is the only
// gate between a caller and the backend's full component type vocabulary, so
// it does not guess at what an unrecognized value meant.
func ResolveIntegrationType(integrationType string) (componentType string, componentSubType string, err error) {
	mapping, ok := IntegrationTypeMappings[integrationType]
	if !ok {
		return "", "", fmt.Errorf(
			"unsupported integration type %q: must be one of %s",
			integrationType, strings.Join(IntegrationTypes, ", "),
		)
	}
	return mapping.ComponentType, mapping.ComponentSubType, nil
}

// MatchIntegrationType canonicalizes a value typed on a command line —
// "ai-agent", "mcp_server", "api" — into the integration type it names, and
// also accepts the platform component types the --type flag took before.
// Interfaces that must not be lenient (the MCP server) call
// ResolveIntegrationType directly instead.
func MatchIntegrationType(input string) (string, bool) {
	normalized := normalizeIntegrationType(input)
	if normalized == "" {
		return "", false
	}
	for _, integrationType := range IntegrationTypes {
		if normalizeIntegrationType(integrationType) == normalized {
			return integrationType, true
		}
	}
	for componentType, integrationType := range legacyComponentTypeSpellings {
		if normalizeIntegrationType(componentType) == normalized {
			return integrationType, true
		}
	}
	return "", false
}

// IsLegacyComponentTypeSpelling reports whether a value names a platform
// component type rather than an integration type, so a caller can say so
// before carrying on.
func IsLegacyComponentTypeSpelling(input string) bool {
	normalized := normalizeIntegrationType(input)
	for componentType := range legacyComponentTypeSpellings {
		if normalizeIntegrationType(componentType) == normalized {
			return true
		}
	}
	return false
}

// normalizeIntegrationType reduces a value to its letters and digits, lowercased,
// so that spacing, casing and separators do not matter on input.
func normalizeIntegrationType(value string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(value) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}
