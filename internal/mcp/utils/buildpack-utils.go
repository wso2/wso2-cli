package utils

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/devops"
)

func GetTargetBuildPack(ctx context.Context, targetOrg api.Organization, componentType string, request mcp.CallToolRequest) (*devops.BuildPack, error) {
	buildPackId := GetArgument(request, "buildpack_id")
	if buildPackId == nil {
		return nil, fmt.Errorf("buildpack_id is required")
	}
	if buildPackId == "" {
		return nil, fmt.Errorf("buildpack_id cannot be empty")
	}

	buildPackIdStr := buildPackId.(string)
	availableBuildPacks, err := GetDevopsClient(ctx).GetBuildPackOptions(targetOrg.UUID, targetOrg.ID, componentType)
	if err != nil {
		return nil, fmt.Errorf("failed to get build pack options: %w", err)
	}

	for _, buildPack := range availableBuildPacks {
		if buildPack.ID == buildPackIdStr {
			return &buildPack, nil
		}
	}

	return nil, fmt.Errorf("build pack with id %s not found", buildPackIdStr)
}
