package component

import "slices"

// These will be used when dealing with the platform API

type ComponentCreateResponse struct {
	Id             string `yaml:"id"`
	OrganizationId string `yaml:"organizationId"`
	ProjectId      string `yaml:"projectId"`
	Handle         string `yaml:"handle"`
}

// This will be used in when dealing with the local component yaml
type ComponentInformationMetadata struct {
	Org                   ComponentInformationMetadataOrg              `yaml:"org"`
	DisplayName           string                                       `yaml:"displayName"`
	DisplayType           string                                       `yaml:"displayType"`
	Description           string                                       `yaml:"description"`
	Version               string                                       `yaml:"version"`
	ProjectID             string                                       `yaml:"projectId"`
	Accessibility         string                                       `yaml:"accessibility"`
	Repository            ComponentInformationMetadataRepo             `yaml:"repository"`
	ByocConfig            *ComponentInformationMetadataByocConfig      `yaml:"byocConfig,omitempty"`
	ByocWebAppBuildConfig *ComponentInformationMetadataWebAppConfig    `yaml:"byocWebAppBuildConfig,omitempty"`
	BuildpackConfig       *ComponentInformationMetadataBuildPackConfig `yaml:"buildpackConfig,omitempty"`
}

type ComponentInformationMetadataOrg struct {
	ID      string `yaml:"id"`
	Handle  string `yaml:"handle"`
	OrgUUID string `yaml:"orgUuid"`
}

type ComponentInformationMetadataRepo struct {
	AppSubPath            string `yaml:"appSubPath"`
	OrgApp                string `yaml:"orgApp"`
	NameApp               string `yaml:"nameApp"`
	BranchApp             string `yaml:"branchApp"`
	GitProvider           string `yaml:"gitProvider"`
	BitbucketCredentialID string `yaml:"bitbucketCredentialId,omitempty"`
}

type ComponentInformationMetadataWebAppConfig struct {
	WebAppType            string `yaml:"webAppType"`
	BuildCommand          string `yaml:"buildCommand"`
	PackageManagerVersion string `yaml:"packageManagerVersion"`
	OutputDirectory       string `yaml:"outputDirectory"`
}

type ComponentInformationMetadataByocConfig struct {
	DockerContext    string `yaml:"dockerContext"`
	DockerfilePath   string `yaml:"dockerfilePath"`
	SrcGitRepoBranch string `yaml:"srcGitRepoBranch"`
	SrcGitRepoUrl    string `yaml:"srcGitRepoUrl"`
}

type ComponentInformationMetadataBuildPackConfig struct {
	BuildContext     string `yaml:"buildContext"`
	SrcGitRepoUrl    string `yaml:"srcGitRepoUrl"`
	SrcGitRepoBranch string `yaml:"srcGitRepoBranch"`
	LanguageVersion  string `yaml:"languageVersion"`
	BuildpackId      string `yaml:"buildpackId"`
	BuildpackName    string `yaml:"buildpackName"`
	Port             uint   `yaml:"port,omitempty"`
}

type ComponentInformation struct {
	// other properties such as endpoints, configs & connector details will also come here
	Metadata   ComponentInformationMetadata `yaml:"metadata"`
	Id         string                       `yaml:"id"`
	CommitHash string                       `yaml:"commitHash"`
}

