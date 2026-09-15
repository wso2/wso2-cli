package connect

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/cmd/test-token/create"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api"
	apimpublisher "github.com/wso2/integration-platform-tools/pkg/api/apimPublisher"
	configmapping "github.com/wso2/integration-platform-tools/pkg/api/config-mapping"
	"github.com/wso2/integration-platform-tools/pkg/api/connections"
	"github.com/wso2/integration-platform-tools/pkg/api/devops"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
	"github.com/wso2/integration-platform-tools/pkg/util/constants"
)

var (
	BRIDGE_IMAGE_NAME    = "choreo-local-bridge"
	BRIDGE_IMAGE_VERSION = "1.0.2"
	BRIDGE_COMP_NAME     = "choreo-local-bridge"
)

type CachedApiKey struct {
	response  *apimpublisher.ApiKeyResponse
	lastFetch time.Time
	mu        sync.RWMutex
}

var restApiKeyCache CachedApiKey
var wsApiKeyCache CachedApiKey

// Define the block of code to add
const condition = `
# Show integration platform shell prompt while in the sub-shell
if [ -n "$WSO2IP_SHELL" ]; then
  export PS1="(integration-platform) $PS1"
fi
`

// Get the default shell configuration file based on the SHELL environment variable
func getShellConfigFile(shell string) string {
	// TODO: handle other shell types
	homeDir, _ := os.UserHomeDir()
	switch {
	case strings.Contains(shell, "bash"):
		// Bash shell
		return fmt.Sprintf("%s/.bashrc", homeDir)
	case strings.Contains(shell, "zsh"):
		// Zsh shell
		return fmt.Sprintf("%s/.zshrc", homeDir)
	default:
		return ""
	}
}

func EncodeOpenAPIToBase64(openapiSpec string) string {
	return base64.StdEncoding.EncodeToString([]byte(openapiSpec))
}

func addPromptToShellConfig(shellPath string) {
	if shellPath == "" {
		return
	}

	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return
	}

	// Get the relevant shell config file
	shellConfigFile := getShellConfigFile(shellPath)
	if shellConfigFile == "" {
		return
	}

	// Expand the tilde (~) in the file path
	shellConfigFile = expandTilde(shellConfigFile)
	if shellConfigFile == "" {
		return
	}

	// Read the content of the shell config file
	content, err := os.ReadFile(shellConfigFile)
	if err != nil {
		return
	}

	// Check if the condition already exists in the file
	if !strings.Contains(string(content), condition) {
		// Append the condition to the file
		content = append(content, []byte("\n"+condition)...)

		// Write the updated content back to the file
		err = os.WriteFile(shellConfigFile, content, 0644)
		if err != nil {
			return
		}
	}
}

// expandTilde expands the tilde (~) in the file path to the user's home directory
func expandTilde(path string) string {
	if len(path) > 0 && path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return homeDir + path[1:]
	}
	return path
}

// TODO: uncomment following if enabling ngrok like feature
/*
func getCompPort(portFlag *ConnectCmdOpts, component *models.Component, deploymentTrack models.DeploymentTrack, orgId string) error {
	if portFlag.PreviewUrl {
		return nil
	}
	if strings.HasSuffix(strings.ToLower(component.DisplayType), "proxy") {
		return nil
	}
	if portFlag != nil && len(portFlag.Ports) > 0 {
		return nil
	}

	if strings.HasSuffix(strings.ToLower(component.DisplayType), "service") || strings.HasSuffix(strings.ToLower(component.DisplayType), "webhook") {
		endpointsSpinner := utils.CreateSpinner(i18n.T(" Fetching endpoints..."), "")
		endpointsSpinner.Start()
		endpoints, err := auth.ProjectClient.GetComponentEndpoints(component.Id, deploymentTrack.Id, orgId)
		endpointsSpinner.Stop()
		if err != nil {
			utils.PrintError("failed to fetch endpoints of service")
			ports, err := promptForPort()
			if err != nil {
				return err
			}
			portFlag.Ports = ports
		}

		if len(*endpoints) == 0 {
			utils.PrintError("no endpoint found for the service component")
			ports, err := promptForPort()
			if err != nil {
				return err
			}
			portFlag.Ports = ports
		}

		for _, item := range *endpoints {
			portFlag.Ports = append(portFlag.Ports, item.Port)
		}
	} else if strings.HasSuffix(strings.ToLower(component.DisplayType), "webapp") {
		ports, err := promptForPort()
		if err != nil {
			return err
		}
		portFlag.Ports = ports
	}
	return nil
}

func promptForPort() ([]int, error) {
	tInput := ""
	err := prompt.NewPromptInputMessage(
		prompt.PromptInputOpts{
			Message:  i18n.T("Port:"),
			Default:  "8080",
			Validate: prompt.ValidateNotEmptyInt,
		},
		&tInput,
	).Prompt()
	if err != nil {
		return []int{}, err
	}

	nInput, err := strconv.Atoi(tInput)
	if err != nil {
		return []int{}, err
	}

	return []int{nInput}, nil
}
*/

