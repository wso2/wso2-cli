package create

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var deployParams DeployComponentParams

var deployConfigMapBaseName = "deploy-envs"

var ComponentDeployCommand = &cobra.Command{
	Use:   "deployment [component-name] [flags]",
	Args:  cobra.MaximumNArgs(1),
	Short: i18n.T("trigger a deployment for a component"),
	Long:  i18n.T("Trigger a deployment for a component in a selected environment within your project."),
	Example: heredoc.Docf(i18n.T(`
		To trigger a deployment of a specific component to the development environment:
			%s
	`),
		"$ wso2-integration-platform create deployment <component-name> --project=<project-name> --deployment-track=main --env=Development"),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			deployParams.ComponentFlag = args[0]
		}

		envFlagsStr, err := cmd.Flags().GetString("env-vars")
		if err != nil {
			utils.HandleErr(err)
		}

		deployParams.EnvVars = common.ConvertStrToEnvVarSlice(envFlagsStr)

		err = HandleDeployComponent(&deployParams)

		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddEnvFlag(ComponentDeployCommand.Flags(), &deployParams.EnvFlag)
	common.AddOrgFlag(ComponentDeployCommand.Flags(), &deployParams.OrgFlag)
	common.AddProjectFlag(ComponentDeployCommand.Flags(), &deployParams.ProjectFlag)
	common.AddDeploymentTrackFlag(ComponentDeployCommand.Flags(), &deployParams.DeploymentTrackFlag)
	ComponentDeployCommand.Flags().StringVar(&deployParams.RunIdFlag, "build-id", "0", i18n.T("Build ID of the build"))
	common.AddEnvVarsFlag(ComponentDeployCommand.Flags())

	// cron specific flags
	ComponentDeployCommand.Flags().
		StringVarP(
			&deployParams.cronDeployOpts.Expression,
			"cron-expression",
			"",
			"",
			i18n.T("schedule expression for the cron job (scheduled tasks only)"),
		)
	ComponentDeployCommand.Flags().
		StringVarP(
			&deployParams.cronDeployOpts.TimeZone,
			"cron-timezone",
			"",
			"",
			i18n.T("schedule timezone for the cron job (scheduled tasks only)"),
		)

	// proxy specific flags
	ComponentDeployCommand.Flags().
		StringVarP(
			&deployParams.proxyDeployOpts.TargetEp,
			"proxy-target-url",
			"",
			"",
			i18n.T("target endpoint for the proxy (git based proxies only)"),
		)
	ComponentDeployCommand.Flags().
		StringVarP(
			&deployParams.proxyDeployOpts.SandboxEp,
			"proxy-sandbox-url",
			"",
			"",
			i18n.T("sandbox endpoint for the proxy (git based proxies only"),
		)

	// BYOI specific flags
	ComponentDeployCommand.Flags().
		StringVarP(
			&deployParams.byoiDeployOpts.ImageWithTag,
			"byoi-image",
			"",
			"",
			i18n.T("deployment image with tag (image based components only)"),
		)
	ComponentDeployCommand.Flags().
		StringArrayVarP(
			&deployParams.byoiDeployOpts.ApiSchemaFile,
			"byoi-api-schema-file",
			"",
			[]string{},
			i18n.T("API schema file path (image based components only)"),
		)
	ComponentDeployCommand.Flags().
		StringVarP(
			&deployParams.byoiDeployOpts.EndpointsFile,
			"byoi-endpoints-file",
			"",
			"",
			i18n.T("service endpoints file path (image based components only)"),
		)
}
