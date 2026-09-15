package devops

import "time"

type GatewayInfo struct {
	ID                         string `json:"id"`
	ExternalGatewayVirtualHost string `json:"external_gateway_virtual_host"`
	InternalGatewayVirtualHost string `json:"internal_gateway_virtual_host"`
	Region                     string `json:"region"`
	IsCilium                   bool   `json:"is_cilium"`
}

type DataPlaneItemLabels struct {
	Private       bool `json:"private"`
	Logs          bool `json:"logs"`
	Observability bool `json:"observability"`
	SystemMetrics bool `json:"systemMetrics"`
}

type DataPlaneItem struct {
	ID                           string              `json:"id"`
	Name                         string              `json:"name"`
	Type                         string              `json:"type"`
	OrganizationID               int                 `json:"organizationId"`
	ProjectID                    interface{}         `json:"projectId"`
	CreatedOn                    time.Time           `json:"createdOn"`
	Labels                       DataPlaneItemLabels `json:"labels"`
	ExternalGatewayVirtualHost   string              `json:"externalGatewayVirtualHost"`
	InternalGatewayVirtualHost   string              `json:"internalGatewayVirtualHost"`
	ExternalIngressDefaultDomain string              `json:"externalIngressDefaultDomain"`
	ScaleToZeroEnabled           bool                `json:"scaleToZeroEnabled"`
	GatewayType                  string              `json:"gatewayType"`
	ActiveAgentConnections       int                 `json:"activeAgentConnections"`
	IsActive                     bool                `json:"isActive"`
}

type PromotionTreeNode struct {
	EnvTemplateId *string              `json:"env_template_id"`
	EnvName       *string              `json:"env_name"`
	Name          *string              `json:"name"`
	Children      *[]PromotionTreeNode `json:"children"`
}

type DeploymentPipelineResponse struct {
	ID               string            `json:"id"`
	CreatedAt        string            `json:"created_at"`
	OrganizationUuid string            `json:"organization_uuid"`
	Name             string            `json:"name"`
	IsDefaultProject bool              `json:"is_project_default"`
	IsDefault        string            `json:"is_default"`
	PromotionTree    PromotionTreeNode `json:"promotion_tree"`
}
type BuildPack struct {
	ID                       string `json:"id"`
	Language                 string `json:"language"`
	SupportedVersions        string `json:"supportedVersions"`
	DisplayName              string `json:"displayName"`
	IsDefault                bool   `json:"isDefault"`
	VersionEnvVariable       string `json:"versionEnvVariable"`
	IconURL                  string `json:"iconUrl"`
	Provider                 string `json:"provider"`
	BuildpackProviderOrgUUID string `json:"buidpackProviderOrgUuid"`
	Builder                  struct {
		ID           string `json:"id"`
		BuilderImage string `json:"builderImage"`
		DisplayName  string `json:"displayName"`
		ImageHash    string `json:"imageHash"`
	}
	ComponentTypes []struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Type        string `json:"type"`
	}
}

type ConfigItem struct {
	ID               string                 `json:"ID"`
	SecretType       string                 `json:"secret_type"` // will be a config-map if this equals empty string
	Metadata         map[string]interface{} `json:"metadata"`
	Name             string                 `json:"name"`
	OrganizationID   string                 `json:"organization_id"`
	ProjectID        string                 `json:"project_id"`
	EnvironmentID    string                 `json:"environment_id"`
	Data             map[string]string      `json:"data"`
	KubernetesName   string                 `json:"kubernetes_name"`
	AppEnvironmentID string                 `json:"app_environment_id"`
	Keys             []string               `json:"keys"`
	ConfigType       string                 `json:"config_type"`
	Versions         []struct {
		ID        string                 `json:"ID"`
		CreatedAt time.Time              `json:"CreatedAt"`
		UpdatedAt time.Time              `json:"UpdatedAt"`
		Metadata  map[string]interface{} `json:"metadata"`
		Version   int                    `json:"Version"`
		ParentID  string                 `json:"ParentID"`
	} `json:"versions"`
	Version        int         `json:"version"`
	Placeholder    bool        `json:"placeholder"`
	IsBase64       bool        `json:"isBase64"`
	Namespace      string      `json:"namespace"`
	CopiedFrom     interface{} `json:"copied_from"`
	CreationSource string      `json:"creation_source"`
	CreatedAt      time.Time   `json:"CreatedAt"`
	UpdatedAt      time.Time   `json:"UpdatedAt"`
	Environment    interface{} `json:"environment"`
}

