package configmapping

type ResolveSecretsReqItem struct {
	Key      string `json:"key"`
	ValueRef string `json:"valueRef"`
}

type ResolveSecretsReq struct {
	ProjectID     string                  `json:"projectId"`
	ComponentID   string                  `json:"componentId"`
	EnvTemplateID string                  `json:"envTemplateId"`
	Secrets       []ResolveSecretsReqItem `json:"secrets"`
}

type ResolveSecretsResponseItem struct {
	Key      string `json:"key"`
	ValueRef string `json:"valueRef"`
	Value    string `json:"value"`
}

type ResolveSecretsResponse struct {
	Secrets []ResolveSecretsResponseItem `json:"secrets"`
}

type ConfigMappingValueItem struct {
	Value           string `json:"value"`
	ValueRef        string `json:"valueRef"`
	EnvironmentUUID string `json:"environmentUuid"`
}

type ConfigMappingItem struct {
	KeyId           string                   `json:"keyId"`
	Key             string                   `json:"key"`
	Values          []ConfigMappingValueItem `json:"values"`
	IsDynamic       bool                     `json:"isDynamic"`
	IsSensitive     bool                     `json:"isSensitive"`
	IsFile          bool                     `json:"isFile"`
	ConfigGroupId   string                   `json:"configGroupId"`
	ConfigKeyId     string                   `json:"configKeyId"`
	ConfigGroupName string                   `json:"configGroupName"`
	ConfigKeyName   string                   `json:"configKeyName"`
	GroupType       string                   `json:"groupType"`
}

type ConfigMappingResponse struct {
	MappingID         string              `json:"mappingId"`
	OrganizationID    string              `json:"organizationId"`
	ProjectID         string              `json:"projectId"`
	ComponentID       string              `json:"componentId"`
	DeploymentTrackID string              `json:"deploymentTrackId"`
	EnvTemplateID     string              `json:"envTemplateId"`
	Configurations    []ConfigMappingItem `json:"configurations"`
}

type ConfigMappingValue struct {
	Value           string `json:"value"`
	EnvironmentUUID string `json:"environmentUuid"`
}

type ConfigMapping struct {
	Key           string               `json:"key"`
	IsSensitive   bool                 `json:"isSensitive"`
	IsFile        bool                 `json:"isFile"`
	IsDynamic     bool                 `json:"isDynamic"`
	ConfigGroupID string               `json:"configGroupId"`
	ConfigKeyID   string               `json:"configKeyId"`
	Values        []ConfigMappingValue `json:"values"`
}
type CreateConfigMappingReq struct {
	ProjectID         string          `json:"projectId"`
	ComponentID       string          `json:"componentId"`
	EnvTemplateID     string          `json:"envTemplateId"`
	DeploymentTrackID string          `json:"deploymentTrackId"`
	MappingID         string          `json:"mappingId"`
	Configurations    []ConfigMapping `json:"configurations"`
}
