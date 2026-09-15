package create

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var getTokenParams GetTokenParams

var ComponentTokenCommand = &cobra.Command{
	Use:   "test-key [flags]",
	Short: i18n.T("Get the test key for a publicly deployed service."),
	Long:  i18n.T("Get the test key needed to invoke a publicly deployed service in the Integration Platform."),
	Example: heredoc.Docf(i18n.T(`
	
		To get the test key of a particular component in the development environment :
			%s
	`),
		"$ wso2-integration-platform create test-key --project=<project-name> --component=<component-name> --deployment-track=main --env=Development",
	),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		err := handleGetApiKey(&getTokenParams)
		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddEnvFlag(ComponentTokenCommand.Flags(), &getTokenParams.envFlag)
	common.AddOrgFlag(ComponentTokenCommand.Flags(), &getTokenParams.orgFlag)
	common.AddProjectFlag(ComponentTokenCommand.Flags(), &getTokenParams.projectFlag)
	common.AddComponentFlag(ComponentTokenCommand.Flags(), &getTokenParams.componentFlag)
	common.AddDeploymentTrackFlag(ComponentTokenCommand.Flags(), &getTokenParams.deploymentTrackFlag)
	ComponentTokenCommand.Flags().StringVar(&getTokenParams.endpointFlag, "endpoint", "", "name of the endpoint")
	common.AddGenericHelper(ComponentTokenCommand)
}
