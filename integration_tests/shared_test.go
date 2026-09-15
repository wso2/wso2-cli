package integrationtests_test

import (
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	"github.com/wso2/integration-platform-tools/internal/config"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

const BASE_PROJECT_NAME = "ip-cli-e2e-test"
const USER_EMAIL = "REDACTED"
const STAGE_USER_EMAIL = "REDACTED"

// NOTE: TEST_COMPONENT_NAME, DEPLOYMENT_TRACK, BUILD_PACK_TYPE, SUB_PATH and the
// REPO_* sample-repo URLs were removed with the nodejs webApp E2E tests. Any
// future integration-level component test should define its own fixtures for a
// Ballerina or WSO2 MI integration.

var (
	TEST_USER_NAME  = ""
	TEST_USER_PASS  = ""
	RUN_OS          = ""
	ENV             = ""
	TestProjectName = ""
)

func TestMain(m *testing.M) {

	TEST_USER_NAME = os.Getenv("CHOREO_CLI_TEST_USER_NAME")
	TEST_USER_PASS = os.Getenv("CHOREO_CLI_TEST_USER_PASS")
	RUN_OS = os.Getenv("RUN_OS")
	ENV = os.Getenv("WSO2IP_ENV")

	// Enable non-interactive mode for all integration tests
	config.NonInteractive = true
	log.Printf("=== Integration Test Setup ===")
	log.Printf("Non-interactive mode enabled: %t", config.NonInteractive)

	// Enhanced environment variable logging for CI debugging
	log.Printf("=== CI Environment Debug Info ===")
	log.Printf("RUN_OS: %s", RUN_OS)
	log.Printf("WSO2IP_ENV: %s", ENV)
	log.Printf("CI: %s", os.Getenv("CI"))
	log.Printf("TRACE_ENABLED: %s", os.Getenv("TRACE_ENABLED"))
	log.Printf("LOG_MODE: %s", os.Getenv("LOG_MODE"))
	log.Printf("=== End CI Environment Debug Info ===")

	// Set derived values
	if RUN_OS == "" {
		TestProjectName = BASE_PROJECT_NAME
	} else {
		TestProjectName = fmt.Sprintf("%s-%s", BASE_PROJECT_NAME, RUN_OS)
	}

	os.Exit(m.Run())
}

func handleCaptureOutput(fn func()) string {
	old := utils.IO.Out
	errOld := utils.IO.ErrOut

	r, w, _ := os.Pipe()

	utils.IO.Out = w
	utils.IO.ErrOut = w

	fn()

	w.Close()

	out, _ := io.ReadAll(r)

	defer func() {
		utils.IO.Out = old
		utils.IO.ErrOut = errOld
	}()

	return string(out)
}
