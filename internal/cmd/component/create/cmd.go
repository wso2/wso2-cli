package create

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"

	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
)

var createParams CreateComponentParams
var buildConfigurations []string
var autoBuild bool
var autoDeploy bool

var ComponentCreateCommand = &cobra.Command{
	Use:     "component [component-name] [flags]",
	Aliases: []string{"components"},
	Args:    cobra.MaximumNArgs(1),
	Short:   i18n.T("create a component"),
	Long: i18n.T(
		"Create a new integration component in your project. Supported types: service (API, AI Agent, MCP Server), " +
			"scheduleTask (Automation), eventHandler (Event Integration, File Integration). " +
			"Buildpacks: ballerina, microintegrator.",
	),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			createParams.ComponentName = args[0]
		}
		createParams.AutoBuild = &autoBuild
		createParams.AutoDeploy = &autoDeploy

		err := HandleComponentCreate(
			&createParams,
			cmd.Flags().Lookup("build-configs").Changed,
			&buildConfigurations,
		)
		if err != nil {
			utils.HandleErr(err)
		}
	},
	Example: heredoc.Docf(i18n.T(`
		To create a Ballerina API integration in your default project :
			%s

		To create a WSO2 MI automation integration in your default project :
			%s
	`),
		"$ wso2-integration-platform create component <name> --project='Default Project' --type=service --build-pack='ballerina' --repo=<repo-url> --repo-branch=main",
		"$ wso2-integration-platform create component <name> --project='Default Project' --type=scheduleTask --build-pack='microintegrator' --repo=<repo-url> --repo-branch=main",
	),
}

func init() {
	common.AddOrgFlag(ComponentCreateCommand.Flags(), &createParams.OrgFlag)
	common.AddProjectFlag(ComponentCreateCommand.Flags(), &createParams.ProjectFlag)
	common.AddGitCredentialFlag(ComponentCreateCommand.Flags(), &createParams.GitCredName)
	ComponentCreateCommand.Flags().StringVarP(&createParams.ComponentName, "name", "c", "", "name of the new component")
	ComponentCreateCommand.Flags().StringVarP(&createParams.ComponentType, "type", "t", "", fmt.Sprintf("type of the component (%s)", strings.Join(component.ComponentTypes, ", ")))
	ComponentCreateCommand.Flags().StringVar(&createParams.BuildPack, "build-pack", "", "build pack used by the component (ballerina, microintegrator)")
	ComponentCreateCommand.Flags().StringVar(&createParams.Subpath, "dir", "", "subpath of the directory containing the component")
	ComponentCreateCommand.Flags().StringVarP(&createParams.Repo, "repo", "r", "", "git repository url (only for component with multi repo project)")
	ComponentCreateCommand.Flags().StringVarP(&createParams.RepoBranch, "repo-branch", "b", "", "git branch (only for component with multi repo project)")
	ComponentCreateCommand.Flags().StringSliceVarP(&buildConfigurations, "build-configs", "", []string{}, heredoc.Docf(`
		Build configurations for the component provided as a comma separated list of key=value pairs. or multiple flags can be used.
		Example: --build-configs='key1=value1,key2=value2'
		Example: --build-configs='key1=value1' --build-configs='key2=value2'
	`))
	ComponentCreateCommand.Flags().BoolVar(&autoBuild, "auto-build", true, "automatically trigger a build after the component is created")
	ComponentCreateCommand.Flags().BoolVar(&autoDeploy, "auto-deploy", true, "automatically deploy the component after it builds successfully")
	common.AddGenericHelper(ComponentCreateCommand)

}
