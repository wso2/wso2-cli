package list

import (
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
)

func handleListProjects(params *ListProjectsParams) error {
	// An invalid --output falls back to a persisted OUTPUT_FORMAT (with a
	// warning); with no global set it is an error.
	outputFormat, err := common.ResolveOutputFormat(params.outputFlag)
	if err != nil {
		return err
	}

	selectedOrg, err := common.ResolveTargetOrganization(params.orgFlag)
	if err != nil {
		return err
	}

	projects, err := common.FetchProjectsOfOrg(*selectedOrg)
	if err != nil {
		return err
	}

	if outputFormat.IsStructured() {
		return printProjectsStructured(projects, outputFormat, *selectedOrg)
	}

	printProjects(projects, *selectedOrg)
	return nil
}
