package describe

import (
	"fmt"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

func handleProjectDescribe(flags *ProjectDescribeFlags) error {
	outputFormat, err := common.ParseOutputFormat(flags.outputFlag)
	if err != nil {
		return err
	}

	org, err := common.ResolveTargetOrganization(flags.Org)
	if err != nil {
		return fmt.Errorf(i18n.T("Error resolving organization: %w"), err)
	}

	project, err := common.ResolveTargetProject(org, flags.Project)
	if err != nil {
		return fmt.Errorf(i18n.T("Error resolving project: %w"), err)
	}

	// Project metadata already succeeded; a failure fetching components should
	// not blank out the rest of the describe. Warn on stderr and continue with
	// an empty list so JSON consumers still get a parseable object and text-mode
	// users still see the project details.
	spinner := utils.CreateSpinner(i18n.T("Fetching project components..."), "")
	spinner.Start()
	cmpsPtr, cmpsErr := auth.ProjectClient.GetProjectComponents(org.ID, org.Handle, project.ID)
	spinner.Stop()
	var components []models.Component
	if cmpsErr != nil {
		fmt.Fprintf(utils.IO.ErrOut,
			i18n.T("warning: failed to fetch project components: %s\n"), cmpsErr)
	} else if cmpsPtr != nil {
		components = *cmpsPtr
	}

	if outputFormat.IsStructured() {
		return printProjectStructured(project, org, components, outputFormat)
	}

	printProjectInfo(project, org)
	printProjectComponents(components)
	return nil
}
