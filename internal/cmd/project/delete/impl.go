package delete

import (
	"errors"
	"fmt"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/commons"
)

func HandleDeleteProject(opts *ProjectDeleteOpts) error {

	org, err := common.ResolveTargetOrganization(opts.Org)
	if err != nil {
		return err
	}

	project, err := common.ResolveTargetProject(org, opts.Project)
	if err != nil {
		return err
	}

	if !opts.Force {
		okTodelete := commons.ConfirmDeleteWithName(project.Name)

		if !okTodelete {
			return errors.New("Project deletion cancelled.")
		}
	}

	comps, err := common.GetComponentsForProject(org, project)

	if err != nil {
		return err
	}

	if len(comps) > 0 && !opts.Force {
		return errors.New("Project contains components. Delete the components before deleting the project.")
	}

	if opts.Force {
		deleteCmpSpinner := utils.CreateSpinner("Deleting components.", "")
		if len(comps) > 0 {
			fmt.Fprintln(
				utils.IO.Out,
				i18n.T("Detected components in the project. Deleting components prior to project deletion."),
			)
		}
		for _, comp := range comps {
			deleteCmpSpinner.Start()
			resp, err := auth.ComponentClient.DeleteComponent(org.ID, org.Handle, comp.Id, project.ID)
			deleteCmpSpinner.Stop()

			if err != nil {
				return err
			}

			if resp.Status == "success" {
				fmt.Fprintln(utils.IO.Out, fmt.Sprintf("Component %s deleted successfully.", comp.Name))
			} else {
				return errors.New("Error deleting component.")
			}
		}
	}

	deleteProjSpinner := utils.CreateSpinner("Deleting project", "")
	deleteProjSpinner.Start()
	resp, err := auth.ProjectClient.DeleteProject(org.ID, project.ID)
	deleteProjSpinner.Stop()

	if err != nil {
		return err
	}

	if resp.Status == "success" {
		fmt.Fprintln(utils.IO.Out, "Project deleted successfully.")
		return nil
	} else {
		return errors.New("Error deleting project.")
	}
}
