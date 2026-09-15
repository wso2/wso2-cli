package describe

import (
	"fmt"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api/connections"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

// emitConnectionJSON renders the resolved connection as a JSON object. It is
// only called when --output=json. UX affordances (the "refer to the docs" hint
// in the text path) are intentionally skipped: the JSON payload describes the
// resource, not the CLI's suggestions.
//
// The per-environment config extraction mirrors GetConnectionConfigString
// exactly — including redacting ConsumerSecret. The JSON path must not become a
// way to exfiltrate the secret material.
func emitConnectionJSON(connection connections.Connection, envs []project.ProjectEnvironment, format common.OutputFormat) error {
	out := ConnectionDescribeOutput{
		Name:         connection.Name,
		ConnectingTo: connection.ServiceName,
		Schema:       connection.SchemaName,
		Environments: make([]ConnectionEnvConfigOutput, 0, len(envs)),
	}

	for _, envItem := range envs {
		entries := connection.Configurations[envItem.TemplateId].Entries

		serviceURL := entries["ServiceURL"].Value
		consumerKey := entries["ConsumerKey"].Value
		tokenURL := entries["TokenURL"].Value

		envOut := ConnectionEnvConfigOutput{Name: envItem.Name}

		if serviceURL == "" {
			// No ServiceURL — usually the target service isn't deployed. The
			// text path prints "Service URL not found." here.
			envOut.ServiceURLFound = false
			out.Environments = append(out.Environments, envOut)
			continue
		}

		envOut.ServiceURLFound = true
		envOut.ServiceURL = serviceURL

		// ConsumerKey + TokenURL are only present for public/org-level
		// connections; project-level connections only carry a ServiceURL.
		if consumerKey != "" && tokenURL != "" {
			envOut.ConsumerKey = consumerKey
			// Always redacted, never the real secret — matching the text path.
			envOut.ConsumerSecret = "*****"
			envOut.TokenURL = tokenURL
		}

		out.Environments = append(out.Environments, envOut)
	}

	rendered, err := common.RenderStructured(format, out)
	if err != nil {
		return err
	}
	fmt.Fprintln(utils.IO.Out, rendered)
	return nil
}

func resolveTargetConnection(
	connectionList []connections.Connection,
	connectionFlag string) (connections.Connection, error) {

	var selectedConnection *connections.Connection

	if connectionFlag != "" {
		for _, conItem := range connectionList {
			if conItem.Name == connectionFlag {
				selectedConnection = &conItem
				break
			}
		}
	} else {
		selectedConnName, err := promptToSelectConnection(connectionList)
		if err != nil {
			return connections.Connection{}, fmt.Errorf("%s", i18n.T(" failed to select the connection"))
		}
		for _, item := range connectionList {
			if item.Name == selectedConnName {
				selectedConnection = &item
				break
			}
		}
	}

	if selectedConnection == nil {
		return connections.Connection{}, fmt.Errorf("%s", i18n.T(" invalid connection selection"))
	}

	return *selectedConnection, nil
}

func promptToSelectConnection(connectionList []connections.Connection) (string, error) {
	var selectedConnectionName string
	var connectionNames []string

	for _, item := range connectionList {
		connectionNames = append(connectionNames, item.Name)
	}

	err := prompt.NewPromptSelectMessage[string](
		prompt.PromptSelectOpts[string]{Message: "Connection:", Values: connectionNames},
		&selectedConnectionName,
	).Prompt()

	if err != nil {
		return "", err
	}

	return selectedConnectionName, nil
}
