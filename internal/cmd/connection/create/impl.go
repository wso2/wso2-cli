package create

import (
	"fmt"
	"time"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/marketplace"
)

func handleConfigCreateCommand() error {
	err := resolveConnectionName(&params.nameFlag)
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

	services, err := getMarketplaceServices(selectedOrg.ID, project.ID)
	if err != nil {
		return err
	}

	filteredServices := []marketplace.MarketplaceService{}
	for _, item := range services {
		if item.Component.ComponentID != remoteComponent.Id {
			filteredServices = append(filteredServices, item)
		}
	}

	if len(filteredServices) == 0 {
		return fmt.Errorf("no APIs available in the marketplace")
	}

	selectedService, err := resolveMarketplaceService(filteredServices, params.serviceFlag)
	if err != nil {
		return err
	}

	visibility, err := resolveConnectionVisibility(selectedService)
	if err != nil {
		return err
	}

	schema, err := resolveConnectionSchema(selectedService, visibility)
	if err != nil {
		return err
	}

	envs, err := common.GetAllProjectEnv(selectedOrg.UUID, selectedOrg.ID, project.ID)
	if err != nil {
		return err
	}

	connection, err := createNewConnection(selectedOrg.ID, selectedOrg.UUID, project.ID, *remoteComponent, params.nameFlag, selectedService, visibility, *schema, *envs)
	if err != nil {
		return err
	}

	utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

			Successfully created connection %s
			
			Connecting to: %s
			Connection schema: %s
			%s
		`),
		connection.Name,
		connection.ServiceName,
		connection.SchemaName,
		common.GetConnectionConfigString(connection, *envs),
	))

	if selectedService.Visibility[0] != "PROJECT" {
		utils.PrintInfo("%s", heredoc.Docf(i18n.T(`
			Notice: Secret values are only visible once. 
		`),
		))
	}

	time.Sleep(1 * time.Second)

	utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

		To list all the connections available within your project :
			%s

	`),
		fmt.Sprintf(`$ wso2-integration-platform list connections --project="%s""`, project.Name),
	))

	return nil
}
