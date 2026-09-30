package component

import "testing"

func TestResolveIntegrationType(t *testing.T) {
	cases := []struct {
		integrationType string
		wantType        string
		wantSubType     string
	}{
		{IntegrationTypeAutomation, ComponentTypeScheduledTask, ""},
		{IntegrationTypeAPI, ComponentTypeService, ""},
		{IntegrationTypeAIAgent, ComponentTypeService, ComponentSubTypeAiAgent},
		{IntegrationTypeMCPServer, ComponentTypeService, ComponentSubTypeMCP},
		{IntegrationTypeEventIntegration, ComponentTypeEventHandler, ""},
		{IntegrationTypeFileIntegration, ComponentTypeEventHandler, ComponentSubTypeFileIntegration},
	}

	for _, tc := range cases {
		t.Run(tc.integrationType, func(t *testing.T) {
			gotType, gotSubType, err := ResolveIntegrationType(tc.integrationType)
			if err != nil {
				t.Fatalf("ResolveIntegrationType(%q) returned unexpected error: %v", tc.integrationType, err)
			}
			if gotType != tc.wantType {
				t.Errorf("componentType = %q, want %q", gotType, tc.wantType)
			}
			if gotSubType != tc.wantSubType {
				t.Errorf("componentSubType = %q, want %q", gotSubType, tc.wantSubType)
			}
		})
	}
}

// ResolveIntegrationType is the only gate between a caller and the backend's
// full component type vocabulary, so it matches exactly and rejects everything
// else -- including the component types the CLI's --type flag used to take,
// which reach it through MatchIntegrationType instead.
func TestResolveIntegrationTypeRejectsUnsupportedValues(t *testing.T) {
	unsupported := []string{
		ComponentTypeByocWebApp,
		ComponentTypeManualTrigger,
		ComponentTypeWebhook,
		ComponentTypeProxyGH,
		ComponentTypeService,
		"api", // wrong case
		"",
	}
	for _, integrationType := range unsupported {
		t.Run(integrationType, func(t *testing.T) {
			if _, _, err := ResolveIntegrationType(integrationType); err == nil {
				t.Errorf("ResolveIntegrationType(%q) expected an error, got nil", integrationType)
			}
		})
	}
}

// What a person types on a command line carries whatever casing and separators
// they reached for, and the three component types --type accepted before must
// keep resolving so existing scripts do not break.
func TestMatchIntegrationType(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"API", IntegrationTypeAPI},
		{"api", IntegrationTypeAPI},
		{"AI Agent", IntegrationTypeAIAgent},
		{"ai-agent", IntegrationTypeAIAgent},
		{"aiagent", IntegrationTypeAIAgent},
		{"mcp_server", IntegrationTypeMCPServer},
		{"MCP SERVER", IntegrationTypeMCPServer},
		{"file integration", IntegrationTypeFileIntegration},
		{"Event-Integration", IntegrationTypeEventIntegration},
		{"automation", IntegrationTypeAutomation},

		// The pre-rename spellings of --type.
		{ComponentTypeService, IntegrationTypeAPI},
		{ComponentTypeScheduledTask, IntegrationTypeAutomation},
		{ComponentTypeEventHandler, IntegrationTypeEventIntegration},
		{"scheduletask", IntegrationTypeAutomation},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, ok := MatchIntegrationType(tc.input)
			if !ok {
				t.Fatalf("MatchIntegrationType(%q) did not match", tc.input)
			}
			if got != tc.want {
				t.Errorf("MatchIntegrationType(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// Leniency about spelling must not become leniency about which component types
// are reachable: the ones this product does not create stay out.
func TestMatchIntegrationTypeRejectsNonIntegrations(t *testing.T) {
	for _, input := range []string{
		ComponentTypeByocWebApp,
		ComponentTypeManualTrigger,
		ComponentTypeWebhook,
		ComponentTypeProxyGH,
		"web app",
		"",
		"   ",
	} {
		t.Run(input, func(t *testing.T) {
			if got, ok := MatchIntegrationType(input); ok {
				t.Errorf("MatchIntegrationType(%q) matched %q, want no match", input, got)
			}
		})
	}
}

func TestIsLegacyComponentTypeSpelling(t *testing.T) {
	for _, input := range []string{ComponentTypeService, ComponentTypeScheduledTask, ComponentTypeEventHandler} {
		if !IsLegacyComponentTypeSpelling(input) {
			t.Errorf("IsLegacyComponentTypeSpelling(%q) = false, want true", input)
		}
	}
	for _, input := range IntegrationTypes {
		if IsLegacyComponentTypeSpelling(input) {
			t.Errorf("IsLegacyComponentTypeSpelling(%q) = true, want false", input)
		}
	}
}

func TestIntegrationTypesListMatchesMappings(t *testing.T) {
	if len(IntegrationTypes) != len(IntegrationTypeMappings) {
		t.Fatalf("IntegrationTypes has %d entries, IntegrationTypeMappings has %d",
			len(IntegrationTypes), len(IntegrationTypeMappings))
	}
	for _, integrationType := range IntegrationTypes {
		if _, ok := IntegrationTypeMappings[integrationType]; !ok {
			t.Errorf("IntegrationTypes contains %q which has no mapping", integrationType)
		}
	}
}

// IntegrationKind reports what a component already on the platform is;
// IntegrationTypes is what can be created. The two name the same things, so a
// kind that is not a creatable type would be a vocabulary split.
func TestIntegrationKindNamesAreCreatableTypes(t *testing.T) {
	kinds := []string{
		IntegrationKind(DisplayTypeService, ComponentSubTypeAiAgent),
		IntegrationKind(DisplayTypeService, ComponentSubTypeMCP),
		IntegrationKind(DisplayTypeBallerinaFileIntegration, ""),
		IntegrationKind(DisplayTypeScheduledTask, ""),
		IntegrationKind(DisplayTypeBallerinaEventHandler, ""),
		IntegrationKind(DisplayTypeService, ""),
	}
	for _, kind := range kinds {
		if _, ok := IntegrationTypeMappings[kind]; !ok {
			t.Errorf("IntegrationKind returned %q, which is not one of IntegrationTypes", kind)
		}
	}
}
