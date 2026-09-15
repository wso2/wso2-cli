package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/devops"
	"github.com/wso2/integration-platform-tools/pkg/util/constants"
)

// codeServerSampleName is how the Code Server image is listed in the samples
// registry. The image URL is looked up by this name rather than hardcoded.
const codeServerSampleName = "Code Server"

// IsCloudEditor reports whether the server is running inside the cloud editor.
// BASE_URL is injected only in that environment; the same signal gates the
// Source Control UI git instructions in get_started.
func IsCloudEditor() bool {
	return os.Getenv("BASE_URL") != ""
}

// BindCodeServer binds the running cloud editor instance to a component, so the
// editor knows which integration it is working on. Mirrors the clirpc handler
// component/updateCodeServer, which is what the VS Code extension calls.
//
// Only meaningful in the cloud editor; callers should gate on IsCloudEditor().
// sourceCommitHash may be empty — at creation time there is no build commit yet.
func BindCodeServer(ctx context.Context, org *api.Organization, projectID, componentID, sourceCommitHash string) error {
	if org == nil {
		return fmt.Errorf("organization is required")
	}
	if projectID == "" || componentID == "" {
		return fmt.Errorf("project id and component id are required")
	}

	// The IDP subject is part of the mutation input. In stdio mode this comes
	// from the stored session, the same way org/impl.go and auth/impl.go read it.
	userInfo, err := auth.GetCurrentUser()
	if err != nil {
		return fmt.Errorf("resolving current user: %w", err)
	}

	registry, err := resolveContainerRegistry(ctx, org.ID, org.UUID)
	if err != nil {
		return err
	}

	imageURL, err := resolveCodeServerImage(ctx, org.ID, org.UUID, projectID)
	if err != nil {
		return err
	}

	return GetGraphQLClient(ctx).UpdateCodeServer(
		org.ID,
		userInfo.IDPId,
		org.UUID,
		projectID,
		componentID,
		org.Handle,
		imageURL,
		registry.Id,
		sourceCommitHash,
	)
}

// resolveContainerRegistry finds the platform container registry for the org,
// registering it first if the org does not have one yet.
//
// Unlike the clirpc equivalent, failures are returned rather than swallowed as
// (nil, nil) — a nil registry would otherwise surface later as a confusing nil
// dereference instead of the actual cause.
func resolveContainerRegistry(ctx context.Context, orgID, orgUUID string) (*devops.ContainerRegistry, error) {
	devopsClient := GetDevopsClient(ctx)

	find := func() (*devops.ContainerRegistry, error) {
		registries, err := devopsClient.GetContainerRegistries(orgID, orgUUID)
		if err != nil {
			return nil, fmt.Errorf("listing container registries: %w", err)
		}
		for i := range registries {
			if registries[i].Host == constants.WSO2IP_AZURECR {
				return &registries[i], nil
			}
		}
		return nil, nil
	}

	registry, err := find()
	if err != nil {
		return nil, err
	}
	if registry != nil {
		return registry, nil
	}

	// Not registered yet for this org — register, then look again.
	if err := devopsClient.RegisterNewContainerRegistry(orgID, orgUUID); err != nil {
		return nil, fmt.Errorf("registering container registry: %w", err)
	}
	registry, err = find()
	if err != nil {
		return nil, err
	}
	if registry == nil {
		return nil, fmt.Errorf("container registry %s not available for the organization after registration", constants.WSO2IP_AZURECR)
	}
	return registry, nil
}

// resolveCodeServerImage looks up the Code Server image URL from the samples
// registry.
func resolveCodeServerImage(ctx context.Context, orgID, orgUUID, projectID string) (string, error) {
	samples, err := GetDevopsClient(ctx).GetSamples(orgID, orgUUID, projectID)
	if err != nil {
		return "", fmt.Errorf("listing samples: %w", err)
	}
	for i := range samples {
		if samples[i].Name == codeServerSampleName {
			return samples[i].ImageUrl, nil
		}
	}
	return "", fmt.Errorf("%q not found in the samples registry", codeServerSampleName)
}
