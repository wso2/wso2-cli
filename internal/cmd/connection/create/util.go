package create

import (
	"errors"
	"fmt"
	"strings"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/connections"
	"github.com/wso2/integration-platform-tools/pkg/api/marketplace"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

func resolveConnectionName(connectionName *string) error {
	if *connectionName == "" {
		err := prompt.NewPromptInputMessage(
			prompt.PromptInputOpts{
				Message:  i18n.T("Connection name:"),
				Validate: prompt.ValidateNotEmpty,
			},
			connectionName,
		).Prompt()

		if err != nil {
			return fmt.Errorf("%s", i18n.T(" failed to get a valid connection name"))
		}
	}
	return nil
}

func resolveMarketplaceService(
	serviceList []marketplace.MarketplaceService,
	serviceFlag string) (marketplace.MarketplaceService, error) {

	var selectedService *marketplace.MarketplaceService

	if serviceFlag != "" {
		for _, conItem := range serviceList {
			if conItem.Name == serviceFlag {
				selectedService = &conItem
				break
			}
		}
	} else {
		selectedConnName, err := promptToSelectService(serviceList)
		if err != nil {
			return marketplace.MarketplaceService{}, fmt.Errorf("%s", i18n.T(" failed to select the service"))
		}
		for _, item := range serviceList {
			if item.Name == selectedConnName {
				selectedService = &item
				break
			}
		}
	}

	if selectedService == nil {
		return marketplace.MarketplaceService{}, fmt.Errorf("%s", i18n.T(" invalid service selection"))
	}

	return *selectedService, nil
}

func promptToSelectService(serviceList []marketplace.MarketplaceService) (string, error) {
	var selectedServiceName string
	var serviceNames []string

	for _, item := range serviceList {
		serviceNames = append(serviceNames, item.Name)
	}

	err := prompt.NewPromptSelectMessage[string](
		prompt.PromptSelectOpts[string]{Message: "Service:", Values: serviceNames},
		&selectedServiceName,
	).Prompt()

	if err != nil {
		return "", err
	}

	return selectedServiceName, nil
}

func createNewConnection(orgId string, orgUuid string, projectId string, componentItem models.Component, name string, serviceItem marketplace.MarketplaceService, visibility string, schema marketplace.MarketplaceServiceScheme, envs []project.ProjectEnvironment) (connections.Connection, error) {
	createConnectionSpinner := utils.CreateSpinner(i18n.T(" Creating connection..."), "")
	createConnectionSpinner.Start()
	connectionItem, err := auth.ConnectionsClient.CreateNewConnection(orgId, orgUuid, projectId, name, serviceItem.ServiceID, schema.ID, visibility, envs, componentItem.Id, componentItem.DisplayType != component.DisplayTypeByocWebAppDockerLess)
	createConnectionSpinner.Stop()
	if err != nil {
		return connections.Connection{}, err
	}

	return connectionItem, nil
}

func getMarketplaceServices(orgId string, projectId string) ([]marketplace.MarketplaceService, error) {
	serviceListSpinner := utils.CreateSpinner(i18n.T(" Fetching services from marketplace..."), "")
	serviceListSpinner.Start()
	servicesResp, err := auth.MarketplaceClient.GetMarketplaceServices(orgId, marketplace.MarketplaceGetServicesReq{NetworkVisibilityprojectId: &projectId, SortBy: "name", NetworkVisibilityFilter: "all"})
	serviceListSpinner.Stop()
	if err != nil {
		return nil, err
	}
	return servicesResp.Data, nil
}

func resolveConnectionVisibility(marketplaceItem marketplace.MarketplaceService) (string, error) {
	if len(marketplaceItem.Visibility) == 0 {
		return "", errors.New("selected connection item does not have any visibilities")
	} else if len(marketplaceItem.Visibility) == 1 {
		return marketplaceItem.Visibility[0], nil
	} else {
		var selectedVisibilityName string
		err := prompt.NewPromptSelectMessage(
			prompt.PromptSelectOpts[string]{
				Message: i18n.T("Visibility:"),
				Values:  marketplaceItem.Visibility,
			},
			&selectedVisibilityName,
		).Prompt()

		if err != nil {
			return "", err
		}
		return selectedVisibilityName, nil
	}
}

func resolveConnectionSchema(marketplaceItem marketplace.MarketplaceService, selectedVisibility string) (*marketplace.MarketplaceServiceScheme, error) {
	filteredSchemas := []marketplace.MarketplaceServiceScheme{}
	for _, item := range marketplaceItem.ConnectionSchemas {
		if strings.ToLower(selectedVisibility) == "organization" || strings.ToLower(selectedVisibility) == "public" {
			if strings.Contains(strings.ToLower(item.Name), strings.ToLower(selectedVisibility)) || strings.Contains(strings.ToLower(item.Name), "unsecured") {
				filteredSchemas = append(filteredSchemas, item)
			}
		} else if strings.Contains(strings.ToLower(item.Name), "project") {
			filteredSchemas = append(filteredSchemas, item)
		}
	}

	if len(filteredSchemas) == 0 {
		return nil, errors.New("no schemas for selected visibility")
	} else if len(filteredSchemas) == 1 {
		return &filteredSchemas[0], nil
	} else {
		var selectedSchemaName string
		var schemaNames []string

		for _, item := range filteredSchemas {
			schemaNames = append(schemaNames, item.Name)
		}

		err := prompt.NewPromptSelectMessage(
			prompt.PromptSelectOpts[string]{
				Message: i18n.T("schema:"),
				Values:  schemaNames,
			},
			&selectedSchemaName,
		).Prompt()

		if err != nil {
			return nil, err
		}

		for _, item := range filteredSchemas {
			if item.Name == selectedSchemaName {
				return &item, nil
			}
		}

		return nil, errors.New("failed to select schema")
	}
}
