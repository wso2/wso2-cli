package list

import (
	"fmt"
	"strings"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

// printComponentListStructured renders the component list in a machine-readable
// format (json) and writes it to stdout. When the list is empty, a
// human-readable hint is written to stderr instead of stdout, so a redirected
// file or a pipe (e.g. `-o json > components.json` or `| jq`) still receives
// only valid JSON.
func printComponentListStructured(components []models.Component, format common.OutputFormat, currentProject models.Project) error {
	// Normalize a nil slice to an empty slice so JSON renders `[]` rather
	// than `null` when the project has no components.
	if components == nil {
		components = []models.Component{}
	}

	rendered, err := common.RenderStructured(format, components)
	if err != nil {
		return err
	}

	// The JSON payload always goes to stdout — even when it is just `[]`.
	fmt.Fprintln(utils.IO.Out, rendered)

	// On an empty result, guide the user via stderr (kept out of stdout so
	// the JSON stays machine-parseable).
	if len(components) == 0 {
		fmt.Fprintf(utils.IO.ErrOut, i18n.T("No components found in project %s.\n"),
			utils.CS.Bold(currentProject.Name))
	}
	return nil
}

func getComponents(orgHandler string, orgId string, projectId string) ([]models.Component, error) {
	componentsSpinner := utils.CreateSpinner(" Fetching components...", "\n")
	componentsSpinner.Start()
	components, err := auth.ComponentClient.GetAllComponents(orgHandler, orgId, projectId, false)
	componentsSpinner.Stop()

	if err != nil {
		return nil, fmt.Errorf(i18n.T("failed to fetch components: %w"), err)
	}

	return components, nil
}

func printComponentList(orgId string, projectId string, components []models.Component) error {
	if len(components) == 0 {
		fmt.Fprintf(utils.IO.Out, i18n.T("%s No components found\n"), utils.CS.Yellow("!"))
		return nil
	} else {
		data := [][]string{}

		for _, componentItem := range components {

			buildPack := "n/a"

			if componentItem.DisplayType == component.DisplayTypeByocWebAppDockerLess {
				// ByocWebAppBuildConfig is a pointer and may be nil; guard the
				// dereference so the listing does not panic when it is unset.
				if componentItem.Repository.ByocWebAppBuildConfig != nil {
					buildPack = componentItem.Repository.ByocWebAppBuildConfig.WebAppType
				}
			} else if strings.HasPrefix(componentItem.DisplayType, "byoi") {
				buildPack = "Container Image"
			} else if strings.HasPrefix(componentItem.DisplayType, "byoc") {
				buildPack = component.ComponentBuildPackDocker
			} else if strings.HasPrefix(componentItem.DisplayType, "buildpack") {
				componentsSpinner := utils.CreateSpinner(" Fetching component details...", "\n")
				componentsSpinner.Start()
				compDetails, err := auth.ComponentClient.GetComponentInfo(orgId, componentItem.Handler, projectId)
				if err != nil {
					return fmt.Errorf(i18n.T("failed to fetch component details: %w"), err)
				}
				componentsSpinner.Stop()
				if len(compDetails.Repository.BuildPackConfig) > 0 {
					buildPack = compDetails.Repository.BuildPackConfig[0].Buildpack.Language
				}
			} else if strings.HasPrefix(componentItem.DisplayType, "mi") {
				buildPack = component.ComponentBuildPackMI
			} else {
				buildPack = component.ComponentBuildPackBallerina
			}

			data = append(data, []string{
				componentItem.DisplayName,
				component.GetTypeForDisplayType(componentItem.DisplayType),
				strings.ToLower(buildPack),
			})
		}

		tableStr := utils.CreateTable(data, []string{i18n.T("NAME"), i18n.T("TYPE"), i18n.T("BUILD PACK")}, "")
		fmt.Fprintln(utils.IO.Out, tableStr)
	}
	return nil
}
