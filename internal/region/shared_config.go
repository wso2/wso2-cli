package region

import (
	"github.com/wso2/integration-platform-tools/internal/config"
	"github.com/wso2/integration-platform-tools/pkg/api/platformgit"
)

type RegionOverrides struct {
	AsgardeoClientId     string
	STSClientID          string
	STSTokenUrl          string
	AsgardeoScope        string
	STSScopes            string
	GhApp                platformgit.GHAppConfig
	BillingUrl           string
	SysApiPrefixUrl      string
	DevopsApiBaseUrl     string
	ApimPublisherBaseUrl string
	TokenUrl             string
	PlatformHostname     string
	DevantConfig         DevantConfig
}

// Devant config needed by vscode platform extension
type DevantConfig struct {
	ConsoleUrl       string
	AsgardeoClientId string
}

func BuildEnvConfig(
	consoleHost string,
	apiHost string,
	stsHost string,
	insightsHost string,
	appHost string,
	marketplaceHost string,
	connectionsHost string,
	regionApisPath string,
	overrides RegionOverrides,
) config.EnvConfig {
	return config.EnvConfig{
		TokenUrl: overrides.TokenUrl,
		ConsoleUrls: config.ConsoleUrls{
			BaseUrl:     consoleHost,
			LoginUrl:    consoleHost + "/login",
			RedirectUrl: consoleHost + "/vscode-auth",
			SignUpUrl:   consoleHost + "/signup",
		},
		AsgardeoClientId: overrides.AsgardeoClientId,
		STSClientID:      overrides.STSClientID,
		STSTokenUrl:      overrides.STSTokenUrl,
		AsgardeoScope:    overrides.AsgardeoScope,
		STSScopes:        overrides.STSScopes,
		GhApp:            overrides.GhApp,
		Apis: config.Apis{
			OrgsAPI:             apiHost + "/orgs/1.0.0/orgs",
			ProjectAPI:          apiHost + "/projects/1.0.0/graphql",
			ComponentManageAPI:  apiHost + "/component-mgt/1.0.0",
			UserMgtAPI:          apiHost + "/user-mgt/1.0.0",
			ComponentAPI:        apiHost + "/projects/1.0.0/graphql",
			ConfigAPI:           apiHost + "/config-svc/v1.0",
			ConfigMgtAPI:        apiHost + "/config-mgt/1.0.0",
			ConfigMappingAPI:    apiHost + "/config-mapping-svc/v1.0",
			DeclarativeAPI:      apiHost + "/declarative-api/v1.0/core/v1alpha1",
			ConnectionsAPI:      connectionsHost + "/connections/v1",
			MarketplaceAPI:      marketplaceHost + "/marketplace/0.1.0",
			ComponentUtilAPI:    apiHost + "/component-utils/1.0.0",
			GraphQLAPI:          apiHost + "/projects/1.0.0/graphql",
			ProxyDeployEp:       appHost + "/proxy/deployer/v1",
			InsightsQueryApi:    insightsHost + "/insights/1.0.0/query-api",
			APIKeyService:       apiHost + "/api-key-service/v1.0",
			WorkflowMgtAPI:      apiHost + "/workflow-mgt/v1.0",
			PlatformServicesAPI: apiHost + "/platform-services/v1.0",
			UserStoreMgtAPI:     apiHost + "/user-store-mgt/v1.0",
			CustomDomainAPI:     apiHost + "/url-mgt/v1.0/domains",
			UrlMappingAPI:       apiHost + "/url-mgt/v1.0/url-mappings",
		},
		BillingConsoleBaseUrl: overrides.BillingUrl,
		SysApiPrefixUrl:       overrides.SysApiPrefixUrl,
		DevopsApiBaseUrl:      overrides.DevopsApiBaseUrl,
		ApimPublisherBaseUrl:  overrides.ApimPublisherBaseUrl,
		DevantConfig: config.DevantConfig{
			AsgardeoClientId: overrides.DevantConfig.AsgardeoClientId,
			ConsoleUrls: config.ConsoleUrls{
				BaseUrl:     overrides.DevantConfig.ConsoleUrl,
				LoginUrl:    overrides.DevantConfig.ConsoleUrl + "/login",
				RedirectUrl: overrides.DevantConfig.ConsoleUrl + "/vscode-auth",
				SignUpUrl:   overrides.DevantConfig.ConsoleUrl + "/signup",
			},
		},
		PlatformHostnames: config.PlatformHostnames{
			WebApp: overrides.PlatformHostname + ".choreoapps.dev",
			API:    overrides.PlatformHostname + ".choreoapis.dev",
		},
	}
}
