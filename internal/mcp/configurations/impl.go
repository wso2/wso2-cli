package configurations

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
	configmapping "github.com/wso2/integration-platform-tools/pkg/api/config-mapping"
)

// BallerinaConfigMountPath is the only path a Ballerina (WSO2 Integrator)
// integration reads its Config.toml from. A file mounted anywhere else is
// silently ignored by the runtime.
const BallerinaConfigMountPath = "/config/Config.toml"

func createConfigurations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create configurations."), nil
	}

	selectedProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create configurations."), nil
	}

	component, err := utils.GetTargetComponent(ctx, *targetOrg, *selectedProject, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create configurations."), nil
	}

	environmentId, err := utils.GetRequiredStringArgument(request, "environment_id")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create configurations."), nil
	}

	dTrack, err := utils.GetDeploymentTrack(*component, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create configurations."), nil
	}

	var appEnvId = ""
	for _, apiVersion := range component.ApiVersions {
		if apiVersion.ApiVersion == dTrack.ApiVersion {
			for _, env := range apiVersion.AppEnvVersions {
				if env.EnvironmentId == environmentId {
					appEnvId = env.ReleaseId
					break
				}
			}
			break
		}
	}
	if appEnvId == "" {
		return utils.NewMCPErrorResponse(fmt.Errorf("App Environment ID is not found. Deploy the component before creating configurations"), "Failed to create configurations."), nil
	}

	configType, err := utils.GetRequiredStringArgument(request, "configuration_type")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create configurations."), nil
	}

	mountType, err := utils.GetRequiredStringArgument(request, "mount_type")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create configurations."), nil
	}
	// The tool schema previously advertised "env variables" while the canonical
	// value is common.MountTypeEnv ("env variable"). mcp-go does not validate
	// enums server-side, so the plural reached this handler, matched neither
	// branch below, and the configuration was silently dropped while the tool
	// still reported success and restarted the component. Normalise here, and
	// reject anything unrecognised rather than falling through.
	mountType = normalizeMountType(mountType)
	if mountType != common.MountTypeEnv && mountType != common.MountTypeFile {
		return utils.NewMCPErrorResponse(
			fmt.Errorf("invalid mount_type %q: must be %q or %q",
				mountType, common.MountTypeEnv, common.MountTypeFile),
			"Failed to create configurations."), nil
	}

	environmentTemplateId, err := utils.GetRequiredStringArgument(request, "environment_template_id")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get environment template id", err), nil
	}

	env_variables := utils.GetOptionalStringArgument(request, "env_variables", "")
	var envVariablesParsed []common.KeyValOpt
	if env_variables != "" {
		err = json.Unmarshal([]byte(env_variables), &envVariablesParsed)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to parse environment variables."), nil
		}
	}

	fileMountPath := utils.GetOptionalStringArgument(request, "file_mount_path", "")

	fileMountContent := utils.GetOptionalStringArgument(request, "file_mount_content", "")

	configMapplingApiClient := utils.GetConfigMappingApiClient(ctx)

	configMappinng, err := configMapplingApiClient.GetConfigMapping(targetOrg.ID, selectedProject.ID,
		component.Id, environmentTemplateId, dTrack.Id)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get deployed configurations", err), nil
	}
	var configs []configmapping.ConfigMapping
	if len(configMappinng.Configurations) > 0 {
		for _, config := range configMappinng.Configurations {
			if len(config.Values) == 0 {
				return mcp.NewToolResultErrorFromErr("configuration values are not found", fmt.Errorf("configuration %s has no values", config.Key)), nil
			}
			configMappinng := configmapping.ConfigMapping{
				Key:         config.Key,
				IsSensitive: config.IsSensitive,
				IsFile:      config.IsFile,
				IsDynamic:   config.IsDynamic,
				Values: []configmapping.ConfigMappingValue{
					{
						Value:           config.Values[0].Value,
						EnvironmentUUID: config.Values[0].EnvironmentUUID,
					},
				},
			}
			configs = append(configs, configMappinng)
		}
	}
	if mountType == common.MountTypeEnv {
		if len(envVariablesParsed) == 0 {
			return mcp.NewToolResultError("env_variables is required when mount_type is env"), nil
		}
		for _, item := range envVariablesParsed {
			configs = append(configs, configmapping.ConfigMapping{
				Key:         item.Key,
				IsSensitive: configType == common.ConfigTypeSecret,
				IsFile:      false,
				IsDynamic:   true,
				Values: []configmapping.ConfigMappingValue{
					{
						Value:           item.Val,
						EnvironmentUUID: environmentTemplateId,
					},
				},
			})
		}
	}

	nextSteps := []string{
		"Configuration has been applied and the component will automatically restart to apply the changes.",
		"Use `get_deployment` to get the deployment status and wait for it to be ready.",
		"Use `get_logs` with `log_type`: 'application' to ensure the changes are applied successfully.",
	}

	// Warn when the mount type does not match how the buildpack reads
	// configuration. Ballerina reads configurable values from Config.toml, so
	// env variables are ignored; Micro Integrator reads env variables and
	// ignores Config.toml. Either mismatch applies cleanly and then does
	// nothing at runtime, which is very hard to diagnose from the deployment
	// status alone.
	if isBallerinaComponent(component.DisplayType) && mountType == common.MountTypeEnv {
		nextSteps = append(nextSteps,
			"WARNING: this is a WSO2 Integrator integration on the default Ballerina profile, which reads configurable values from Config.toml. "+
				"Values supplied as environment variables are typically not picked up. Re-apply with mount_type 'file mount', "+
				"file_mount_path '"+BallerinaConfigMountPath+"' and a Config.toml body, or set them in the web console via the "+
				"Configure form on the integration overview page.")
	}
	if isMicroIntegratorComponent(component.DisplayType) && mountType == common.MountTypeFile {
		nextSteps = append(nextSteps,
			"WARNING: this is a WSO2 Integrator integration on the Micro Integrator (MI) profile, which reads configuration from environment variables. "+
				"A mounted file is typically not picked up. Re-apply with mount_type 'env variable'.")
	}
	// A Ballerina file mount at any other path is silently ignored by the
	// runtime — it only reads Config.toml from BallerinaConfigMountPath.
	if isBallerinaComponent(component.DisplayType) && mountType == common.MountTypeFile &&
		fileMountPath != BallerinaConfigMountPath {
		nextSteps = append(nextSteps,
			"WARNING: file_mount_path is '"+fileMountPath+"', but a WSO2 Integrator integration on the default Ballerina profile only reads its "+
				"configuration from '"+BallerinaConfigMountPath+"'. A file mounted elsewhere will not be picked up.")
	}
	// Secret values sent through a tool call pass through the conversation and
	// may be retained in logs or transcripts. The console's Configure form keeps
	// them write-only.
	if configType == common.ConfigTypeSecret {
		nextSteps = append(nextSteps,
			"NOTE: this was created as a secret. Prefer setting sensitive values in the web console via the Configure form on "+
				"the integration overview page — values set there are write-only, whereas anything passed through this tool has "+
				"already travelled through the conversation. If this value is a real credential, treat it as exposed and rotate it.")
	}

	if mountType == common.MountTypeFile {
		if fileMountContent == "" {
			return mcp.NewToolResultError("file_mount_content is required when mount_type is file"), nil
		}
		if fileMountPath == "" {
			return mcp.NewToolResultError("file_mount_path is required when mount_type is file"), nil
		}
		configs = append(configs, configmapping.ConfigMapping{
			Key:         fileMountPath,
			IsSensitive: configType == common.ConfigTypeSecret,
			IsFile:      true,
			IsDynamic:   true,
			Values: []configmapping.ConfigMappingValue{
				{
					Value:           fileMountContent,
					EnvironmentUUID: environmentTemplateId,
				},
			},
		})
	}
	err = configMapplingApiClient.CreateConfigMapping(
		targetOrg.ID,
		selectedProject.ID,
		component.Id,
		environmentTemplateId,
		dTrack.Id,
		configs,
	)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to create configuration", err), nil
	}

	return mcp.NewToolResultText("Configuration created successfully. Trigger a redeploy application to reflect the changes."), nil
}

// normalizeMountType maps the accepted spellings of mount_type onto the
// canonical constants. Agents may send either the singular form from the tool
// description or the plural that the schema enum used to advertise.
func normalizeMountType(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "env variable", "env variables", "env":
		return common.MountTypeEnv
	case "file mount", "file":
		return common.MountTypeFile
	default:
		return v
	}
}

// isBallerinaComponent reports whether the component uses the Ballerina
// (WSO2 Integrator) buildpack. Matches the DisplayType prefix convention used
// elsewhere in the codebase, e.g. internal/cmd/build/create/impl.go.
func isBallerinaComponent(displayType string) bool {
	return strings.HasPrefix(strings.ToLower(displayType), "ballerina")
}

// isMicroIntegratorComponent reports whether the component uses the WSO2 Micro
// Integrator buildpack (displayType values such as miRestApi, miEventHandler).
func isMicroIntegratorComponent(displayType string) bool {
	return strings.HasPrefix(strings.ToLower(displayType), "mi")
}
