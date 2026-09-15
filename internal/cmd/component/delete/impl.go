package delete

import (
	"errors"
	"fmt"

	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/commons"
)

func HandleDeleteComponent(opts *ComponentDeleteOpts) error {

	orgId, projectId, err := common.ResolveContext(opts.Org, opts.Project)

	if err == nil {
		if opts.Org == "" {
			opts.Org = orgId
		}

		if opts.Project == "" {
			opts.Project = projectId
		}
	}

	org, err := common.ResolveTargetOrganization(opts.Org)
	if err != nil {
		utils.HandleErr(err)
	}

	project, err := common.ResolveTargetProject(org, opts.Project)
	if err != nil {
		utils.HandleErr(err)
	}

	cmp, err := common.ResolveTargetComponent(org, project.ID, opts.ComponentName)
	if err != nil {
		utils.HandleErr(err)
	}

	if !opts.SkipConfirm {
		okTodelete := commons.ConfirmDeleteWithName(cmp.Name)

		if !okTodelete {
			return errors.New("Component delete cancelled")
		}
	}

	delCmpSpinner := utils.CreateSpinner("Deleting component", "")
	delCmpSpinner.Start()
	resp, err := auth.ComponentClient.DeleteComponent(org.ID, org.Handle, cmp.Id, project.ID)
	delCmpSpinner.Stop()

	if err != nil {
		utils.HandleErr(err)
	}

	if resp.Status == "success" {
		fmt.Fprintln(utils.IO.Out, "Component deleted successfully")
		return nil
	} else {
		return fmt.Errorf("Error deleting component: %s", resp.Message)
	}
}
