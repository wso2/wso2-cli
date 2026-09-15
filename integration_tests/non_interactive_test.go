package integrationtests_test

import (
	"testing"

	"github.com/wso2/integration-platform-tools/internal/config"
)

func TestNonInteractiveMode(t *testing.T) {
	// Verify that non-interactive mode is enabled in integration tests
	if !config.NonInteractive {
		t.Error("Non-interactive mode should be enabled for integration tests")
	}
}
