package list

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var buildListFlags = &BuildListFlags{}

var BuildListCommand = &cobra.Command{
	Use:     "builds",
	Aliases: []string{"build"},
	Short:   i18n.T("list all builds of a component"),
	Long:    i18n.T(`This command allows you to list all builds of a component.`),
	Example: heredoc.Docf(i18n.T(`

		To list builds of a component :
			%s

		To get the output in JSON format :
			%s

		To save the output to a file :
			%s

	`),
		"$ wso2-integration-platform list builds --component=<component-name> --project=<project-name> --deployment-track=<deployment-track>",
		"$ wso2-integration-platform list builds --component=<component-name> --project=<project-name> --output=json",
		"$ wso2-integration-platform list builds --component=<component-name> --project=<project-name> --output=json > builds.json"),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handleBuildListCommand(buildListFlags); err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddComponentFlag(BuildListCommand.Flags(), &buildListFlags.Component)
	common.AddProjectFlag(BuildListCommand.Flags(), &buildListFlags.Project)
	common.AddOrgFlag(BuildListCommand.Flags(), &buildListFlags.Org)
	common.AddDeploymentTrackFlag(BuildListCommand.Flags(), &buildListFlags.DeploymentTrack)
	common.AddOutputFlag(BuildListCommand.Flags(), &buildListFlags.OutputFlag)
}
