package common

import (
	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/connections"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

func GetProjectConnectionList(orgId string, projectId string) ([]connections.Connection, error) {
	connectionsSpinner := utils.CreateSpinner(i18n.T(" Fetching connections..."), "")
	connectionsSpinner.Start()
	connections, err := auth.ConnectionsClient.GetConnectionList(orgId, projectId, "")
	connectionsSpinner.Stop()
	if err != nil {
		return nil, err
	}

	return connections, nil
}

func GetComponentConnectionList(orgId string, projectId string, componentId string) ([]connections.Connection, error) {
	connectionsSpinner := utils.CreateSpinner(i18n.T(" Fetching component connections..."), "")
	connectionsSpinner.Start()
	connections, err := auth.ConnectionsClient.GetConnectionList(orgId, projectId, componentId)
	connectionsSpinner.Stop()
	if err != nil {
		return nil, err
	}

	return connections, nil
}

func GetConnectionItem(orgId string, id string) (connections.Connection, error) {
	connectionSpinner := utils.CreateSpinner(i18n.T(" Fetching connection item..."), "")
	connectionSpinner.Start()
	connectionItem, err := auth.ConnectionsClient.GetConnectionItem(orgId, id)
	connectionSpinner.Stop()
	if err != nil {
		return connections.Connection{}, err
	}

	return connectionItem, nil
}

func GetConnectionConfigString(connectionItem connections.Connection, envs []project.ProjectEnvironment) string {
	configs, hasConfigs := "", false

	for _, envItem := range envs {
		connectionConfig := connectionItem.Configurations[envItem.TemplateId].Entries

		configSecret, serviceUrl, consumerKey, tokenUrl := "*****", "", "", ""

		if connectionConfig["ConsumerSecret"].Value != "" {
			configSecret = connectionConfig["ConsumerSecret"].Value
		}

		serviceUrl = connectionConfig["ServiceURL"].Value
		consumerKey = connectionConfig["ConsumerKey"].Value
		tokenUrl = connectionConfig["TokenURL"].Value

		if serviceUrl != "" {
			hasConfigs = true
			if consumerKey != "" && tokenUrl != "" {
				// Connections with public or org level visibility
				configs += heredoc.Docf(i18n.T(`

					%s
					ServiceURL: %s
					ConsumerKey: %s
					ConsumerSecret: %s
					TokenURL: %s
				`),
					envItem.Name,
					serviceUrl,
					consumerKey,
					configSecret,
					tokenUrl,
				)
			} else {
				// Connections with project level visibility
				configs += heredoc.Docf(i18n.T(`

					%s
					ServiceURL: %s
				`),
					envItem.Name,
					serviceUrl,
				)
			}
		} else {
			// Connections without service URL (probably due to service not being deployed)
			configs += heredoc.Docf(i18n.T(`

				%s
				Service URL not found.
			`),
				envItem.Name,
			)
		}
	}

	if hasConfigs {
		configs += heredoc.Docf(i18n.T(`

			Refer to the instructions provided in the following link to learn how to use these configurations :
				https://wso2.com/choreo/docs/develop-components/sharing-and-reusing-services/
		`))
	}

	return configs
}