const (
	DisplayTypeService               = "ballerinaService"
	DisplayTypeByocService           = "byocService"
	DisplayTypeByocWebApp            = "byocWebApp"
	DisplayTypeRestApi               = "restAPI"
	DisplayTypeManualTrigger         = "manualTrigger"
	DisplayTypeScheduledTask         = "scheduledTask"
	DisplayTypeWebhook               = "webhook"
	DisplayTypeWebsocket             = "webSocket"
	DisplayTypeProxy                 = "proxy"
	DisplayTypeByocCronjob           = "byocCronjob"
	DisplayTypeByocJob               = "byocJob"
	DisplayTypeGraphQL               = "graphql"
	DisplayTypeThirdPartyAPI         = "thirdPartyAPI"
	DisplayTypeByocWebAppDockerLess  = "byocWebAppsDockerfileLess"
	DisplayTypeByocRestApi           = "byocRestApi"
	DisplayTypeByocWebhook           = "byocWebhook"
	DisplayTypeByocEventHandler      = "byocEventHandler"
	DisplayTypeMiRestApi             = "miRestApi"
	DisplayTypeMiEventHandler        = "miEventHandler"
	DisplayTypeMiApiService          = "miApiService"
	DisplayTypePrismMockService      = "prismMockService"
	DisplayTypeMiCronjob             = "miCronjob"
	DisplayTypeMiJob                 = "miJob"
	DisplayTypeByoiService           = "byoiService"
	DisplayTypeByoiJob               = "byoiJob"
	DisplayTypeByoiCronjob           = "byoiCronjob"
	DisplayTypeByoiWebApp            = "byoiWebApp"
	DisplayTypeByocTestRunner        = "byocTestRunner"
	DisplayTypeBuildpackService      = "buildpackService"
	DisplayTypeBuildpackWebhook      = "buildpackWebhook"
	DisplayTypeBuildpackJob          = "buildpackJob"
	DisplayTypeBuildpackTestRunner   = "buildpackTestRunner"
	DisplayTypeBuildpackCronJob      = "buildpackCronjob"
	DisplayTypeBuildpackWebApp       = "buildpackWebApp"
	DisplayTypeBuildpackRestApi      = "buildpackRestApi"
	DisplayTypeBuildpackEventHandler = "buildpackEventHandler"
	DisplayTypePostmanTestRunner     = "byocTestRunnerDockerfileLess"
	DisplayTypeBallerinaEventHandler = "ballerinaEventHandler"
	DisplayTypeGitProxy              = "gitProxy"
)

const (
	ComponentBuildPackDocker      = "docker"
	ComponentBuildPackPrism       = "prism"
	ComponentBuildPackBallerina   = "ballerina"
	ComponentBuildPackMI          = "microintegrator"
	ComponentBuildPackReact       = "react"
	ComponentBuildPackVue         = "vuejs"
	ComponentBuildPackAngular     = "angular"
	ComponentBuildPackStaticFiles = "staticweb"
)

var ComponentBuildPackSPAs = []string{ComponentBuildPackReact, ComponentBuildPackVue, ComponentBuildPackAngular}

const (
	ComponentTypeManualTrigger = "manualTask"
	ComponentTypeScheduledTask = "scheduleTask"
	ComponentTypeByocWebApp    = "webApp"
	ComponentTypeService       = "service"
	ComponentTypeEventHandler  = "eventHandler"
	ComponentTypeWebhook       = "webhook"
	ComponentTypeProxyGH       = "proxy"
)

var ComponentTypes = []string{
	ComponentTypeService,
	ComponentTypeScheduledTask,
	ComponentTypeEventHandler,
}

const (
	ComponentServiceTypeRest    = "REST"
	ComponentServiceTypeGraphQL = "GraphQL"
	ComponentServiceTypeGRPC    = "GRPC"
	ComponentServiceTypeTCP     = "tcp"
	ComponentServiceTypeUDP     = "udp"
)

var ComponentServiceTypes = []string{
	ComponentServiceTypeRest,
	ComponentServiceTypeGraphQL,
	ComponentServiceTypeGRPC,
}

const (
	ComponentServiceVisibilityProject      = "Project"
	ComponentServiceVisibilityOrganization = "Organization"
	ComponentServiceVisibilityPublic       = "Public"
)

var ComponentServiceVisibilityOptions = []string{
	ComponentServiceVisibilityProject,
	ComponentServiceVisibilityOrganization,
	ComponentServiceVisibilityPublic,
}

