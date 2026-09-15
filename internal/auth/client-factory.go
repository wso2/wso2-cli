package auth

import (
	cfg "github.com/wso2/integration-platform-tools/internal/config"
	"github.com/wso2/integration-platform-tools/internal/region"
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

// ClientFactory creates API clients with request-specific token stores
type ClientFactory struct {
	config     *cfg.EnvConfig
	tokenStore api.ReadOnlyTokenStore
}

// NewClientFactory creates a new client factory with the given token store
func NewClientFactory(tokenStore api.ReadOnlyTokenStore) *ClientFactory {
	return &ClientFactory{
		config:     region.GetRegionConfig(),
		tokenStore: tokenStore,
	}
}

// GetOrgClient returns an organization client
func (cf *ClientFactory) GetOrgClient() *api.OrgClient {
	return api.NewOrgClient(cf.config.Apis.OrgsAPI, cf.tokenStore)
}

// GetProjectClient returns a project client
func (cf *ClientFactory) GetProjectClient() *project.ProjectClient {
	return project.NewProjectClient(
		cf.config.Apis.ProjectAPI,
		cf.config.BillingConsoleBaseUrl,
		cf.config.Apis.InsightsQueryApi,
		cf.tokenStore,
	)
}

// GetComponentClient returns a component client
func (cf *ClientFactory) GetComponentClient() *component.ComponentClient {
	return component.NewComponentClient(
		cf.config.Apis.ConfigMgtAPI,
		cf.config.Apis.ComponentAPI,
		cf.config.Apis.ComponentManageAPI,
		cf.config.Apis.DeclarativeAPI,
		cf.tokenStore,
	)
}

// GetGraphQLClient returns a GraphQL client
func (cf *ClientFactory) GetGraphQLClient() *component.GraphQLClient {
	return component.NewGraphQLClient(cf.config.Apis.GraphQLAPI, cf.tokenStore)
}

// GetGitClient returns a Git client
func (cf *ClientFactory) GetGitClient() *platformgit.GitClient {
	return platformgit.NewGitClient(
		cf.config.Apis.ProjectAPI,
		cf.config.Apis.ComponentUtilAPI,
		&cf.config.GhApp,
		cf.tokenStore,
	)
}

// GetLogsClient returns a logs client
func (cf *ClientFactory) GetLogsClient() *logs.LogsClient {
	return logs.NewLogsClient(cf.config.SysApiPrefixUrl, cf.tokenStore)
}

// GetApimPublisherClient returns an APIM publisher client
func (cf *ClientFactory) GetApimPublisherClient() *apimpublisher.ApimPublisherClient {
	return apimpublisher.NewApimPublisherClient(cf.config.ApimPublisherBaseUrl, cf.tokenStore)
}

// GetDevopsClient returns a DevOps client
func (cf *ClientFactory) GetDevopsClient() *devops.DevopsClient {
	return devops.NewDevopsClient(cf.config.DevopsApiBaseUrl, cf.tokenStore)
}

// GetDeploymentBuildClient returns a deployment build client
func (cf *ClientFactory) GetDeploymentBuildClient() *deploymentbuild.DeploymentBuildClient {
	return deploymentbuild.NewDeploymentBuildClient(
		cf.config.Apis.DeclarativeAPI,
		cf.config.Apis.ProxyDeployEp,
		cf.config.Apis.GraphQLAPI,
		cf.tokenStore,
	)
}

// GetConnectionsClient returns a connections client
func (cf *ClientFactory) GetConnectionsClient() *connections.ConnectionsClient {
	return connections.NewConnectionsClient(cf.config.Apis.ConnectionsAPI, cf.tokenStore)
}

// GetMarketplaceClient returns a marketplace client
func (cf *ClientFactory) GetMarketplaceClient() *marketplace.MarketplaceClient {
	return marketplace.NewMarketplaceClient(cf.config.Apis.MarketplaceAPI, cf.tokenStore)
}

// GetGQLBuildClient returns a GQL build client
func (cf *ClientFactory) GetGQLBuildClient() *gqlbuild.GQLBuildClient {
	return gqlbuild.NewGQLBuildClient(cf.config.Apis.ComponentAPI, cf.tokenStore)
}

// GetCredentialClient returns a credential client
func (cf *ClientFactory) GetCredentialClient() *credentials.CredentialClient {
	return credentials.NewCredentialClient(cf.config.Apis.ComponentAPI, cf.tokenStore)
}

// GetWorkflowMgtClient returns a workflow management client
func (cf *ClientFactory) GetWorkflowMgtClient() *workflowmgt.WorkflowMgtClient {
	return workflowmgt.NewWorkflowMgtClient(cf.config.Apis.WorkflowMgtAPI, cf.tokenStore)
}

// GetSubscriptionsClient returns a subscriptions client
func (cf *ClientFactory) GetSubscriptionsClient() *subscription.SubscriptionsClient {
	return subscription.NewSUbscriptionsClient(cf.config.BillingConsoleBaseUrl, cf.tokenStore)
}

// GetConfigMappingApiClient returns a config mapping client
func (cf *ClientFactory) GetConfigMappingApiClient() *configmapping.ConfigMappingSvcClient {
	return configmapping.NewConfigMappingSvcClient(
		cf.config.Apis.ConfigMappingAPI,
		cf.config.Apis.ConfigAPI,
		cf.tokenStore,
	)
}

// GetPlatformServicesClient returns a platform services client
func (cf *ClientFactory) GetPlatformServicesClient() *platformservices.PlatformServicesClient {
	return platformservices.NewPlatformServicesClient(cf.config.Apis.PlatformServicesAPI, cf.tokenStore)
}

// GetUserStoreMgtClient returns a user store management client
func (cf *ClientFactory) GetUserStoreMgtClient() *userstore_mgt.UserStoreManagementClient {
	return userstore_mgt.NewUserStoreManagementClient(cf.config.Apis.UserStoreMgtAPI, cf.tokenStore)
}

// GetCustomDomainClient returns a custom domain client
func (cf *ClientFactory) GetCustomDomainClient() *customDomain.CustomDomainClient {
	return customDomain.NewCustomDomainClient(cf.config.Apis.CustomDomainAPI, cf.config.Apis.UrlMappingAPI, cf.tokenStore)
}