type ConfigMountData struct {
	ConfigMapID      string    `json:"configmap_id"` // Will be only available for config-maps
	SecretID         string    `json:"secret_id"`    // Will be only available for secrets
	ContainerID      string    `json:"container_id"`
	AppEnvironmentID string    `json:"app_environment_id"`
	MountPath        string    `json:"mount_path"`
	KeyOverride      string    `json:"key_override"`
	ConfigKey        string    `json:"config_key"`
	MountType        string    `json:"mount_type"`
	MountPermissions string    `json:"mount_permissions"`
	ID               string    `json:"ID"`
	CreatedAt        time.Time `json:"CreatedAt"`
	UpdatedAt        time.Time `json:"UpdatedAt"`
	Status           bool      `json:"status"`
}

type ConfigReleaseContainer struct {
	Name                string        `json:"name"`
	Metadata            interface{}   `json:"metadata"`
	ImageRegistryID     string        `json:"image_registry_id"`
	ImageID             string        `json:"image_id"`
	Image               interface{}   `json:"image"`
	CustomImage         string        `json:"custom_image"`
	CPU                 int           `json:"cpu"`
	Memory              int           `json:"memory"`
	CPULimit            int           `json:"cpu_limit"`
	MemoryLimit         int           `json:"memory_limit"`
	LimitDisabled       bool          `json:"limit_disabled"`
	Ports               []interface{} `json:"ports"`
	Type                string        `json:"type"`
	ImagePullPolicy     string        `json:"image_pull_policy"`
	Args                interface{}   `json:"args"`
	Command             interface{}   `json:"command"`
	ID                  string        `json:"ID"`
	CreatedAt           time.Time     `json:"CreatedAt"`
	UpdatedAt           time.Time     `json:"UpdatedAt"`
	AppArmorProfileName string        `json:"apparmor_profile_name"`
	SecurityContext     interface{}   `json:"securityContext"`
	ImageRegistry       interface{}   `json:"image_registry"`
	AppEnvironmentID    string        `json:"app_environment_id"`
	AppEnvironment      interface{}   `json:"app_environment"`
}