func getEndpointsOfProxyAgent(ctx context.Context, selectedOrg *api.Organization, selectedProject *models.Project, selectedEnv *project.ProjectEnvironment, allComps *[]models.Component, recreate bool, deleteBridge bool) (*models.Component, *[]project.Endpoint, error) {
	imageUrl := fmt.Sprintf("%s/%s:%s", constants.WSO2IP_AZURECR, BRIDGE_IMAGE_NAME, BRIDGE_IMAGE_VERSION)
	componentVersion := fmt.Sprintf("v%s", BRIDGE_IMAGE_VERSION)

	var proxyAgentComp *models.Component
	for _, componentItem := range *allComps {
		if componentItem.Name == BRIDGE_COMP_NAME {
			if componentItem.IsSystemComponent != "true" {
				return nil, nil, fmt.Errorf("unable to proceed as %s component already exists. Please delete it and try again", BRIDGE_COMP_NAME)
			}
			if componentItem.Version == componentVersion {
				if recreate || deleteBridge {
					delCmpSpinner := utils.CreateSpinner("Deleting previous bridge", "")
					delCmpSpinner.Start()
					auth.ComponentClient.DeleteComponent(selectedOrg.ID, selectedOrg.Handle, componentItem.Id, selectedProject.ID)
					delCmpSpinner.Stop()

					if deleteBridge {
						utils.PrintInfo("%s", i18n.T("\nPrevious bridge deleted successfully.\n"))
						return nil, nil, nil
					}
				} else {
					proxyAgentComp = &componentItem
				}
				break
			} else {
				delCmpSpinner := utils.CreateSpinner("Deleting previous bridge", "")
				delCmpSpinner.Start()
				auth.ComponentClient.DeleteComponent(selectedOrg.ID, selectedOrg.Handle, componentItem.Id, selectedProject.ID)
				delCmpSpinner.Start()

				if deleteBridge {
					utils.PrintInfo("%s", i18n.T("\nPrevious bridge deleted successfully.\n"))
					return nil, nil, nil
				}
			}
		}
	}

	if deleteBridge {
		utils.PrintInfo("%s", i18n.T("\nExiting as previous bridge was not found.\n"))
		return nil, nil, nil
	}

	if proxyAgentComp == nil {
		getComponentSpinner := utils.CreateSpinner(i18n.T(" Fetching container registries..."), "")
		getComponentSpinner.Start()
		containerRegs, err := auth.DevopsClient.GetContainerRegistries(selectedOrg.ID, selectedOrg.UUID)
		getComponentSpinner.Stop()
		if err != nil {
			return nil, nil, err
		}

		var registryItem *devops.ContainerRegistry
		for _, item := range containerRegs {
			if item.Host == constants.WSO2IP_AZURECR {
				registryItem = &item
				break
			}
		}

		if registryItem == nil {
			getComponentSpinner := utils.CreateSpinner(i18n.T(" Registering container samples registry..."), "")
			getComponentSpinner.Start()
			err := auth.DevopsClient.RegisterNewContainerRegistry(selectedOrg.ID, selectedOrg.UUID)
			getComponentSpinner.Stop()
			if err != nil {
				return nil, nil, err
			}

			getComponentSpinner.Start()
			containerRegs, err = auth.DevopsClient.GetContainerRegistries(selectedOrg.ID, selectedOrg.UUID)
			getComponentSpinner.Stop()
			if err != nil {
				return nil, nil, err
			}

			for _, item := range containerRegs {
				if item.Host == constants.WSO2IP_AZURECR {
					registryItem = &item
					break
				}
			}

			if registryItem == nil {
				return nil, nil, fmt.Errorf("%s", i18n.T("Failed to find registry"))
			}
		}

		createProxyComp := utils.CreateSpinner(i18n.T(" Creating bridge..."), "")
		createProxyComp.Start()
		err = auth.ComponentClient.CreateByoiComponent(
			selectedOrg.ID,
			BRIDGE_COMP_NAME,
			"Connect Bridge",
			"System component to connect local environment with deployed project environment",
			selectedProject.ID,
			"byoiService",
			componentVersion,
			imageUrl,
			registryItem.Id,
			true,
		)
		createProxyComp.Stop()
		if err != nil {
			return nil, nil, err
		}

		*allComps, err = common.GetComponentsWithSystemCompsForProject(selectedOrg, selectedProject)
		if err != nil {
			return nil, nil, err
		}

		for _, componentItem := range *allComps {
			if componentItem.Name == BRIDGE_COMP_NAME && componentItem.IsSystemComponent == "true" && componentItem.Version == componentVersion {
				proxyAgentComp = &componentItem
				break
			}
		}

		if proxyAgentComp == nil {
			return nil, nil, fmt.Errorf("%s", i18n.T("Failed to find bridge component after creating. Please try again."))
		}
	}

	deploymentTrack, err := common.ResolveDeploymentTrack(proxyAgentComp.DeploymentTracks, "")
	if err != nil {
		utils.HandleErr(fmt.Errorf(i18n.T("Error resolving deployment track: %w"), err))
	}

	remoteComWithRepoData, err := common.GetComponentWithRepoData(selectedOrg.ID, proxyAgentComp.Handler, selectedProject.ID)
	if err != nil {
		return proxyAgentComp, nil, err
	}

	matchingAppEnv, err := common.GetReleaseEnvForDeploymentTrack(remoteComWithRepoData, deploymentTrack.Id, selectedEnv.ID)
	if err != nil {
		return proxyAgentComp, nil, err
	}

	endpointsSpinner := utils.CreateSpinner(i18n.T(" Fetching endpoints of bridge component..."), "")
	endpointsSpinner.Start()
	endpoints, err := auth.ProjectClient.GetComponentEndpoints(proxyAgentComp.Id, deploymentTrack.Id, selectedOrg.ID)
	if err != nil {
		return proxyAgentComp, nil, err
	}
	endpointsSpinner.Stop()

	if len(*endpoints) == 0 {
		createEndppointsSpinner := utils.CreateSpinner(i18n.T(" Creating endpoints for bridge component..."), "")
		createEndppointsSpinner.Start()
		createEndpointsReq := devops.CreateByoiEndpointsRequest{
			Main: EncodeOpenAPIToBase64(endpointsYamlSpec),
			ApiSchemas: []devops.CreateApiSchemaRequest{
				{Filename: "asyncapi.yaml", Content: EncodeOpenAPIToBase64(asyncApiSpec)},
				{Filename: "openapi.yaml", Content: EncodeOpenAPIToBase64(openApiSpec)},
			},
		}
		err := auth.DevopsClient.CreateByoiEndpoints(selectedOrg.ID, selectedOrg.UUID, proxyAgentComp.Id, matchingAppEnv.ReleaseId, selectedProject.ID, createEndpointsReq)
		createEndppointsSpinner.Stop()
		if err != nil {
			return proxyAgentComp, nil, err
		}
	}

	deploymentStatusSpinner := utils.CreateSpinner(fmt.Sprintf("%s", i18n.T(" Fetching deployment status of bridge component...")), "")
	deploymentStatusSpinner.Start()
	componentDeployment, err := auth.ComponentClient.GetComponentDeployment(
		selectedOrg.Handle,
		selectedOrg.UUID,
		selectedOrg.ID,
		proxyAgentComp.Id,
		deploymentTrack.Id,
		selectedEnv.ID)
	deploymentStatusSpinner.Stop()
	if err != nil && !errors.Is(err, api.ErrNotFound) {
		return proxyAgentComp, nil, err
	}

	activeCount := 0
	for _, item := range *endpoints {
		if item.State == "Active" {
			activeCount++
		}
	}

	if errors.Is(err, api.ErrNotFound) || componentDeployment.DeploymentStatusV2 != "ACTIVE" || activeCount < len(*endpoints) {
		deploymentSpinner := utils.CreateSpinner(fmt.Sprintf("%s", i18n.T(" Deploying bridge component...")), "")
		deploymentSpinner.Start()
		resp, err := auth.ComponentClient.DeployByoiComponent(
			selectedOrg.ID,
			proxyAgentComp.Id,
			matchingAppEnv.ReleaseId,
			imageUrl)
		deploymentSpinner.Stop()
		if err != nil {
			return proxyAgentComp, nil, err
		}
		if !resp {
			return proxyAgentComp, nil, fmt.Errorf("%s", i18n.T("Failed to deploy bridge component"))
		}

		deploymentStatusSpinner := utils.CreateSpinner(fmt.Sprintf("%s", i18n.T(" Checking deployment status of bridge component...")), "")
		deploymentStatusSpinner.Start()
		err = waitForDeploy(ctx, selectedOrg, proxyAgentComp, selectedEnv, deploymentTrack)
		deploymentStatusSpinner.Stop()
		if err != nil {
			return proxyAgentComp, nil, err
		}

		endpointsSpinner := utils.CreateSpinner(i18n.T(" Refetching endpoints of bridge component..."), "")
		endpointsSpinner.Start()
		endpoints, err = waitActiveEndpoints(ctx, selectedOrg, proxyAgentComp, deploymentTrack)
		endpointsSpinner.Stop()
		if err != nil {
			return proxyAgentComp, nil, err
		}
	}

	return proxyAgentComp, endpoints, nil
}

