package config

import (
	"github.com/wso2/integration-platform-tools/pkg/api/platformgit"
)

type PlatformHostnames struct {
	WebApp string
	API    string
}

type Apis struct {
	GraphQLAPI          string
	OrgsAPI             string
	ProjectAPI          string
	ComponentManageAPI  string
	UserMgtAPI          string
	ComponentAPI        string
	ConfigAPI           string
	ConfigMgtAPI        string
	ConfigMappingAPI    string
	DeclarativeAPI      string
	ConnectionsAPI      string
	MarketplaceAPI      string
	ComponentUtilAPI    string
	ProxyDeployEp       string
	InsightsQueryApi    string
	APIKeyService       string
	WorkflowMgtAPI      string
	PlatformServicesAPI string
	UserStoreMgtAPI     string
	CustomDomainAPI     string
	UrlMappingAPI       string
}

type EnvConfig struct {
	ConsoleUrls           ConsoleUrls
	TokenUrl              string
	AsgardeoClientId      string
	STSClientID           string
	STSTokenUrl           string
	AsgardeoScope         string
	GhApp                 platformgit.GHAppConfig
	Apis                  Apis
	STSScopes             string
	BillingConsoleBaseUrl string
	SysApiPrefixUrl       string
	DevopsApiBaseUrl      string
	ApimPublisherBaseUrl  string
	DevantConfig          DevantConfig
	PlatformHostnames     PlatformHostnames
}

type DevantConfig struct {
	ConsoleUrls      ConsoleUrls
	AsgardeoClientId string
}

type ConsoleUrls struct {
	BaseUrl     string
	LoginUrl    string
	RedirectUrl string
	SignUpUrl   string
}

const (
	ENV_DEV   = "dev"
	ENV_PROD  = "prod"
	ENV_STAGE = "stage"
)
