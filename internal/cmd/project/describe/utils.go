package describe

import (
	"fmt"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

func printProjectInfo(project *models.Project, org *api.Organization) {
	tableRows := [][]string{
		{"ID:", project.ID},
		{"Name:", project.Name},
		{"Description:", sanitizeDisplayText(project.Description)},
		{"Created Date:", formatDate(project.CreatedDate)},
		{"Handler:", project.Handler},
		{"Organization:", org.Name},
	}

	if project.Repository != "" && project.Branch != "" && project.GitOrganization != "" {
		tableRows = append(tableRows, []string{"Type:", "mono-repository"})
		tableRows = append(tableRows, []string{"Repository URL:", common.GenRepoUrl(project.GitProvider, project.GitOrganization, project.Repository, "")})
	} else {
		tableRows = append(tableRows, []string{"Type:", "multi-repository"})
	}

	utils.PrintInfo("%s", heredoc.Docf(
		i18n.T(`
			Project details:
			
			%s
		`),
		utils.CreateTable(tableRows, []string{}, ""),
	))
}

func printProjectComponents(components []models.Component) {
	if len(components) == 0 {
		utils.PrintInfo("%s", i18n.T("No components found in the project"))
		return
	}

	data := [][]string{}

	for _, cmp := range components {
		data = append(
			data,
			[]string{
				cmp.Name,
				sanitizeDisplayText(cmp.Description),
				component.GetTypeForDisplayType(cmp.DisplayType),
				utils.GetRelativeTime(cmp.LastBuildDate),
			},
		)
	}

	utils.PrintInfo("%s", heredoc.Docf(
		i18n.T(`
			Listing components in the project:

			%s
		`),
		utils.CreateTable(
			data,
			[]string{i18n.T("Name"), i18n.T("Description"), i18n.T("Type"), i18n.T("Last Updated")},
			"",
		),
	))
}

func sanitizeDisplayText(text string) string {
	if strings.Trim(text, " ") == "" {
		return "N/A"
	}
	return text
}

func formatDate(t string) string {
	time, err := time.Parse(time.RFC3339, t)

	if err != nil {
		return "N/A"
	}

	return time.Format("2006-01-02")
}

func printProjectStructured(project *models.Project, org *api.Organization,
	components []models.Component, format common.OutputFormat) error {
	out := ProjectDescribeOutput{
		ID:           project.ID,
		Name:         project.Name,
		Description:  project.Description,
		CreatedDate:  project.CreatedDate,
		Handler:      project.Handler,
		Organization: org.Name,
		Components:   []ProjectComponentOutput{},
	}

	if project.Repository != "" && project.Branch != "" && project.GitOrganization != "" {
		out.Type = "mono-repository"
		out.RepositoryURL = common.GenRepoUrl(project.GitProvider, project.GitOrganization, project.Repository, "")
	} else {
		out.Type = "multi-repository"
	}

	for _, c := range components {
		out.Components = append(out.Components, ProjectComponentOutput{
			Name:          c.Name,
			Description:   c.Description,
			Type:          component.GetTypeForDisplayType(c.DisplayType),
			LastBuildDate: c.LastBuildDate,
		})
	}

	rendered, err := common.RenderStructured(format, out)
	if err != nil {
		return err
	}
	fmt.Fprintln(utils.IO.Out, rendered)
	return nil
}
