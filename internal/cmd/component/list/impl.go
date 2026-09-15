package list

import (
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
)

func HandleComponentListCommand(opts *CmpListOptions) error {
	// An invalid --output falls back to a persisted OUTPUT_FORMAT (with a
	// warning); with no global set it is an error.
	outputFormat, err := common.ResolveOutputFormat(opts.OutputFlag)
	if err != nil {
		return err
	}

	orgId, projectId, err := common.ResolveContext(opts.OrgFlag, opts.ProjectFlag)

	if err == nil {
		if opts.OrgFlag == "" {
			opts.OrgFlag = orgId
		}

		if opts.ProjectFlag == "" {
			opts.ProjectFlag = projectId
		}
	}

	targetOrg, err := common.ResolveTargetOrganization(opts.OrgFlag)
	if err != nil {
		return err
	}

	project, err := common.ResolveTargetProject(targetOrg, opts.ProjectFlag)
	if err != nil {
		return err
	}

	components, err := getComponents(targetOrg.Handle, targetOrg.ID, project.ID)
	if err != nil {
		return err
	}

	if outputFormat.IsStructured() {
		return printComponentListStructured(components, outputFormat, *project)
	}

	err = printComponentList(targetOrg.ID, project.ID, components)
	if err != nil {
		return err
	}

	return nil
}