func GetDisplayNameForComponentType(componentType string, buildPackType string) (returnStr string) {
	switch componentType {
	case ComponentTypeService:
		if buildPackType == ComponentBuildPackMI {
			return DisplayTypeMiApiService
		}
		return DisplayTypeService
	case ComponentTypeScheduledTask:
		if buildPackType == ComponentBuildPackMI {
			return DisplayTypeMiCronjob
		}
		return DisplayTypeScheduledTask
	case ComponentTypeEventHandler:
		if buildPackType == ComponentBuildPackMI {
			return DisplayTypeMiEventHandler
		}
		return DisplayTypeBallerinaEventHandler
	}
	return
}

type ComponentEndpointsData struct {
	Id                string `json:"id"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
	ReleaseId         string `json:"releaseId"`
	Port              int    `json:"port"`
	EnvironmentId     string `json:"environmentId"`
	DisplayName       string `json:"displayName"`
	InvokeUrl         string `json:"invokeUrl"`
	HostName          string `json:"hostName"`
	IsAutoGenerated   bool   `json:"isAutoGenerated"`
	Protocol          string `json:"protocol"`
	ApiContext        string `json:"apiContext"`
	ApiDefinitionPath string `json:"apiDefinitionPath"`
	Visibility        string `json:"visibility"`
	Type              string `json:"type"`
	ApimId            string `json:"apimId"`
	ApimRevisionId    string `json:"apimRevisionId"`
	ApimName          string `json:"apimName"`
	ProjectUrl        string `json:"projectUrl"`
	OrganizationUrl   string `json:"organizationUrl"`
	PublicUrl         string `json:"publicUrl"`
	State             string `json:"state"`
	StateReason       struct {
		Code     string `json:"code"`
		Message  string `json:"message"`
		Details  string `json:"details"`
		WorkerId string `json:"workerId"`
	} `json:"stateReason"`
	IsDeleted bool   `json:"isDeleted"`
	DeletedAt string `json:"deletedAt"`
}

type Author struct {
	Name      string `json:"name"`
	Date      string `json:"date"`
	Email     string `json:"email"`
	AvatarUrl string `json:"avatarUrl"`
}

type Commit struct {
	Author   Author `json:"author"`
	Sha      string `json:"sha"`
	Message  string `json:"message"`
	IsLatest bool   `json:"isLatest"`
}

type Build struct {
	BuildId    string `json:"buildId"`
	DeployedAt string `json:"deployedAt"`
	Commit     Commit `json:"commit"`
	RunId      string `json:"runId"`
}

type ComponentDeployment struct {
	EnvironmentId      string `json:"environmentId"`
	ConfigCount        int    `json:"configCount"`
	ApiId              string `json:"apiId"`
	ReleaseId          string `json:"releaseId"`
	ApiRevision        string `json:"apiRevision"`
	Build              Build  `json:"build"`
	ImageUrl           string `json:"imageUrl"`
	InvokeUrl          string `json:"invokeUrl"`
	VersionId          string `json:"versionId"`
	DeploymentStatus   string `json:"deploymentStatus"`
	DeploymentStatusV2 string `json:"deploymentStatusV2"`
	Version            string `json:"version"`
	Cron               string `json:"cron"`
	CronTimezone       string `json:"cronTimezone"`
}

type AutoBuildToggleResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type AutoBuildStatusResponse struct {
	Success bool `json:"success"`
	Data    struct {
		AutoBuildId      string `json:"autoBuildId"`
		AutoBuildEnabled bool   `json:"autoBuildEnabled"`
		ComponentId      string `json:"componentId"`
		VersionId        string `json:"versionId"`
		EnvId            string `json:"envId"`
	} `json:"data"`
	Message string `json:"message"`
}

type ComponentInitStatusResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
	} `json:"data"`
	Message string `json:"message"`
}

type ComponentDeploymentResponse struct {
	DeployComponent struct {
		Message string `json:"message"`
		Success bool   `json:"success"`
	} `json:"deployComponent"`
}

type ComponentDeploymentStatusForVersion struct {
	Id             int    `json:"id"`
	Sha            string `json:"sha"`
	CompletedAt    string `json:"completed_at"`
	StartedAt      string `json:"started_at"`
	Name           string `json:"name"`
	Status         string `json:"status"`
	Conclusion     string `json:"conclusion"`
	IsAutoDeploy   bool   `json:"isAutoDeploy"`
	FailureReason  int    `json:"failureReason"`
	SourceCommitId string `json:"sourceCommitId"`
	ClusterId      string `json:"clusterId"`
	BuildRef       string `json:"buildRef"`
}

type ComponentDeploymentTrackImages struct {
	ImageId       string `json:"imageId"`
	CommitHash    string `json:"commitHash"`
	CommitMessage string `json:"commitMessage"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type DeploymentBuildStatusResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Init struct {
			BuildPhase
			Log string `json:"log"`
		} `json:"init"`
		Build struct {
			BuildPhase
			Log string `json:"log"`
		} `json:"build"`
		OasValidation struct {
			BuildPhase
			Log string `json:"log"`
		} `json:"oasValidation"`
		GovernanceCheck struct {
			BuildPhase
			Log string `json:"log"`
		} `json:"governanceCheck"`
		UpdateApi struct {
			BuildPhase
			Log string `json:"log"`
		} `json:"updateApi"`
		Deploy BuildPhase `json:"deploy"`
	} `json:"data"`
}

