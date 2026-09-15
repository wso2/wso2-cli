package describe

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

func handleDescribeConnection(params *DescribeConnectionsParams) error {
	outputFormat, err := common.ParseOutputFormat(params.outputFlag)
	if err != nil {
		return err
	}

	// When --output json is piped or redirected (stdout is not a TTY), any
	// interactive picker would contaminate the JSON stream. A connection is
	// resolved through component -> connection, and each of those falls back to
	// an interactive prompt when its flag is omitted. Require both up front in
	// structured non-TTY mode, rather than letting a downstream resolver fail
	// with a low-level "could not open a new TTY" error.
	//
	// --project is intentionally not required here: it can be resolved from the
	// local context file (see ResolveContext below), so demanding the flag
	// would reject a valid context-based setup.
	//
	// Use a connection-describe-specific message rather than the generic
	// CreateNonInteractiveError: the quickest fix is often to drop the pipe and
	// let the pickers run.
	if outputFormat.IsStructured() && !utils.IO.IsStdoutTTY() &&
		(params.nameFlag == "" || params.componentFlag == "") {
		return fmt.Errorf("%s", heredoc.Doc(i18n.T(`
			--component and --name are required with --output=json when output is piped or redirected.

			Provide the connection (and the flags that identify it) explicitly:
			  wso2-integration-platform describe connection --name=<connection-name> --project=<project> --component=<component>

			Or run the command on a terminal without a pipe to choose them interactively:
			  wso2-integration-platform describe connection --output=json`)))
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

	connectionList, err := common.GetComponentConnectionList(selectedOrg.ID, project.ID, remoteComponent.Id)
	if err != nil {
		return err
	}

	connectionsListItem, err := resolveTargetConnection(connectionList, params.nameFlag)
	if err != nil {
		return err
	}

	connection, err := common.GetConnectionItem(selectedOrg.ID, connectionsListItem.GroupUuid)
	if err != nil {
		return err
	}

	envs, err := common.GetAllProjectEnv(selectedOrg.UUID, selectedOrg.ID, project.ID)
	if err != nil {
		return err
	}

	if outputFormat.IsStructured() {
		return emitConnectionJSON(connection, *envs, outputFormat)
	}

	utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

			Connection details:

			Name: %s
			Connecting to: %s
			Connection schema: %s
			%s
		`),
		connection.Name,
		connection.ServiceName,
		connection.SchemaName,
		common.GetConnectionConfigString(connection, *envs),
	))

	return nil
}
