package auth

import (
	"fmt"
	"os"
	"sync"

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

var (
	usrStore               *UserStore
	orgStore               *OrgStore
	tokenStore             *OrgTokenStore
	authClient             *api.AuthClient
	config                 *cfg.EnvConfig
	usrMgtClient           *api.UserManagementClient
	tokenRefreshMutex      sync.Mutex
	rOTokenStore           *ReadOnlyTokenStore
	currentTokenStore      api.ReadOnlyTokenStore
	defaultClientFactory   *ClientFactory
	OrgClient              *api.OrgClient
	ProjectClient          *project.ProjectClient
	GitClient              *platformgit.GitClient
	ComponentClient        *component.ComponentClient
	LogsClient             *logs.LogsClient
	ApimPublisherClient    *apimpublisher.ApimPublisherClient
	DevopsClient           *devops.DevopsClient
	DeploymentBuildClient  *deploymentbuild.DeploymentBuildClient
	ConnectionsClient      *connections.ConnectionsClient
	MarketplaceClient      *marketplace.MarketplaceClient
	GQLBuildClient         *gqlbuild.GQLBuildClient
	CredentialClient       *credentials.CredentialClient
	WorkflowMgtClient      *workflowmgt.WorkflowMgtClient
	SubscriptionsClient    *subscription.SubscriptionsClient
	ConfigMappingApiClient *configmapping.ConfigMappingSvcClient
	GraphQLClient          *component.GraphQLClient
	PlatformServicesClient *platformservices.PlatformServicesClient
	UserStoreMgtClient     *userstore_mgt.UserStoreManagementClient
	CustomDomainClient     *customDomain.CustomDomainClient
)

func init() {
	usrStore = NewUserStore()
	orgStore = NewOrgStore()
	tokenStore = NewOrgTokenStore(*orgStore)

	initAuth("")
}

func ReInit() {
	initAuth(authClient.Verifier)
}

func initAuth(verifier string) {
	config = region.GetRegionConfig() // TODO: Select the correct config from a config file/environment variable
	authClient = &api.AuthClient{
		Config: &api.AuthClientConfig{
			AsgardeoClientId: config.DevantConfig.AsgardeoClientId,
			STSClientId:      config.STSClientID,
			AsgardeoTokenUrl: config.TokenUrl,
			STSTokenUrl:      config.STSTokenUrl,
			STSScopes:        config.STSScopes,
			LoginUrl:         config.DevantConfig.ConsoleUrls.LoginUrl,
			SignUpUrl:        config.DevantConfig.ConsoleUrls.SignUpUrl,
			RedirectUrl:      config.DevantConfig.ConsoleUrls.RedirectUrl,
		},
		Verifier: verifier,
	}
	usrMgtClient = api.NewUserManagementClient(config.Apis.UserMgtAPI)
	rOTokenStore = &ReadOnlyTokenStore{}
	currentTokenStore = rOTokenStore

	initializeClients()
}

// initializeClients initializes all API clients using the default client factory
func initializeClients() {
	// Update the default client factory with current token store
	defaultClientFactory = NewClientFactory(currentTokenStore)

	// Initialize global clients using the client factory
	OrgClient = defaultClientFactory.GetOrgClient()
	ProjectClient = defaultClientFactory.GetProjectClient()
	ComponentClient = defaultClientFactory.GetComponentClient()
	GraphQLClient = defaultClientFactory.GetGraphQLClient()
	GitClient = defaultClientFactory.GetGitClient()
	LogsClient = defaultClientFactory.GetLogsClient()
	ApimPublisherClient = defaultClientFactory.GetApimPublisherClient()
	DevopsClient = defaultClientFactory.GetDevopsClient()
	DeploymentBuildClient = defaultClientFactory.GetDeploymentBuildClient()
	ConnectionsClient = defaultClientFactory.GetConnectionsClient()
	MarketplaceClient = defaultClientFactory.GetMarketplaceClient()
	GQLBuildClient = defaultClientFactory.GetGQLBuildClient()
	CredentialClient = defaultClientFactory.GetCredentialClient()
	WorkflowMgtClient = defaultClientFactory.GetWorkflowMgtClient()
	SubscriptionsClient = defaultClientFactory.GetSubscriptionsClient()
	ConfigMappingApiClient = defaultClientFactory.GetConfigMappingApiClient()
	PlatformServicesClient = defaultClientFactory.GetPlatformServicesClient()
	UserStoreMgtClient = defaultClientFactory.GetUserStoreMgtClient()
	CustomDomainClient = defaultClientFactory.GetCustomDomainClient()
}

func IsLoggedIn() bool {
	org, err := rOTokenStore.GetActiveOrg()
	if err != nil || org == nil {
		if TryAuthenticateWithEnvPAT() {
			return true
		}
		// Bootstrap a full session from the STS token (stores to token store so
		// subsequent calls use the normal path including expiry/refresh).
		if stsToken := os.Getenv("CLOUD_STS_TOKEN"); stsToken != "" {
			return LoginWithSTSToken(stsToken) == nil
		}
		return false
	}

	token, err := rOTokenStore.getToken(org.ID, false, []api.Organization{*org})
	if err != nil || token == "" {
		return TryAuthenticateWithEnvPAT()
	}
	return true
}

func GetStsToken() (string, error) {
	org, err := rOTokenStore.GetActiveOrg()
	if err != nil {
		return "", err
	}
	if org == nil {
		return "", fmt.Errorf("active org not found")
	}

	token, err := rOTokenStore.getToken(org.ID, false, []api.Organization{*org})
	if err != nil {
		return "", err
	}
	if token == "" {
		return "", fmt.Errorf("sts token not found")
	}
	return token, nil
}

func GetCurrentUser() (*api.UserInfo, error) {
	userInfo, err := usrStore.RetrieveUserInfo()
	if err != nil || userInfo == nil {
		// If no stored user info, try PAT authentication from environment variable
		if TryAuthenticateWithEnvPAT() {
			return usrStore.RetrieveUserInfo()
		}
		return nil, err
	}
	return userInfo, nil
}

// TryAuthenticateWithEnvPAT attempts to authenticate using WSO2IP_PAT environment variable
// Returns true if authentication was successful, false otherwise
func TryAuthenticateWithEnvPAT() bool {
	patToken := os.Getenv("WSO2IP_PAT")
	if patToken == "" {
		return false
	}

	// Try to authenticate with the PAT from environment variable
	userInfo, err := LoginWithToken(patToken)
	if err != nil || userInfo == nil {
		return false
	}

	return true
}

func ClearAuthStores() {
	orgs, _ := orgStore.GetOrgs()
	orgStore.RemoveDefaultOrg()
	usrStore.ClearUserInfo()

	if orgs != nil {
		for _, org := range orgs {
			tokenStore.DeleteToken(region.GetCurrentRegion(), org.ID)
		}
	}
}

func HasValidToken(region string, orgId string) bool {
	return tokenStore.HasValidToken(region, orgId)
}
