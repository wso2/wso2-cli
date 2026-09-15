package list

import (
	"fmt"
	"time"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

func handleConnectionListCommand() error {
	// An invalid --output falls back to a persisted OUTPUT_FORMAT (with a
	// warning); with no global set it is an error.
	outputFormat, err := common.ResolveOutputFormat(params.outputFlag)
	if err != nil {
		return err
	}

	orgId, projectId, err := common.ResolveContext(params.orgFlag, params.projectFlag)
	if err == nil {
		if params.orgFlag == "" {
			params.orgFlag = orgId
		}

		if params.projectFlag == "" {
			params.projectFlag = projectId
		}
	}

	selectedOrg, err := common.ResolveTargetOrganization(params.orgFlag)
	if err != nil {
		return err
	}

	project, err := common.ResolveTargetProject(selectedOrg, params.projectFlag)
	if err != nil {
		return err
	}

	remoteComponent, err := common.ResolveTargetComponent(selectedOrg, project.ID, params.componentFlag)
	if err != nil {
		return err
	}

	connections, err := common.GetComponentConnectionList(selectedOrg.ID, project.ID, remoteComponent.Id)
	if err != nil {
		return err
	}

	if outputFormat.IsStructured() {
		return printConnectionListStructured(connections, outputFormat, *project)
	}

	err = printConnectionList(connections)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	if len(connections) > 0 {
		utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

			To view details of a particular connection :
				%s

		`),
			fmt.Sprintf(`$ wso2-integration-platform describe connection --project="%s" --name=<connection-name>`, project.Name)))
	} else {
		utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

			To create a new connection :
				%s

		`),
			fmt.Sprintf(`$ wso2-integration-platform create connection --project="%s" --service=<service-name> --name=<connection-name>`, project.Name)))
	}

	return nil
}
