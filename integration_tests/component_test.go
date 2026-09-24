package integrationtests_test

import (
	"log"
	"strings"
	"testing"

	componentList "github.com/wso2/integration-platform-tools/internal/cmd/component/list"
)

// NOTE: The previous create/build/deploy E2E coverage here exercised a nodejs
// `webApp` component. The Integration Platform has no web-app component type
// (see IsIntegrationComponent in pkg/api/component/models.go), so those tests
// were removed rather than ported. Re-adding E2E coverage means driving a
// Ballerina or WSO2 MI integration instead.

func TestListComponentsEmpty(t *testing.T) {
	log.Printf("=== Starting TestListComponentsEmpty ===")
	log.Printf("Target project: %s", TestProjectName)

	log.Printf("Step 1: Listing integrations in empty project")
	var opts = &componentList.CmpListOptions{
		ProjectFlag: TestProjectName,
	}

	got := handleCaptureOutput(func() {
		err := componentList.HandleComponentListCommand(opts)

		if err != nil {
			log.Printf("Error listing integrations: %v", err)
			t.Errorf("Error: %v", err)
		}
	})

	log.Printf("Step 2: Validating empty integrations response")
	log.Printf("Command output: %s", got)

	if !strings.Contains(got, "No integrations found") {
		log.Printf("Expected 'No integrations found' but got: %s", got)
		t.Errorf("Expected: No integrations found, Got: %s", got)
	} else {
		log.Printf("Successfully verified project has no integrations")
	}

	log.Printf("=== TestListComponentsEmpty Completed ===")
}
