package list

import (
	"fmt"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

// Renders the project list in a machine-readable format (json) and writes it
// to stdout. When the list is empty, display a message along with empty array
func printProjectsStructured(projects []models.Project, format common.OutputFormat, currentOrg api.Organization) error {
	// Normalize a nil slice to an empty slice so JSON renders `[]` rather
	// than `null` when the organization has no projects.
	if projects == nil {
		projects = []models.Project{}
	}

	rendered, err := common.RenderStructured(format, projects)
	if err != nil {
		return err
	}

	// The JSON payload always goes to stdout — even when it is just `[]`.
	fmt.Fprintln(utils.IO.Out, rendered)

	// On an empty result, guide the user via stderr
	if len(projects) == 0 {
		fmt.Fprintf(utils.IO.ErrOut, i18n.T("No projects found in organization %s.\n"),
			utils.CS.Bold(currentOrg.Name))
	}
	return nil
}

func printProjects(projects []models.Project, currentOrg api.Organization) {
	if (projects == nil) || (len(projects) == 0) {

		utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

			No projects found in organization %s.
			
			To create a new project :
				%s	
		`),
			utils.CS.Bold(currentOrg.Name),
			"$ wso2-integration-platform create project"))

	} else {
		data := [][]string{}
		for _, project := range projects {
			description := project.Description
			if strings.TrimSpace(description) == "" {
				description = i18n.T("No description")
			}
			parsedDate, _ := time.Parse(time.RFC3339, project.CreatedDate)
			formattedDate := parsedDate.Format("2006-01-02")
			if len(description) > 75 {
				description = description[:75] + "..."
			}
			data = append(data, []string{project.Name, description, formattedDate})
		}

		utils.PrintInfo("%s", heredoc.Docf(i18n.T(`
		
			Showing projects in organization %s.

		`), utils.CS.Bold(currentOrg.Name)))

		tableStr := utils.CreateTable(data, []string{i18n.T("NAME"), i18n.T("DESCRIPTION"), i18n.T("CREATED DATE")}, "")
		fmt.Println(tableStr)
	}

}
