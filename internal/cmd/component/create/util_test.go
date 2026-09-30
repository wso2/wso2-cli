package create

import (
	"testing"

	"github.com/wso2/integration-platform-tools/pkg/api/component"
)

func newTestCreateParams() *CreateComponentParams {
	return &CreateComponentParams{
		ComponentName: "test-integration",
		ComponentType: component.ComponentTypeService,
		BuildPack:     component.ComponentBuildPackBallerina,
		RepoBranch:    "main",
	}
}

func TestGetComponentKindForCreate_AutoBuildDeployDefaultTrue(t *testing.T) {
	got, err := GetComponentKindForCreate(newTestCreateParams(), "test-project", "org-id", "org-uuid", "https://github.com/org/repo.git")
	if err != nil {
		t.Fatalf("GetComponentKindForCreate returned unexpected error: %v", err)
	}

	if !got.Spec.Build.EnableAutoBuild {
		t.Errorf("EnableAutoBuild = false, want true when AutoBuild is unset")
	}
	if !got.Spec.Build.EnableAutoDeploy {
		t.Errorf("EnableAutoDeploy = false, want true when AutoDeploy is unset")
	}
}

func TestGetComponentKindForCreate_AutoBuildDeployOverride(t *testing.T) {
	falseVal := false
	params := newTestCreateParams()
	params.AutoBuild = &falseVal
	params.AutoDeploy = &falseVal

	got, err := GetComponentKindForCreate(params, "test-project", "org-id", "org-uuid", "https://github.com/org/repo.git")
	if err != nil {
		t.Fatalf("GetComponentKindForCreate returned unexpected error: %v", err)
	}

	if got.Spec.Build.EnableAutoBuild {
		t.Errorf("EnableAutoBuild = true, want false (explicit override)")
	}
	if got.Spec.Build.EnableAutoDeploy {
		t.Errorf("EnableAutoDeploy = true, want false (explicit override)")
	}
}

func TestGetComponentKindForCreate_AutoBuildDeployPartialOverride(t *testing.T) {
	falseVal := false
	params := newTestCreateParams()
	params.AutoBuild = &falseVal // AutoDeploy left unset, should still default true

	got, err := GetComponentKindForCreate(params, "test-project", "org-id", "org-uuid", "https://github.com/org/repo.git")
	if err != nil {
		t.Fatalf("GetComponentKindForCreate returned unexpected error: %v", err)
	}

	if got.Spec.Build.EnableAutoBuild {
		t.Errorf("EnableAutoBuild = true, want false (explicit override)")
	}
	if !got.Spec.Build.EnableAutoDeploy {
		t.Errorf("EnableAutoDeploy = false, want true (left unset)")
	}
}

// The --type flag and the interactive prompt both speak the product's
// vocabulary; everything downstream of resolveIntegrationType speaks the
// platform's. An AI Agent and an MCP Server are services, and a File
// Integration is an event handler, so the subtype has to survive the trip.
func TestResolveIntegrationType(t *testing.T) {
	cases := []struct {
		input       string
		wantType    string
		wantSubType string
	}{
		{"API", component.ComponentTypeService, ""},
		{"AI Agent", component.ComponentTypeService, component.ComponentSubTypeAiAgent},
		{"mcp-server", component.ComponentTypeService, component.ComponentSubTypeMCP},
		{"Automation", component.ComponentTypeScheduledTask, ""},
		{"Event Integration", component.ComponentTypeEventHandler, ""},
		{"file integration", component.ComponentTypeEventHandler, component.ComponentSubTypeFileIntegration},

		// The component types --type took before this command spoke in
		// integration types.
		{component.ComponentTypeService, component.ComponentTypeService, ""},
		{component.ComponentTypeScheduledTask, component.ComponentTypeScheduledTask, ""},
		{component.ComponentTypeEventHandler, component.ComponentTypeEventHandler, ""},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			params := newTestCreateParams()
			params.ComponentType = tc.input

			if err := resolveIntegrationType(params); err != nil {
				t.Fatalf("resolveIntegrationType(%q) returned unexpected error: %v", tc.input, err)
			}
			if params.ComponentType != tc.wantType {
				t.Errorf("ComponentType = %q, want %q", params.ComponentType, tc.wantType)
			}
			if params.ComponentSubType != tc.wantSubType {
				t.Errorf("ComponentSubType = %q, want %q", params.ComponentSubType, tc.wantSubType)
			}
		})
	}
}

func TestResolveIntegrationTypeRejectsUnknownType(t *testing.T) {
	for _, input := range []string{"webApp", "proxy", "nonsense", ""} {
		t.Run(input, func(t *testing.T) {
			params := newTestCreateParams()
			params.ComponentType = input

			if err := resolveIntegrationType(params); err == nil {
				t.Errorf("resolveIntegrationType(%q) expected an error, got nil", input)
			}
		})
	}
}

// A File Integration is sent as the buildpack-specific subtype the platform
// matches on, not the generic one the user picked.
func TestGetComponentKindForCreate_FileIntegrationSubtypeByBuildpack(t *testing.T) {
	for _, tc := range []struct {
		buildPack   string
		wantSubType string
	}{
		{component.ComponentBuildPackBallerina, "ballerinaFileIntegration"},
		{component.ComponentBuildPackMI, "miFileIntegration"},
	} {
		t.Run(tc.buildPack, func(t *testing.T) {
			params := newTestCreateParams()
			params.ComponentType = component.IntegrationTypeFileIntegration
			params.BuildPack = tc.buildPack

			if err := resolveIntegrationType(params); err != nil {
				t.Fatalf("resolveIntegrationType returned unexpected error: %v", err)
			}

			got, err := GetComponentKindForCreate(params, "test-project", "org-id", "org-uuid", "https://github.com/org/repo.git")
			if err != nil {
				t.Fatalf("GetComponentKindForCreate returned unexpected error: %v", err)
			}
			if got.Spec.SubType != tc.wantSubType {
				t.Errorf("SubType = %q, want %q", got.Spec.SubType, tc.wantSubType)
			}
		})
	}
}
