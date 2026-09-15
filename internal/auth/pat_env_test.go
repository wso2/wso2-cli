package auth

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTryAuthenticateWithEnvPAT(t *testing.T) {
	// Test case 1: No WSO2IP_PAT environment variable set
	os.Unsetenv("WSO2IP_PAT")
	result := TryAuthenticateWithEnvPAT()
	assert.False(t, result, "Should return false when WSO2IP_PAT is not set")

	// Test case 2: Empty WSO2IP_PAT environment variable
	os.Setenv("WSO2IP_PAT", "")
	result = TryAuthenticateWithEnvPAT()
	assert.False(t, result, "Should return false when WSO2IP_PAT is empty")

	// Test case 3: Invalid PAT token (this will fail authentication)
	os.Setenv("WSO2IP_PAT", "invalid-token")
	result = TryAuthenticateWithEnvPAT()
	assert.False(t, result, "Should return false when PAT is invalid")

	// Clean up
	os.Unsetenv("WSO2IP_PAT")
}

func TestIsLoggedInWithEnvPAT(t *testing.T) {
	// Clear any existing authentication
	ClearAuthStores()

	// Test case 1: No stored authentication and no WSO2IP_PAT
	os.Unsetenv("WSO2IP_PAT")
	result := IsLoggedIn()
	assert.False(t, result, "Should return false when no stored auth and no WSO2IP_PAT")

	// Test case 2: No stored authentication but invalid WSO2IP_PAT
	os.Setenv("WSO2IP_PAT", "invalid-token")
	result = IsLoggedIn()
	assert.False(t, result, "Should return false when WSO2IP_PAT is invalid")

	// Clean up
	os.Unsetenv("WSO2IP_PAT")
	ClearAuthStores()
}

func TestGetCurrentUserWithEnvPAT(t *testing.T) {
	// Clear any existing authentication
	ClearAuthStores()

	// Test case 1: No stored user info and no WSO2IP_PAT
	os.Unsetenv("WSO2IP_PAT")
	userInfo, err := GetCurrentUser()
	assert.Error(t, err, "Should return error when no stored user info and no WSO2IP_PAT")
	assert.Nil(t, userInfo, "Should return nil user info when no stored user info and no WSO2IP_PAT")

	// Test case 2: No stored user info but invalid WSO2IP_PAT
	os.Setenv("WSO2IP_PAT", "invalid-token")
	userInfo, err = GetCurrentUser()
	assert.Error(t, err, "Should return error when WSO2IP_PAT is invalid")
	assert.Nil(t, userInfo, "Should return nil user info when WSO2IP_PAT is invalid")

	// Clean up
	os.Unsetenv("WSO2IP_PAT")
	ClearAuthStores()
}