type BuildStep struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Conclusion  string `json:"conclusion"`
	Number      int    `json:"number"`
	StartedAt   string `json:"started_at"`
	CompletedAt string `json:"completed_at"`
}

type BuildPhase struct {
	Status string      `json:"status"`
	Steps  []BuildStep `json:"steps"`
}

type ComponentSpecImage struct {
	Registry   string `yaml:"registry"`
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
}

type ComponentSpecBuild struct {
	Branch   string `yaml:"branch"`
	Revision string `yaml:"revision,omitempty"` // optional
}

type ComponentConfigFromTo struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

type ComponentConfigGroupVolume struct {
	MountPath string                  `yaml:"mountPath,omitempty"` // Optional
	Files     []ComponentConfigFromTo `yaml:"files"`
}

type ComponentConfigOutboundServiceRefs struct {
	Name            string                  `yaml:"name"`
	DependentConfig string                  `yaml:"dependentConfig"`
	Env             []ComponentConfigFromTo `yaml:"env"`
}

type ComponentSpecConfigKey struct {
	Name    string `yaml:"name"`
	EnvName string `yaml:"envName"`
	Volume  struct {
		MountPath string `yaml:"mountPath"`
	} `yaml:"volume,omitempty"` // Optional
}

type ComponentSpecConfigGroup struct {
	Name   string                     `yaml:"name"`
	Env    []ComponentConfigFromTo    `yaml:"env,omitempty"`    // Optional
	Volume ComponentConfigGroupVolume `yaml:"volume,omitempty"` // Optional
}

func GetTypeForDisplayType(displayType string) string {
	switch displayType {
	case DisplayTypeRestApi, DisplayTypeService, DisplayTypeByocService, DisplayTypeGraphQL, DisplayTypeMiApiService, DisplayTypeMiRestApi, DisplayTypeBuildpackService, DisplayTypeBuildpackRestApi, DisplayTypeWebsocket, DisplayTypePrismMockService, DisplayTypeByoiService, DisplayTypeByocRestApi:
		return "service"
	case DisplayTypeManualTrigger, DisplayTypeByocJob, DisplayTypeBuildpackJob, DisplayTypeMiJob, DisplayTypeByoiJob:
		return "manual-task"
	case DisplayTypeScheduledTask, DisplayTypeByocCronjob, DisplayTypeBuildpackCronJob, DisplayTypeMiCronjob, DisplayTypeByoiCronjob:
		return "scheduled-task"
	case DisplayTypeWebhook, DisplayTypeByocWebhook, DisplayTypeBuildpackWebhook:
		return "web-hook"
	case DisplayTypeProxy, DisplayTypeGitProxy:
		return "proxy"
	case DisplayTypeByocWebApp, DisplayTypeByocWebAppDockerLess, DisplayTypeBuildpackWebApp, DisplayTypeByoiWebApp:
		return "web-app"
	case DisplayTypeMiEventHandler, DisplayTypeBallerinaEventHandler:
		return "event-handler"
	case DisplayTypeBuildpackTestRunner:
		return "test"
	}
	return ""
}