type ContainerRegistry struct {
	Id               string    `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	Name             string    `json:"name"`
	OrganizationId   int       `json:"organization_id"`
	OrganizationUuid string    `json:"organization_uuid"`
	Type             string    `json:"type"`
	Provider         string    `json:"provider"`
	Scope            string    `json:"scope"`
	ReferenceToken   string    `json:"reference_token"`
	Metadata         string    `json:"metadata"`
	Host             string    `json:"host"`
}

type ContainerImage struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageUrl    string `json:"image_url"`
	Type        string `json:"type"`
	IsTemporary bool   `json:"is_temporary"`
}

type ImageHistoryResponse struct {
	ID               string   `json:"ID"`
	ImageNameWithTag string   `json:"image_name_with_tag"`
	ImageName        string   `json:"image_name"`
	Tags             []string `json:"tags"`
	CreatedAt        string   `json:"CreatedAt"`
	UpdatedAt        string   `json:"UpdatedAt"`
	ImageRegistryID  string   `json:"image_registry_id"`
	TriggerSource    string   `json:"trigger_source"` // "MANUAL" or "EXTERNAL"
}

type CreateApiSchemaRequest struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}
type CreateByoiEndpointsRequest struct {
	Main       string                   `json:"main"`
	ApiSchemas []CreateApiSchemaRequest `json:"apiSchemas"`
}
type DeploymentHistoryEntry struct {
	ID                 string            `json:"ID"`
	CreatedAt          time.Time         `json:"CreatedAt"`
	AppEnvironmentID   string            `json:"app_environment_id"`
	ContainerImages    []ContainerImages `json:"container_images"`
	Metadata           any               `json:"metadata"`
	DoraDeploymentBool bool              `json:"DoraDeploymentBool"`
	ChangeMessage      string            `json:"change_message"`
	DeploymentStatus   any               `json:"deployment_status"`
	UpdatedAt          time.Time         `json:"UpdatedAt"`
	AppEnvironment     any               `json:"app_environment"`
}

type ClusterImageTags struct {
	RegistryID       string   `json:"registry_id"`
	Clusters         []string `json:"clusters"`
	ImageNameWithTag string   `json:"image_name_with_tag"`
}

type Image struct {
	GitHashCommitTimestamp time.Time          `json:"git_hash_commit_timestamp"`
	GitHash                string             `json:"git_hash"`
	Metadata               any                `json:"metadata"`
	IsBalImage             bool               `json:"is_bal_image"`
	CreatedAt              time.Time          `json:"CreatedAt"`
	ImagePorts             any                `json:"image_ports"`
	APIVersionID           string             `json:"api_version_id"`
	ImageNameWithTag       string             `json:"image_name_with_tag"`
	TriggerSource          string             `json:"trigger_source"`
	CommitMsg              string             `json:"commit_msg"`
	ImageName              string             `json:"image_name"`
	Committer              string             `json:"committer"`
	ProjectID              string             `json:"project_id"`
	BuiltAt                time.Time          `json:"built_at"`
	ClusterImageTags       []ClusterImageTags `json:"cluster_image_tags"`
	ID                     string             `json:"ID"`
	RunID                  string             `json:"run_id"`
	TagName                string             `json:"tag_name"`
	UpdatedAt              time.Time          `json:"UpdatedAt"`
	ImageVersion           any                `json:"image_version"`
	ImageRegistryID        string             `json:"image_registry_id"`
	Tags                   []string           `json:"tags"`
	GitOpsHash             string             `json:"git_ops_hash"`
	ImageRegistry          any                `json:"image_registry"`
	PlatformerTag          string             `json:"platformer_tag"`
	OrganizationID         string             `json:"organization_id"`
	Status                 string             `json:"status"`
}

type ContainerImages struct {
	ImageID                string `json:"image_id"`
	Image                  Image  `json:"image"`
	Container              any    `json:"container"`
	ContainerPorts         []any  `json:"ContainerPorts"`
	CustomImageNameWithTag string `json:"CustomImageNameWithTag"`
	ContainerID            string `json:"container_id"`
}

type GetEnvironmentTemplatesResp struct {
	Data []EnvironmentTemplate `json:"data"`
}

type EnvironmentTemplate struct {
	ID                   string `json:"id"`
	CreatedAt            string `json:"created_at"`
	OrganizationID       int    `json:"organization_id"`
	OrganizationUUID     string `json:"organization_uuid"`
	EnvName              string `json:"env_name"`
	Region               string `json:"region"`
	Env                  string `json:"choreo_env"`
	ClusterID            string `json:"cluster_id"`
	DockerCredentialUUID string `json:"docker_credential_uuid"`
	ExternalApimEnvName  string `json:"external_apim_env_name"`
	InternalApimEnvName  string `json:"internal_apim_env_name"`
	SandboxApimEnvName   string `json:"sandbox_apim_env_name"`
	Critical             bool   `json:"critical"`
	DNSPrefix            string `json:"dns_prefix"`
	PDPWebAppDNSPrefix   string `json:"pdp_web_app_dns_prefix"`
	DeletionStatus       string `json:"deletion_status"`
	Sandbox              bool   `json:"sandbox"`
	CanReadSecrets       bool   `json:"can_read_secrets"`
}
