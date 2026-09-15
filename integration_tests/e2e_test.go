package integrationtests_test

import (
	"testing"
)

func TestE2EMainFlow(t *testing.T) {
	if TEST_USER_NAME == "" || TEST_USER_PASS == "" {
		t.Fatal("CHOREO_CLI_TEST_USER_NAME and CHOREO_CLI_TEST_USER_PASS environment variables are required")
	}

	// Note: Non-interactive mode is enabled globally in TestMain (shared_test.go)
	// This ensures all tests run without interactive prompts and fail fast on missing parameters
	t.Run("Test Non-interactive mode", TestNonInteractiveMode)

	t.Run("Test Login", TestLogin)

	t.Run("Test Create Project", TestCreateProject)

	t.Run("Test List Components empty", TestListComponentsEmpty)

	// Component create/build/deploy coverage was removed along with the nodejs
	// webApp tests — the Integration Platform has no web-app component type.
	// See the note in component_test.go before re-adding E2E coverage.

	t.Run("Test Delete Project", TestDeleteProject)
}
