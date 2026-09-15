package create

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var params ConfigCreateOptions

var ConfigCreateCommand = &cobra.Command{
	Use:   "config [flags]",
	Short: i18n.T("create config"),
	Long:  heredoc.Doc(i18n.T(`Create a new config-map or secret for a component`)),
	Example: heredoc.Docf(i18n.T(`
	
		To create a new config-map containing environment variables within your component :
			%s

		To create a new secret file mount within your component :
			%s
			
	`),
		`$ wso2-integration-platform create config --project=<project-name> --component=<component-name> --type=config-map --mount-type=env-variables  --name=<config-name> --env-vars=key1=val1,key2=val`,
		`$ wso2-integration-platform create config --project=<project-name> --component=<component-name> --type=secret --mount-type=file-mount  --name=<config-name>`),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {

		envFlagsStr, err := cmd.Flags().GetString("env-vars")
		if err != nil {
			utils.HandleErr(err)
		}

		params.envVars = common.ConvertStrToEnvVarSlice(envFlagsStr)

		err = handleConfigCreateCommand()

		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOrgFlag(ConfigCreateCommand.Flags(), &params.orgFlag)
	common.AddProjectFlag(ConfigCreateCommand.Flags(), &params.projectFlag)
	common.AddComponentFlag(ConfigCreateCommand.Flags(), &params.componentFlag)
	common.AddEnvFlag(ConfigCreateCommand.Flags(), &params.envFlag)
	common.AddDeploymentTrackFlag(ConfigCreateCommand.Flags(), &params.deploymentTrackFlag)
	ConfigCreateCommand.Flags().StringVarP(&params.nameFlag, "name", "n", "", "name of the new config")
	ConfigCreateCommand.Flags().StringVarP(&params.typeFlag, "type", "t", "", fmt.Sprintf("type of the config (%s)", strings.Join(common.ConfigTypes, ", ")))
	ConfigCreateCommand.Flags().StringVar(&params.mountTypeFlag, "mount-type", "", fmt.Sprintf("mount type of the config (%s)", strings.Join(common.MountTypes, ", ")))
	ConfigCreateCommand.Flags().StringVar(&params.fileMountPathFlag, "mount-path", "", "mount path of the config file")
	ConfigCreateCommand.Flags().StringVar(&params.fileMountContentFlag, "mount-content", "", "content of the file to be mounted")
	common.AddEnvVarsFlag(ConfigCreateCommand.Flags())
	common.AddGenericHelper(ConfigCreateCommand)
}
