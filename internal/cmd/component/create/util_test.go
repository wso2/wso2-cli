package create

import (
	"testing"

	"github.com/wso2/integration-platform-tools/pkg/api/component"
)

func newTestCreateParams() *CreateComponentParams {
	return &CreateComponentParams{
		ComponentName: "test-component",
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
