package integrationtests_test

import (
	"os"
	"testing"

	"github.com/wso2/integration-platform-tools/internal/auth"
)

func TestPATEnvironmentVariableAuth(t *testing.T) {
	// Clear any existing authentication
	auth.ClearAuthStores()

	// Test case 1: No WSO2IP_PAT set - should not be logged in
	os.Unsetenv("WSO2IP_PAT")
	isLoggedIn := auth.IsLoggedIn()
	if isLoggedIn {
		t.Error("Expected not to be logged in when WSO2IP_PAT is not set")
	}

	// Test case 2: Empty WSO2IP_PAT - should not be logged in
	os.Setenv("WSO2IP_PAT", "")
	isLoggedIn = auth.IsLoggedIn()
	if isLoggedIn {
		t.Error("Expected not to be logged in when WSO2IP_PAT is empty")
	}

	// Test case 3: Invalid WSO2IP_PAT - should not be logged in
	os.Setenv("WSO2IP_PAT", "invalid-token")
	isLoggedIn = auth.IsLoggedIn()
	if isLoggedIn {
		t.Error("Expected not to be logged in when WSO2IP_PAT is invalid")
	}

	// Note: We don't test with a valid PAT here because:
	// 1. We don't want to expose real tokens in tests
	// 2. The PAT authentication involves external API calls
	// 3. The unit tests already cover the logic with invalid tokens

	// Clean up
	os.Unsetenv("WSO2IP_PAT")
	auth.ClearAuthStores()
}

func TestGetCurrentUserWithPATEnv(t *testing.T) {
	// Clear any existing authentication
	auth.ClearAuthStores()

	// Test case 1: No WSO2IP_PAT set - should return error
	os.Unsetenv("WSO2IP_PAT")
	userInfo, err := auth.GetCurrentUser()
	if err == nil {
		t.Error("Expected error when WSO2IP_PAT is not set and no stored user info")
	}
	if userInfo != nil {
		t.Error("Expected nil user info when WSO2IP_PAT is not set and no stored user info")
	}

	// Test case 2: Invalid WSO2IP_PAT - should return error
	os.Setenv("WSO2IP_PAT", "invalid-token")
	userInfo, err = auth.GetCurrentUser()
	if err == nil {
		t.Error("Expected error when WSO2IP_PAT is invalid")
	}
	if userInfo != nil {
		t.Error("Expected nil user info when WSO2IP_PAT is invalid")
	}

	// Clean up
	os.Unsetenv("WSO2IP_PAT")
	auth.ClearAuthStores()
}
