package customDomain

import (
	"strings"
	"testing"
)

// Custom domains and URL mappings only apply to integrations that serve HTTP.
// Both tools accept a single component_type value, 'api', which covers the
// 'API', 'AI Agent' and 'MCP Server' subtypes — that mapping is not obvious
// from the value alone, so it has to be spelled out.
func TestDomainToolsDocumentEndpointOnlyScope(t *testing.T) {
	tools := map[string]string{
		"register_custom_domain": RegisterCustomDomainTool.Description,
		"create_url_mapping":     CreateUrlMappingTool.Description,
	}

	for name, desc := range tools {
		t.Run(name, func(t *testing.T) {
			if !strings.Contains(desc, "APPLIES ONLY TO") {
				t.Error("description must state that the tool is restricted")
			}
			for _, applicable := range []string{"'API'", "'AI Agent'", "'MCP Server'"} {
				if !strings.Contains(desc, applicable) {
					t.Errorf("description must name %s as applicable", applicable)
				}
			}
			for _, excluded := range []string{"Automation", "Event Integration", "File Integration"} {
				if !strings.Contains(desc, excluded) {
					t.Errorf("description must name %q as not applicable", excluded)
				}
			}
			// The single enum value covering three subtypes is the confusing part.
			if !strings.Contains(desc, "'api'") {
				t.Error("description must explain that component_type 'api' covers all three subtypes")
			}
		})
	}
}

// The component_type argument accepts only "api"; both tools must name it, since
// an agent may otherwise try to pass a subtype name like "Automation".
func TestComponentTypeArgumentIsDocumented(t *testing.T) {
	for name, desc := range map[string]string{
		"register_custom_domain": RegisterCustomDomainTool.Description,
		"create_url_mapping":     CreateUrlMappingTool.Description,
	} {
		if !strings.Contains(desc, "component_type") {
			t.Errorf("%s: description should reference the component_type value", name)
		}
	}
}
