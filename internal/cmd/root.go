// Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
//
// WSO2 LLC. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
package cmd

import (
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/clirpc"
	"github.com/wso2/integration-platform-tools/internal/cmd/auth"
	connect "github.com/wso2/integration-platform-tools/internal/cmd/connect"
	ctxCmd "github.com/wso2/integration-platform-tools/internal/cmd/context/set"
	envCmd "github.com/wso2/integration-platform-tools/internal/cmd/env"
	logsCmd "github.com/wso2/integration-platform-tools/internal/cmd/logs"
	orgChange "github.com/wso2/integration-platform-tools/internal/cmd/org/change"
	"github.com/wso2/integration-platform-tools/internal/cmd/region"
	"github.com/wso2/integration-platform-tools/internal/cmd/updatechk"
	"github.com/wso2/integration-platform-tools/internal/config"
	"github.com/wso2/integration-platform-tools/internal/mcp"
)

var Version = "DEV"

func getExampleMessage() string {
	return heredoc.Docf(i18n.T(`
		To list existing projects or integrations :
			%s
			%s

		To create a new project :
			%s

		To add an integration from your locally cloned repository :
			%s

		To build & deploy an integration :
			%s
			%s
			%s

		To change the current organization :
			%s
	`),
		"$ wso2-integration-platform list projects",
		"$ wso2-integration-platform list integrations",
		"$ wso2-integration-platform create project",
		"$ wso2-integration-platform create integration <integration-name>",
		"$ wso2-integration-platform create build <integration-name>",
		"$ wso2-integration-platform create deployment <integration-name> -e=Development",
		"$ wso2-integration-platform suspend deployment <integration-name> -e=Development",
		"$ wso2-integration-platform change-org",
	)
}

func getHelpTemplate(cmd *cobra.Command) string {
	entityManagementCommands := heredoc.Docf(i18n.T(`
	Resource management:
        create    %s
        delete    %s
        describe  %s
        list      %s
        suspend   %s
        resume    %s
        connect   %s
	`),
		createCmd.Short,
		deleteCmd.Short,
		describeCmd.Short,
		listCmd.Short,
		suspendCmd.Short,
		resumeCmd.Short,
		connect.ConnectCmd.Short,
	)

	buildAndDeployCommands := heredoc.Docf(i18n.T(`
	Monitoring:
        logs    %s

	CLI configuration:
        env     %s
	`),
		logsCmd.LogsCommand.Short,
		envCmd.EnvCommand.Short)

	orgAndAuthCommands := heredoc.Docf(i18n.T(`
	Organization and authentication:
        login         %s
        logout        %s
        change-org    %s
        set-context   %s
		region        %s
	`),
		auth.LoginCmd.Short,
		auth.LogoutCmd.Short,
		orgChange.ChangeOrgCommand.Short,
		ctxCmd.SetCtxCmd.Short,
		region.RegionCommand.Short,
	)

	return heredoc.Docf(i18n.T(`
			%s

			Available commands:
			%s
			%s
			%s

			Flags:
				-h, --help            help for wso2-integration-platform
				    --non-interactive disable interactive prompts
				-v, --version         version for wso2-integration-platform

			Use "wso2-integration-platform [command] --help" for more information about a command.


			Examples:
			%s
		`),
		cmd.Long,
		entityManagementCommands,
		buildAndDeployCommands,
		orgAndAuthCommands,
		getExampleMessage(),
	)
}

var rootCmd = &cobra.Command{
	Use:     "wso2-integration-platform",
	Short:   i18n.T("CLI tool for the Integration Platform"),
	Long:    i18n.T("Welcome to the Integration Platform CLI."),
	Example: getExampleMessage(),
	Version: Version,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

	if len(os.Args) == 1 || len(os.Args) > 1 && os.Args[1] != "completion" {
		updatechk.NewUpdateChecker(Version).Run()
	}
}

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.PersistentFlags().BoolVar(&config.NonInteractive, "non-interactive", os.Getenv("CI") == "true", "Disable interactive prompts and fail with an error when required parameters are missing")
	rootCmd.AddCommand(auth.LoginCmd)
	rootCmd.AddCommand(auth.LogoutCmd)
	rootCmd.AddCommand(createCmd)
	// rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(describeCmd)
	rootCmd.AddCommand(orgChange.ChangeOrgCommand)
	rootCmd.AddCommand(logsCmd.LogsCommand)
	rootCmd.AddCommand(ctxCmd.SetCtxCmd)
	rootCmd.AddCommand(clirpc.RPCCmd)
	rootCmd.AddCommand(mcp.MCPCommand)
	rootCmd.AddCommand(suspendCmd)
	rootCmd.AddCommand(resumeCmd)
	rootCmd.AddCommand(region.RegionCommand)
	rootCmd.AddCommand(connect.ConnectCmd)
	rootCmd.AddCommand(envCmd.EnvCommand)

	// Please update the help template manually whenever adding/removing root level commands
	rootCmd.SetHelpTemplate(getHelpTemplate(rootCmd))
}
