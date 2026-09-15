package region

import (
	"github.com/wso2/integration-platform-tools/pkg/api/platformgit"
)

// EU Region Configurations
var EU_ENV_CONFIG_PROD = BuildEnvConfig(
	"https://console.eu.choreo.dev",
	"https://apis.eu.choreo.dev",
	"https://sts.eu.choreo.dev",
	"https://choreocontrolplane.eu.choreo.dev",
	"https://app.eu.choreo.dev",
	"https://apis.eu.choreo.dev",
	"https://apis.eu.choreo.dev",
	"/projects/1.0.0/graphql",
	RegionOverrides{
		AsgardeoClientId: "dIJufnUizYr5UcN6z8rI2LRVVlEa",
		STSClientID:      "choreoclif3b070a3d4866bb5",
		STSTokenUrl:      "https://sts.eu.choreo.dev/oauth2/token",
		AsgardeoScope:    "openid+email+profile",
		STSScopes:        "urn:choreosystem:componentutils:component_manage urn:choreosystem:componentutils:component_file_view urn:choreosystem:configmanagement:config_create urn:choreosystem:configmanagement:config_delete urn:choreosystem:configmanagement:config_manage urn:choreosystem:configmanagement:config_view urn:choreosystem:componentsmanagement:component_create urn:choreosystem:choreodevopsportalapi:component_manage urn:choreosystem:choreodevopsportalapi:deployment_manage urn:choreosystem:choreodevopsportalapi:deployment_view urn:choreosystem:componentsmanagement:component_logs_view urn:choreosystem:componentsmanagement:component_init_view choreo:project_view urn:choreosystem:organizationapi:org_manage urn:choreosystem:componentsmanagement:component_file_view urn:choreosystem:componentsmanagement:component_manage choreo:domain_manage choreo:domain_view choreo:url_mapping_manage choreo:url_mapping_approve choreo:url_mapping_view",
		GhApp: platformgit.GHAppConfig{
			AppUrl:      "https://github.com/apps/wso2-cloud-app",
			InstallUrl:  "https://github.com/apps/wso2-cloud-app/installations/new",
			AuthUrl:     "https://github.com/login/oauth/authorize",
			ClientId:    "Iv1.804167a242012c66",
			RedirectUrl: "https://console.eu.choreo.dev/ghapp",
		},
		BillingUrl:           "https://subscriptions.eu.wso2.com",
		SysApiPrefixUrl:      "https://525aafa3-b9c1-4d4e-b8fb-18fd0e39b137-systemapis",
		DevopsApiBaseUrl:     "https://apis.eu.choreo.dev/devops/1.0.0",
		ApimPublisherBaseUrl: "https://sts.eu.choreo.dev/api/am/publisher/v2",
		TokenUrl:             "https://api.eu.asgardeo.io/t/a/oauth2/token",
		PlatformHostname:     "customdns.e1-eu-north-azure",
		DevantConfig: DevantConfig{
			ConsoleUrl:       "https://console.eu.devant.dev",
			AsgardeoClientId: "IIQPy5qNaETGE_aGXr8nkV5GszYa",
		},
	},
)