func waitForDeploy(ctx context.Context, selectedOrg *api.Organization, proxyAgentComp *models.Component, selectedEnv *project.ProjectEnvironment, deploymentTrack *models.DeploymentTrack) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err() // Handle cancellation
		case <-ticker.C:
			componentDeployment, err := auth.ComponentClient.GetComponentDeployment(
				selectedOrg.Handle,
				selectedOrg.UUID,
				selectedOrg.ID,
				proxyAgentComp.Id,
				deploymentTrack.Id,
				selectedEnv.ID)
			if err != nil && !errors.Is(err, api.ErrNotFound) {
				return err
			}
			if componentDeployment.DeploymentStatusV2 == "ACTIVE" {
				return nil // Success!
			}
		}
	}
}

func waitActiveEndpoints(ctx context.Context, selectedOrg *api.Organization, proxyAgentComp *models.Component, deploymentTrack *models.DeploymentTrack) (*[]project.Endpoint, error) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err() // Handle cancellation
		case <-ticker.C:
			endpoints, err := auth.ProjectClient.GetComponentEndpoints(proxyAgentComp.Id, deploymentTrack.Id, selectedOrg.ID)
			if err != nil {
				return nil, err
			}
			activeCount := 0
			for _, item := range *endpoints {
				if item.State == "Active" {
					activeCount++
				} else if item.State == "Error" {
					return nil, fmt.Errorf("%s", fmt.Sprintf("%s, Details:%s, Code:%s, WorkerId:%s ",
						item.StateReason.Message,
						item.StateReason.Details,
						item.StateReason.Code,
						item.StateReason.WorkerID,
					))
				}
			}
			if activeCount == len(*endpoints) {
				return endpoints, nil
			}
		}
	}
}

