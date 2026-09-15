package buildpack

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/devops"
)

// isSupportedBuildpack returns true only for Ballerina and WSO2 Micro Integrator buildpacks.
func isSupportedBuildpack(bp devops.BuildPack) bool {
	lang := strings.ToLower(bp.Language)
	display := strings.ToLower(bp.DisplayName)
	isBal := strings.Contains(lang, "ballerina") || strings.Contains(display, "ballerina")
	isWSO2MI := strings.Contains(lang, "wso2") || strings.Contains(display, "wso2") ||
		strings.Contains(lang, "micro integrator") || strings.Contains(display, "micro integrator")
	return isBal || isWSO2MI
}

func getBuildPacks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve buildpacks."), nil
	}

	componentType := utils.GetArgument(request, "type")
	if componentType == nil {
		return utils.NewMCPErrorResponse(fmt.Errorf("Component type is required"), "Failed to retrieve buildpacks."), nil
	}
	componentTypeStr := componentType.(string)
	if componentTypeStr == "" {
		return utils.NewMCPErrorResponse(fmt.Errorf("Component type cannot be empty"), "Failed to retrieve buildpacks."), nil
	}

	allBuildPacks, err := utils.GetDevopsClient(ctx).GetBuildPackOptions(targetOrg.UUID, targetOrg.ID, componentTypeStr)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve buildpacks."), nil
	}

	supported := make([]devops.BuildPack, 0, len(allBuildPacks))
	for _, bp := range allBuildPacks {
		if isSupportedBuildpack(bp) {
			supported = append(supported, bp)
		}
	}

	nextSteps := []string{
		"Use the `buildpack_id` and `buildpack_language_version` from this list when calling `create_integration`.",
		"Only Ballerina and WSO2 Micro Integrator buildpacks are supported on this platform.",
		"Note: `language` in the response is case-sensitive and must be used exactly as returned.",
	}
	return utils.NewMCPResponse(supported, "Buildpacks retrieved successfully.", nextSteps)
}
