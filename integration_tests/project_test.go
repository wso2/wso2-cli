package integrationtests_test

import (
	"log"
	"testing"

	projectCreate "github.com/wso2/integration-platform-tools/internal/cmd/project/create"
	projectDelete "github.com/wso2/integration-platform-tools/internal/cmd/project/delete"
)

func TestCreateProject(t *testing.T) {
	log.Printf("=== Starting TestCreateProject ===")
	log.Printf("Target project name: %s", TestProjectName)

	// Start with a clean slate
	log.Printf("Step 1: Cleaning up existing project (if any)")
	var DelOpts = &projectDelete.ProjectDeleteOpts{
		Project: TestProjectName,
		Force:   true,
	}

	err := projectDelete.HandleDeleteProject(DelOpts)
	if err != nil {
		log.Printf("Cleanup warning (expected if project doesn't exist): %v", err)
	} else {
		log.Printf("Successfully cleaned up existing project")
	}

	// Create a new project
	log.Printf("Step 2: Creating new project '%s'", TestProjectName)
	var Opts = &projectCreate.CreateProjectParams{
		Name: TestProjectName,
	}

	err = projectCreate.HandleProjectCreate(Opts)

	if err != nil {
		log.Printf("Failed to create project: %v", err)
		t.Errorf("Error: %v", err)
	} else {
		log.Printf("Successfully created project '%s'", TestProjectName)
	}

	log.Printf("=== TestCreateProject Completed ===")
}

func TestDeleteProject(t *testing.T) {
	log.Printf("=== Starting TestDeleteProject ===")
	log.Printf("Target project name: %s", TestProjectName)

	log.Printf("Step 1: Deleting project '%s'", TestProjectName)
	var Opts = &projectDelete.ProjectDeleteOpts{
		Project: TestProjectName,
		Force:   true,
	}

	err := projectDelete.HandleDeleteProject(Opts)

	if err != nil {
		log.Printf("Failed to delete project: %v", err)
		t.Errorf("Error: %v", err)
	} else {
		log.Printf("Successfully deleted project '%s'", TestProjectName)
	}

	log.Printf("=== TestDeleteProject Completed ===")
}
