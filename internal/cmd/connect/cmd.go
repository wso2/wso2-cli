package connect

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/connections"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

type ConnectCmdOpts struct {
	Org             string
	Component       string
	Project         string
	DeploymentTrack string
	Env             string
	Recreate        bool
	SkipConnection  []string
	DeleteBridge    bool
	// TODO: uncomment following if enabling ngrok like feature
	// Ports           []int
	// PreviewUrl      bool
}

var cmdOpts ConnectCmdOpts

var ConnectCmd = &cobra.Command{
	Use:   "connect [flags]",
	Short: i18n.T("Connect to remote env"),
	Long:  i18n.T("Connect your development environment with your remote project environment"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := handleConnect(cmd.Context(), &cmdOpts, args); err != nil {
			utils.HandleErr(err)
		}
	},
	PreRun: common.VerifyIsUserLoggedIn,
	Example: heredoc.Docf(i18n.T(`

		To create a shell environment that is connected to your project :
			%s

		To create a shell environment that is connected a different non-critical environment :
			%s

		To inject connection configurations for a specific component :
			%s
		
		To start your local component connected to its dependencies :
			%s

		To skip injecting connection configurations for the given connection names :
			%s

	`),
		"$ wso2-integration-platform connect --project <project-name>",
		"$ wso2-integration-platform connect --project <project-name> --env <env-name>",
		"$ wso2-integration-platform connect --project <project-name> --component <component-name>",
		"$ wso2-integration-platform connect --project <project-name> -- <command-to-start-local-component>",
		"$ wso2-integration-platform connect --project <project-name> --skip-connection <conn-name-1> --skip-connection <conn-name-2>"),
}

