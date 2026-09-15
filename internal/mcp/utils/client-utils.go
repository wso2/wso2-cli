package utils

import (
	"context"

	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/pkg/api"
	apimpublisher "github.com/wso2/integration-platform-tools/pkg/api/apimPublisher"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	configmapping "github.com/wso2/integration-platform-tools/pkg/api/config-mapping"
	"github.com/wso2/integration-platform-tools/pkg/api/connections"
	"github.com/wso2/integration-platform-tools/pkg/api/credentials"
	"github.com/wso2/integration-platform-tools/pkg/api/customDomain"
	deploymentbuild "github.com/wso2/integration-platform-tools/pkg/api/deploymentBuild"
	"github.com/wso2/integration-platform-tools/pkg/api/devops"
	"github.com/wso2/integration-platform-tools/pkg/api/gqlbuild"
	"github.com/wso2/integration-platform-tools/pkg/api/logs"
	"github.com/wso2/integration-platform-tools/pkg/api/marketplace"
	platformservices "github.com/wso2/integration-platform-tools/pkg/api/platform-services"
	"github.com/wso2/integration-platform-tools/pkg/api/platformgit"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
	subscription "github.com/wso2/integration-platform-tools/pkg/api/subscriptions"
	userstore_mgt "github.com/wso2/integration-platform-tools/pkg/api/user-store-mgt"
	workflowmgt "github.com/wso2/integration-platform-tools/pkg/api/workflow-mgt"
)

// GetProjectClient returns a project client (context-aware or global)
func GetProjectClient(ctx context.Context) *project.ProjectClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetProjectClient()
	}
	// Stdio mode: use global client
	return auth.ProjectClient
}

// GetComponentClient returns a component client (context-aware or global)
func GetComponentClient(ctx context.Context) *component.ComponentClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetComponentClient()
	}
	// Stdio mode: use global client
	return auth.ComponentClient
}

// GetOrgClient returns an organization client (context-aware or global)
func GetOrgClient(ctx context.Context) *api.OrgClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetOrgClient()
	}
	// Stdio mode: use global client
	return auth.OrgClient
}

// GetGraphQLClient returns a GraphQL client (context-aware or global)
func GetGraphQLClient(ctx context.Context) *component.GraphQLClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetGraphQLClient()
	}
	// Stdio mode: use global client
	return auth.GraphQLClient
}

// GetGitClient returns a Git client (context-aware or global)
func GetGitClient(ctx context.Context) *platformgit.GitClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetGitClient()
	}
	// Stdio mode: use global client
	return auth.GitClient
}

// GetLogsClient returns a logs client (context-aware or global)
func GetLogsClient(ctx context.Context) *logs.LogsClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetLogsClient()
	}
	// Stdio mode: use global client
	return auth.LogsClient
}

// GetApimPublisherClient returns an APIM publisher client (context-aware or global)
func GetApimPublisherClient(ctx context.Context) *apimpublisher.ApimPublisherClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetApimPublisherClient()
	}
	// Stdio mode: use global client
	return auth.ApimPublisherClient
}

// GetDevopsClient returns a DevOps client (context-aware or global)
func GetDevopsClient(ctx context.Context) *devops.DevopsClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetDevopsClient()
	}
	// Stdio mode: use global client
	return auth.DevopsClient
}

// GetDeploymentBuildClient returns a deployment build client (context-aware or global)
func GetDeploymentBuildClient(ctx context.Context) *deploymentbuild.DeploymentBuildClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetDeploymentBuildClient()
	}
	// Stdio mode: use global client
	return auth.DeploymentBuildClient
}

// GetConnectionsClient returns a connections client (context-aware or global)
func GetConnectionsClient(ctx context.Context) *connections.ConnectionsClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetConnectionsClient()
	}
	// Stdio mode: use global client
	return auth.ConnectionsClient
}

// GetMarketplaceClient returns a marketplace client (context-aware or global)
func GetMarketplaceClient(ctx context.Context) *marketplace.MarketplaceClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetMarketplaceClient()
	}
	// Stdio mode: use global client
	return auth.MarketplaceClient
}

// GetGQLBuildClient returns a GQL build client (context-aware or global)
func GetGQLBuildClient(ctx context.Context) *gqlbuild.GQLBuildClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetGQLBuildClient()
	}
	// Stdio mode: use global client
	return auth.GQLBuildClient
}

// GetCredentialClient returns a credential client (context-aware or global)
func GetCredentialClient(ctx context.Context) *credentials.CredentialClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetCredentialClient()
	}
	// Stdio mode: use global client
	return auth.CredentialClient
}

// GetWorkflowMgtClient returns a workflow management client (context-aware or global)
func GetWorkflowMgtClient(ctx context.Context) *workflowmgt.WorkflowMgtClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetWorkflowMgtClient()
	}
	// Stdio mode: use global client
	return auth.WorkflowMgtClient
}

// GetSubscriptionsClient returns a subscriptions client (context-aware or global)
func GetSubscriptionsClient(ctx context.Context) *subscription.SubscriptionsClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetSubscriptionsClient()
	}
	// Stdio mode: use global client
	return auth.SubscriptionsClient
}

// GetConfigMappingApiClient returns a config mapping client (context-aware or global)
func GetConfigMappingApiClient(ctx context.Context) *configmapping.ConfigMappingSvcClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetConfigMappingApiClient()
	}
	// Stdio mode: use global client
	return auth.ConfigMappingApiClient
}

// GetPlatformServicesClient returns a platform services client (context-aware or global)
func GetPlatformServicesClient(ctx context.Context) *platformservices.PlatformServicesClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetPlatformServicesClient()
	}
	// Stdio mode: use global client
	return auth.PlatformServicesClient
}

// GetUserStoreMgtClient returns a user store management client (context-aware or global)
func GetUserStoreMgtClient(ctx context.Context) *userstore_mgt.UserStoreManagementClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetUserStoreMgtClient()
	}
	// Stdio mode: use global client
	return auth.UserStoreMgtClient
}

// GetCustomDomainClient returns a custom domain client (context-aware or global)
func GetCustomDomainClient(ctx context.Context) *customDomain.CustomDomainClient {
	// Check if we have a client factory in context (HTTP mode)
	clientFactory := GetClientFactoryFromContext(ctx)
	if clientFactory != nil {
		return clientFactory.GetCustomDomainClient()
	}
	// Stdio mode: use global client
	return auth.CustomDomainClient
}