// IntegrationComponentTypes are the high-level component types (as returned
// by GetTypeForDisplayType) exposed as "integrations" by the MCP server.
// Webapps, webhooks, manual tasks, proxies, and test runners are not
// integrations and are deliberately excluded.
var IntegrationComponentTypes = []string{"service", "scheduled-task", "event-handler"}

// IsIntegrationDisplayType reports whether a component's displayType maps to
// one of IntegrationComponentTypes.
func IsIntegrationDisplayType(displayType string) bool {
	return slices.Contains(IntegrationComponentTypes, GetTypeForDisplayType(displayType))
}

type ComponentKindSource struct {
	Bitbucket *GitProvider `json:"bitbucket,omitempty"`
	Github    *GitProvider `json:"github,omitempty"`
	Gitlab    *GitProvider `json:"gitlab,omitempty"`
	SecretRef string       `json:"secretRef,omitempty"`
}

type GitProvider struct {
	Repository           string `json:"repository"`
	Branch               string `json:"branch"`
	Path                 string `json:"path"`
	IsPublicRepo         bool   `json:"isPublicRepo"`
	PullLatestSubmodules bool   `json:"pullLatestSubmodules"`
}

type ComponentKindBuildDocker struct {
	DockerFilePath    string `json:"dockerFilePath"`
	DockerContextPath string `json:"dockerContextPath"`
	Port              int    `json:"port,omitempty"`
}

type ComponentKindBuildBallerina struct {
	SampleTemplate    string `json:"sampleTemplate,omitempty"`
	EnableCellDiagram bool   `json:"enableCellDiagram"`
	IsUnitTestEnabled bool   `json:"isUnitTestEnabled"`
}

type ComponentKindBuildWebapp struct {
	BuildCommand string `json:"buildCommand"`
	NodeVersion  string `json:"nodeVersion"`
	OutputDir    string `json:"outputDir"`
	Type         string `json:"type"`
}

type ComponentKindApiProxy struct {
	Version       string `json:"version"`
	Context       string `json:"context"`
	EndpointUrl   string `json:"endpointUrl"`
	Accessibility string `json:"accessibility"`
}

type ComponentKindBuildBuildpack struct {
	Language string `json:"language"`
	Version  string `json:"version"`
	Port     int    `json:"port,omitempty"`
}

type ComponentKindSpec struct {
	Type        string                 `json:"type"`
	SubType     string                 `json:"subType,omitempty"`
	Source      ComponentKindSource    `json:"source"`
	Build       ComponentKindSpecBuild `json:"build"`
	ProxyConfig *ComponentKindApiProxy `json:"proxyConfig,omitempty"`
}

type ComponentKindSpecBuild struct {
	Docker           *ComponentKindBuildDocker    `json:"docker,omitempty"`
	Ballerina        *ComponentKindBuildBallerina `json:"ballerina,omitempty"`
	Webapp           *ComponentKindBuildWebapp    `json:"webapp,omitempty"`
	Buildpack        *ComponentKindBuildBuildpack `json:"buildpack,omitempty"`
	EnableAutoBuild  bool                         `json:"enableAutoBuild"`
	EnableAutoDeploy bool                         `json:"enableAutoDeploy"`
}

type ComponentKindMetadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	ProjectName string `json:"projectName"`
	Description string `json:"description"`
	ID          string `json:"id,omitempty"`
	Handler     string `json:"handler,omitempty"`
}

type ComponentKind struct {
	ApiVersion    string                `json:"apiVersion"`
	Kind          string                `json:"kind"`
	Metadata      ComponentKindMetadata `json:"metadata"`
	Spec          ComponentKindSpec     `json:"spec"`
	OriginCloud   *string               `json:"originCloud,omitempty"`
	CorrelationID string                `json:"-"` // not serialized; populated from X-Correlation-Id response header
}