func handleConnect(ctx context.Context, opts *ConnectCmdOpts, args []string) error {

	if os.Getenv("WSO2IP_SHELL") == "true" {
		return fmt.Errorf("%s", i18n.T("Please exit the sub-shell and try again"))
	}
	envVars := map[string]string{}
	secureHosts := map[string]bool{}

	orgId, projectId, err := common.ResolveContext(opts.Org, opts.Project)

	if err == nil {
		if opts.Org == "" {
			opts.Org = orgId
		}

		if opts.Project == "" {
			opts.Project = projectId
		}
	}

	org, err := common.ResolveTargetOrganization(opts.Org)
	if err != nil {
		return err
	}

	selectedProject, err := common.ResolveTargetProject(org, opts.Project)
	if err != nil {
		return err
	}

	var remoteComponent *models.Component

	remoteComponents, err := common.GetComponentsWithSystemCompsForProject(org, selectedProject)
	if err != nil {
		return err
	}

	if opts.Component != "" {
		remoteComponent, err = common.ResolveTargetComponentFromCompList(org, selectedProject.ID, remoteComponents, opts.Component)
		if err != nil {
			return err
		}

	}

	var selectedEnv *project.ProjectEnvironment
	if opts.Env == "" {
		envs, err := common.GetAllProjectEnv(org.UUID, org.ID, selectedProject.ID)
		if err != nil {
			return err
		}
		for _, item := range *envs {
			if !item.Critical {
				selectedEnv = &item
				break
			}
		}
		if selectedEnv == nil {
			return fmt.Errorf("%s", i18n.T("Unable to select a non-critical project env"))
		}
	} else {
		selectedEnv, err = common.GetProjectEnv(org.UUID, org.ID, selectedProject.ID, opts.Env)
		if err != nil {
			return err
		}
		if selectedEnv.Critical {
			return fmt.Errorf("%s", i18n.T("Cannot connect with critical environments"))
		}
	}

	// TODO: uncomment following if enabling ngrok like feature
	/*
		if err = getCompPort(opts, remoteComponent, *deploymentTrack, org.ID); err != nil {
			return err
		}
	*/

	_, proxyAgentEndpoints, err := getEndpointsOfProxyAgent(ctx, org, selectedProject, selectedEnv, &remoteComponents, opts.Recreate, opts.DeleteBridge)
	if err != nil {
		return err
	}

	if proxyAgentEndpoints == nil {
		return nil
	}

	proxyRestEndpoint, err := getProxyAgentRestEp(org, proxyAgentEndpoints, selectedEnv)
	if err != nil {
		return err
	}

	// TODO: uncomment following if enabling ngrok like feature
	/*
		proxyWsEndpoint, err := getProxyAgentWsEp(org, proxyAgentEndpoints, selectedEnv)
		if err != nil {
			return err
		}
	*/

	allConnections := []connections.Connection{}
	projectConnections, err := common.GetProjectConnectionList(org.ID, selectedProject.ID)
	if err != nil {
		return err
	}
	allConnections = append(allConnections, projectConnections...)

	err = injectConnectionEnvs(org, selectedProject, nil, selectedEnv, projectConnections, &envVars, &secureHosts, opts.SkipConnection)
	if err != nil {
		return err
	}

	if remoteComponent != nil {
		// only selected component
		compConnections, err := common.GetComponentConnectionList(org.ID, selectedProject.ID, remoteComponent.Id)
		if err != nil {
			return err
		}
		allConnections = append(allConnections, compConnections...)
		err = injectConnectionEnvs(org, selectedProject, remoteComponent, selectedEnv, compConnections, &envVars, &secureHosts, opts.SkipConnection)
		if err != nil {
			return err
		}
	} else {
		// all components of project
		for _, compItem := range remoteComponents {
			compConnections, err := common.GetComponentConnectionList(org.ID, selectedProject.ID, compItem.Id)
			if err != nil {
				return err
			}
			allConnections = append(allConnections, compConnections...)
			err = injectConnectionEnvs(org, selectedProject, &compItem, selectedEnv, compConnections, &envVars, &secureHosts, opts.SkipConnection)
			if err != nil {
				return err
			}
		}
	}

	// TODO: uncomment following if enabling ngrok like feature
	/*
		if opts.PreviewUrl && len(opts.Ports) > 0 {
			for _, port := range opts.Ports {
				componentHandle := fmt.Sprintf("%s-%s", remoteComponent.Handler, strconv.Itoa(port))
				if err := ConnectToServerAgent(user.IDPId, componentHandle, org, proxyWsEndpoint, selectedEnv, port, proxyServerPort); err != nil {
					log.Fatal("Failed to connect to server:", err)
				}
			}
		}
	*/

	proxyServerPort, err := StartLocalProxy(org, proxyRestEndpoint, selectedEnv, &secureHosts)
	if err != nil {
		return err
	}

	if len(allConnections) > 0 {
		utils.PrintInfo("%s", i18n.T("Environment variables for the following connections have been added to the shell:\n"))
		for _, connItem := range allConnections {
			resourceType := strings.ToLower(strings.ReplaceAll(connItem.ResourceType, "_", "-"))
			utils.PrintInfo(i18n.T("- '%s' (%s)\n"), connItem.Name, resourceType)
		}
	}

	utils.PrintInfo("%s", i18n.T("\nYou are now connected to your project.\n"))

	// TODO: uncomment following if enabling ngrok like feature
	/*
		if opts.PreviewUrl && len(opts.Ports) > 0 {
			apiKey, err := getProxyAgentWSEpKey(org, proxyWsEndpoint, selectedEnv)
			if err != nil {
				return err
			}
			for _, port := range opts.Ports {
				componentHandle := fmt.Sprintf("%s-%s", remoteComponent.Handler, strconv.Itoa(port))
				invokeUrl := fmt.Sprintf("%s/preview/%s/%s", proxyRestEndpoint.PublicURL, user.IDPId, componentHandle)
				utils.PrintInfo("%s", heredoc.Docf(`

								Local preview (port %d, expires in %d mins):
									curl --header "Api-Key: %s" %s

							`, port, apiKey.ValidityTime/60, utils.CS.Gray(apiKey.Apikey), invokeUrl))
			}
		}
	*/

	if len(args) == 0 && remoteComponent != nil {
		utils.PrintInfo(i18n.T("\nNext, run the command to start your %s locally.\n"), remoteComponent.DisplayName)
	}

	// Define environment variables to be set in the subshell
	envVars["http_proxy"] = fmt.Sprintf("http://127.0.0.1:%d", proxyServerPort)
	envVars["no_proxy"] = "localhost,127.0.0.1"
	envVars["PS1"] = "(integration-platform-connect) %(?:%{%}%1{➜%} :%{%}%1{➜%} ) %{%}%c%{%} $(git_prompt_info)"
	envVars["WSO2IP_SHELL"] = "true"

	// Detect the user's preferred shell
	var shell string
	if runtime.GOOS == "windows" {
		// On Windows, use PowerShell as the default shell
		shell = os.Getenv("COMSPEC")
		if shell == "" {
			shell = "powershell.exe" // Default to PowerShell if COMSPEC is not set
		}
	} else {
		// On macOS or Linux, use the SHELL environment variable
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/zsh" // Default to zsh if no shell is set
		}
	}

	addPromptToShellConfig(shell)

	// Prepare the shell command
	var subShell *exec.Cmd
	if len(args) > 0 {
		subShell = exec.Command(shell, "-c", strings.Join(args[:], " "))
	} else {
		subShell = exec.Command(shell)
	}

	// Pass the current environment along with the new variables
	subShell.Env = append(os.Environ(), flattenEnv(envVars)...)

	// Set input/output to the current terminal
	subShell.Stdin = os.Stdin
	subShell.Stdout = os.Stdout
	subShell.Stderr = os.Stderr

	// Start the subshell
	if err := subShell.Run(); err != nil {
		fmt.Println("Error starting connection to project:", err)
	} else {
		fmt.Println("Exited project.")
	}

	return nil
}

// Helper function to flatten environment variables into a slice
func flattenEnv(envVars map[string]string) []string {
	var env []string
	for key, value := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}

func init() {
	common.AddOrgFlag(ConnectCmd.Flags(), &cmdOpts.Org)
	common.AddComponentFlag(ConnectCmd.Flags(), &cmdOpts.Component)
	common.AddProjectFlag(ConnectCmd.Flags(), &cmdOpts.Project)
	common.AddEnvFlag(ConnectCmd.Flags(), &cmdOpts.Env)
	common.AddDeploymentTrackFlag(ConnectCmd.Flags(), &cmdOpts.DeploymentTrack)
	ConnectCmd.Flags().BoolVar(&cmdOpts.Recreate, "recreate", false, "recreate the bridge component")
	ConnectCmd.Flags().StringSliceVar(&cmdOpts.SkipConnection, "skip-connection", []string{}, "skip injecting connection configurations for the given connection names")
	ConnectCmd.Flags().BoolVar(&cmdOpts.DeleteBridge, "delete-bridge", false, "Delete the bridge between the local and remote environments")
	// TODO: uncomment following if enabling ngrok like feature
	// ConnectCmd.Flags().IntSliceVar(&cmdOpts.Ports, "port", []int{}, "port of the local application")
	// ConnectCmd.Flags().BoolVar(&cmdOpts.PreviewUrl, "preview-url", false, i18n.T("generate sharable URL for the locally running application"))
	common.AddGenericHelper(ConnectCmd)
}