var EU_ENV_CONFIG_STAGE = BuildEnvConfig(
	"https://REDACTED.invalid",
	"https://REDACTED.invalid",
	"https://REDACTED.invalid",
	"https://REDACTED.invalid",
	"https://REDACTED.invalid",
	"https://REDACTED.invalid",
	"https://REDACTED.invalid",
	"/projects/1.0.0/graphql",
	RegionOverrides{
		AsgardeoClientId: "EkkqjY_rv29u2OVmliKkHe9fUuMa",
		STSClientID:      "choreoclif3b070a3d4866bb5",
		STSTokenUrl:      "https://REDACTED.invalid:443/oauth2/token",
		AsgardeoScope:    "openid+email+profile",
		STSScopes:        "urn:choreosystem:componentutils:component_manage urn:choreosystem:componentutils:component_file_view urn:choreosystem:configmanagement:config_create urn:choreosystem:configmanagement:config_delete urn:choreosystem:configmanagement:config_manage urn:choreosystem:configmanagement:config_view urn:choreosystem:componentsmanagement:component_create urn:choreosystem:choreodevopsportalapi:component_manage urn:choreosystem:choreodevopsportalapi:deployment_manage urn:choreosystem:choreodevopsportalapi:deployment_view urn:choreosystem:componentsmanagement:component_logs_view urn:choreosystem:componentsmanagement:component_init_view choreo:project_view urn:choreosystem:organizationapi:org_manage urn:choreosystem:componentsmanagement:component_file_view urn:choreosystem:componentsmanagement:component_manage choreo:domain_manage choreo:domain_view choreo:url_mapping_manage choreo:url_mapping_approve choreo:url_mapping_view",
		GhApp: platformgit.GHAppConfig{
			AppUrl:      "https://github.com/apps/wso2-cloud-app-stage",
			InstallUrl:  "https://github.com/apps/wso2-cloud-app-stage/installations/new",
			AuthUrl:     "https://github.com/login/oauth/authorize",
			ClientId:    "Iv1.20fd2645fc8a5aab",
			RedirectUrl: "https://REDACTED.invalid/ghapp",
		},
		BillingUrl:           "https://REDACTED.invalid",
		SysApiPrefixUrl:      "https://REDACTED.invalid",
		DevopsApiBaseUrl:     "https://REDACTED.invalid/devops/1.0.0",
		ApimPublisherBaseUrl: "https://REDACTED.invalid/api/am/publisher/v2",
		TokenUrl:             "https://REDACTED.invalid/t/a/oauth2/token",
		PlatformHostname:     "customdns.e1-eu-north-azure.st",
		DevantConfig: DevantConfig{
			ConsoleUrl:       "https://REDACTED.invalid",
			AsgardeoClientId: "fTKXHKMQ9zfXEWAkAG8I8y6IzXsa",
		},
	},
)

var EU_ENV_CONFIG_DEV = BuildEnvConfig(
	"https://REDACTED.invalid",
	"https://REDACTED.invalid",
	"https://REDACTED.invalid",
	"https://REDACTED.invalid",
	"https://REDACTED.invalid",
	"https://REDACTED.invalid",
	"https://REDACTED.invalid",
	"/projects/1.0.0/graphql",
	RegionOverrides{
		AsgardeoClientId: "TUUGRYYR7OWf2pfq8eGuluJgEEUa",
		STSClientID:      "Fn4zd7hN6wCkpqInL5klSBuw0cMa",
		STSTokenUrl:      "https://REDACTED.invalid:443/oauth2/token",
		AsgardeoScope:    "openid+email+profile",
		STSScopes:        "urn:choreosystem:componentutils:component_manage urn:choreosystem:componentutils:component_file_view urn:choreosystem:configmanagement:config_create urn:choreosystem:configmanagement:config_delete urn:choreosystem:configmanagement:config_manage urn:choreosystem:configmanagement:config_view urn:choreosystem:componentsmanagement:component_create urn:choreosystem:choreodevopsportalapi:component_manage urn:choreosystem:choreodevopsportalapi:deployment_manage urn:choreosystem:choreodevopsportalapi:deployment_view urn:choreosystem:componentsmanagement:component_logs_view urn:choreosystem:componentsmanagement:component_init_view choreo:project_view urn:choreosystem:organizationapi:org_manage urn:choreosystem:componentsmanagement:component_file_view urn:choreosystem:componentsmanagement:component_manage choreo:domain_manage choreo:domain_view choreo:url_mapping_manage choreo:url_mapping_approve choreo:url_mapping_view",
		GhApp: platformgit.GHAppConfig{
			AppUrl:      "https://github.com/apps/choreo-apps-dev-eu-aws",
			InstallUrl:  "https://github.com/apps/choreo-apps-dev-eu-aws/installations/new",
			AuthUrl:     "https://github.com/login/oauth/authorize",
			ClientId:    "Iv23lioFlLXo2VNznIh5",
			RedirectUrl: "https://REDACTED.invalid/ghapp",
		},
		BillingUrl:           "https://REDACTED.invalid",
		SysApiPrefixUrl:      "https://REDACTED.invalid",
		DevopsApiBaseUrl:     "https://REDACTED.invalid/devops/1.0.0",
		ApimPublisherBaseUrl: "https://REDACTED.invalid/api/am/publisher/v2",
		TokenUrl:             "https://REDACTED.invalid/t/a/oauth2/token",
		PlatformHostname:     "customdns.e1-eu-north-azure.preview-dv",
		DevantConfig: DevantConfig{
			ConsoleUrl:       "https://REDACTED.invalid",
			AsgardeoClientId: "XaIBIF2DiPZyDXO38ZcqVZPkx_ca",
		},
	},
)

func init() {
	InitEURegion(&EU_ENV_CONFIG_DEV, &EU_ENV_CONFIG_STAGE, &EU_ENV_CONFIG_PROD)
}