type ComponentDeleteResponse struct {
	Status    string `json:"status"`
	CanDelete bool   `json:"canDelete"`
	Message   string `json:"message"`
}

type ComponentTypeCount struct {
	ComponentType string `json:"componentType"`
	Count         int    `json:"count"`
}

type ComponentLimits struct {
	BillableComponentCount         int                  `json:"billableComponentCount"`
	ComponentCount                 int                  `json:"componentCount"`
	DistinctTypeCount              []ComponentTypeCount `json:"distinctTypeCount"`
	ExternalConsumerComponentCount int                  `json:"externalConsumerComponentCount"`
	IsWebappConstrained            bool                 `json:"isWebappConstrained"`
	OrgID                          int                  `json:"orgId"`
	SystemComponentCount           int                  `json:"systemComponentCount"`
}

type ComponentLimitsResponse struct {
	Data ComponentLimits `json:"data"`
}

type RepoDirStructResponse struct {
	Success bool        `json:"success"`
	Data    []PathEntry `json:"data"`
}

type PathEntryType string

const (
	PathTreeType PathEntryType = "tree"
	PathBlobType               = "blob"
)

type PathEntry struct {
	Path     string        `json:"path"`
	SubPath  string        `json:"subPath"`
	Type     PathEntryType `json:"type"`
	Children []PathEntry   `json:"children"`
}

type RepoMetadataResponse struct {
	IsBareRepo               bool `json:"isBareRepo"`
	IsSubPathEmpty           bool `json:"isSubPathEmpty"`
	IsSubPathValid           bool `json:"isSubPathValid"`
	IsValidRepo              bool `json:"isValidRepo"`
	HasBallerinaTomlInPath   bool `json:"hasBallerinaTomlInPath"`
	HasBallerinaTomlInRoot   bool `json:"hasBallerinaTomlInRoot"`
	IsDockerfilePathValid    bool `json:"isDockerfilePathValid"`
	HsDockerfileInPath       bool `json:"hasDockerfileInPath"`
	IsDockerContextPathValid bool `json:"isDockerContextPathValid"`
	IsOpenApiFilePathValid   bool `json:"isOpenApiFilePathValid"`
	HasOpenApiFileInPath     bool `json:"hasOpenApiFileInPath"`
	HasPomXmlInPath          bool `json:"hasPomXmlInPath"`
	HasPomXmlInRoot          bool `json:"hasPomXmlInRoot"`
	IsBuildpackPathValid     bool `json:"isBuildpackPathValid"`
	IsTestRunnerPathValid    bool `json:"isTestRunnerPathValid"`
	IsProcfileExists         bool `json:"isProcfileExists"`
	IsEndpointYamlExists     bool `json:"isEndpointYamlExists"`
}

type GitTokenForRepositoryResponse struct {
	Token           string `json:"token"`
	GitOrganization string `json:"gitOrganization"`
	GitRepository   string `json:"gitRepository"`
	Vendor          string `json:"vendor"`
	Username        string `json:"username"`
	ServerUrl       string `json:"serverUrl"`
}

type ProjectBuildLogsData struct {
	IntegrationProjectBuild string `json:"integrationProjectBuild"`
	LibraryTrivyReport      string `json:"libraryTrivyReport"`
	DropinsTrivyReport      string `json:"dropinsTrivyReport"`
	PostBuildCheckLogs      string `json:"postBuildCheckLogs"`
	MainSequenceValidation  string `json:"mainSequenceValidation"`
	MIVersionValidation     string `json:"mIVersionValidation"`
	ProxyBuildLogs          string `json:"proxyBuildLogs"`
	GovernanceLogs          string `json:"governanceLogs"`
	ConfigValidationLogs    string `json:"configValidationLogs"`
}

type RunPodResponse struct {
	RunId string `json:"runId"`
}