func getProxyAgentRestEp(selectedOrg *api.Organization, endpoints *[]project.Endpoint, selectedEnv *project.ProjectEnvironment) (*project.Endpoint, error) {
	var proxyRestEndpoint *project.Endpoint
	for _, endpoint := range *endpoints {
		if endpoint.EnvironmentID == selectedEnv.ID && endpoint.Type == "REST" {
			proxyRestEndpoint = &endpoint
			break
		}
	}

	if proxyRestEndpoint == nil {
		return nil, fmt.Errorf("%s", i18n.T("Failed to find endpoint"))
	}

	if proxyRestEndpoint.State != "Active" {
		return nil, fmt.Errorf("%s", i18n.T("bridge component REST endpoint is not active"))
	}

	return proxyRestEndpoint, nil
}

func getProxyAgentRestEpKey(selectedOrg *api.Organization, endpoint *project.Endpoint, selectedEnv *project.ProjectEnvironment) (*apimpublisher.ApiKeyResponse, error) {
	restApiKeyCache.mu.RLock()
	cachedResponse := restApiKeyCache.response
	lastFetchTime := restApiKeyCache.lastFetch
	restApiKeyCache.mu.RUnlock()

	if cachedResponse != nil && time.Since(lastFetchTime) < time.Duration(cachedResponse.ValidityTime-20)*time.Second {
		return cachedResponse, nil // Return cached value
	}

	newResponse, err := create.GetApiKeyForEndpoint(endpoint.APIMID, selectedOrg, *selectedEnv)
	if err != nil {
		return nil, err
	}

	restApiKeyCache.mu.Lock()
	restApiKeyCache.response = newResponse
	restApiKeyCache.lastFetch = time.Now()
	restApiKeyCache.mu.Unlock()

	return newResponse, nil
}

