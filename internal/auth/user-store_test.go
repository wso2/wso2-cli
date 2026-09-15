package auth

import (
	"testing"

	"github.com/wso2/integration-platform-tools/pkg/api"
)

func TestUserStore(t *testing.T) {
	us := NewUserStore()

	// Test StoreUserInfo and RetrieveUserInfo
	userInfo := &api.UserInfo{
		DisplayName:   "test-user",
		Organizations: []api.Organization{{ID: "test-org", Name: "Test Organization"}},
	}

	err := us.StoreUserInfo(*userInfo)
	if err != nil {
		t.Error("StoreUserInfo failed")
	}

	retrievedUserInfo, err := us.RetrieveUserInfo()
	if err != nil {
		t.Error("RetrieveUserInfo failed")
	}

	if userInfo.DisplayName != retrievedUserInfo.DisplayName {
		t.Error("Retrieved user info does not match stored user info")
	}

	// Test ClearUserInfo
	err = us.ClearUserInfo()
	if err != nil {
		t.Error("ClearUserInfo failed")
	}

	_, err = us.RetrieveUserInfo()

	if err == nil {
		t.Error("RetrieveUserInfo should fail after ClearUserInfo")
	}
}
