package auth

import (
	"testing"

	"github.com/wso2/integration-platform-tools/pkg/api"
)

func TestOrgStore(t *testing.T) {
	orgStore := NewOrgStore()
	orgInfo := &api.Organization{ID: "test-org", Name: "Test Organization"}

	// Test StoreOrganizationInfo and RetrieveOrganizationInfo
	err := orgStore.SetDefaultOrg(orgInfo)

	if err != nil {
		t.Error("StoreOrganizationInfo failed")
	}

	retrievedOrgInfo, err := orgStore.GetDefaultOrg()

	if err != nil {
		t.Error("RetrieveOrganizationInfo failed")
	}

	if orgInfo.ID != retrievedOrgInfo.ID {
		t.Error("Retrieved organization info does not match stored organization info")
	}

	if orgInfo.Name != retrievedOrgInfo.Name {
		t.Error("Retrieved organization info does not match stored organization info")
	}

	// delete the organization info
	err = orgStore.RemoveDefaultOrg()

	if err != nil {
		t.Error("DeleteOrganizationInfo failed")
	}

	info, err := orgStore.GetDefaultOrg()

	if err == nil {
		t.Error("RetrieveOrganizationInfo should fail after DeleteOrganizationInfo")
	}

	if info != nil {
		t.Error("RetrieveOrganizationInfo should return nil after DeleteOrganizationInfo")
	}
}