func getProxyAgentWSEpKey(selectedOrg *api.Organization, endpoint *project.Endpoint, selectedEnv *project.ProjectEnvironment) (*apimpublisher.ApiKeyResponse, error) {
	wsApiKeyCache.mu.RLock()
	cachedResponse := wsApiKeyCache.response
	lastFetchTime := wsApiKeyCache.lastFetch
	wsApiKeyCache.mu.RUnlock()

	if cachedResponse != nil && time.Since(lastFetchTime) < time.Duration(cachedResponse.ValidityTime-20)*time.Second {
		return cachedResponse, nil // Return cached value
	}

	newResponse, err := create.GetApiKeyForEndpoint(endpoint.APIMID, selectedOrg, *selectedEnv)
	if err != nil {
		return nil, err
	}

	wsApiKeyCache.mu.Lock()
	wsApiKeyCache.response = newResponse
	wsApiKeyCache.lastFetch = time.Now()
	wsApiKeyCache.mu.Unlock()

	return newResponse, nil
}

func getProxyAgentWsEp(selectedOrg *api.Organization, endpoints *[]project.Endpoint, selectedEnv *project.ProjectEnvironment) (*project.Endpoint, error) {
	var proxyWsEndpoint *project.Endpoint
	for _, endpoint := range *endpoints {
		if endpoint.EnvironmentID == selectedEnv.ID && endpoint.Type == "WS" {
			proxyWsEndpoint = &endpoint
			break
		}
	}

	if proxyWsEndpoint == nil {
		return nil, fmt.Errorf("%s", i18n.T("Failed to find endpoint"))
	}

	if proxyWsEndpoint.State != "Active" {
		return nil, fmt.Errorf("%s", i18n.T("bridge component WS endpoint is not active"))
	}

	return proxyWsEndpoint, nil
}

func injectConnectionEnvs(
	org *api.Organization,
	project *models.Project,
	remoteComponent *models.Component,
	selectedEnv *project.ProjectEnvironment,
	conns []connections.Connection,
	envVars *map[string]string,
	secureHosts *map[string]bool,
	skippedConns []string,
) error {
	secretGroupMap := map[string][]configmapping.ResolveSecretsReqItem{}
	secretGroupEnvMap := map[string]map[string]string{}

	for _, connItem := range conns {
		if slices.Contains(skippedConns, connItem.Name) {
			continue
		}
		if _, ok := secretGroupEnvMap[connItem.GroupUuid]; !ok {
			secretGroupEnvMap[connItem.GroupUuid] = make(map[string]string)
		}

		connDetails, err := common.GetConnectionItem(org.ID, connItem.GroupUuid)
		if err != nil {
			return err
		}
		matchingConfig := connDetails.Configurations[selectedEnv.TemplateId]
		if matchingConfig.EnvironmentUuid != "" {
			for _, entry := range matchingConfig.Entries {
				if entry.Value != "" {
					if strings.HasPrefix(entry.Value, "https://") {
						parsedURL, _ := url.Parse(entry.Value)
						(*secureHosts)[parsedURL.Host] = true
						(*envVars)[entry.EnvVariableName] = strings.Replace(entry.Value, "https://", "http://", 1)
					} else {
						(*envVars)[entry.EnvVariableName] = entry.Value
					}
				} else if entry.IsSensitive && !entry.IsFile && (*envVars)[entry.EnvVariableName] == "" {
					secretGroupMap[connItem.GroupUuid] = append(secretGroupMap[connItem.GroupUuid], configmapping.ResolveSecretsReqItem{ValueRef: entry.ValueRef, Key: entry.Key})
					secretGroupEnvMap[connItem.GroupUuid][entry.Key] = entry.EnvVariableName
				}
			}
		}
	}

	if len(secretGroupMap) > 0 {
		for groupId, secretGroup := range secretGroupMap {
			getSecretsSpinner := utils.CreateSpinner(i18n.T(" Fetching secure configurations..."), "")
			getSecretsSpinner.Start()
			componentId := ""
			if remoteComponent != nil {
				componentId = remoteComponent.Id
			}
			resp, err := auth.ConfigMappingApiClient.ResolveSecrets(org.ID, groupId, configmapping.ResolveSecretsReq{
				ProjectID:     project.ID,
				ComponentID:   componentId,
				EnvTemplateID: selectedEnv.TemplateId,
				Secrets:       secretGroup,
			})
			getSecretsSpinner.Stop()
			if err != nil {
				return err
			}

			for _, secretItem := range resp.Secrets {
				envKey := secretGroupEnvMap[groupId][secretItem.Key]
				(*envVars)[getInjectedEnvVarNames(envKey)] = secretItem.Value
			}
		}
	}

	return nil
}

func getInjectedEnvVarNames(key string) string {
	parts := strings.Split(key, "_")

	if len(parts) > 1 {
		lastPart := parts[len(parts)-1]
		lastPart, _ = strings.CutPrefix(lastPart, "WIP")
		parts[len(parts)-1] = lastPart
	}

	return strings.Join(parts, "_")
}
