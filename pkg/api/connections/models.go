package connections

type ConnectionStatus struct {
	Stage   string `json:"stage"`
	Result  string `json:"result"`
	Success bool   `json:"success"`
}

type ConnectionConfigEntry struct {
	Key             string `json:"key"`
	KeyUuid         string `json:"keyUuid"`
	Value           string `json:"value"`
	IsSensitive     bool   `json:"isSensitive"`
	IsFile          bool   `json:"isFile"`
	ValueRef        string `json:"valueRef"`
	EnvVariableName string `json:"envVariableName"`
}

type ConnectionConfig struct {
	EnvironmentUuid string                           `json:"environmentUuid"`
	Entries         map[string]ConnectionConfigEntry `json:"entries"`
}

type Connection struct {
	GroupUuid            string                      `json:"groupUuid"`
	ServiceName          string                      `json:"serviceName"`
	SchemaName           string                      `json:"schemaName"`
	Status               any                         `json:"status"`
	IsPartiallyCreated   bool                        `json:"isPartiallyCreated"`
	Configurations       map[string]ConnectionConfig `json:"configurations"`
	Name                 string                      `json:"name"`
	Description          string                      `json:"description"`
	SchemaReference      string                      `json:"schemaReference"`
	Visibilities         []ConnectionVisibility      `json:"visibilities"`
	ServiceId            string                      `json:"serviceId"`
	EnvMapping           any                         `json:"envMapping"`
	ComponentId          string                      `json:"componentId"`
	DependentComponentId string                      `json:"dependentComponentId"`
	Version              string                      `json:"version"`
	ResourceType         string                      `json:"resourceType"`
}

type ConnectionReqEnv struct {
	ID         string `json:"id"`
	IsCritical bool   `json:"isCritical"`
}

type ConnectionVisibility struct {
	ProjectUUID      string `json:"projectUuid"`
	OrganizationUUID string `json:"organizationUuid"`
	ComponentUuid    string `json:"componentUuid"`
	ComponentType    string `json:"componentType,omitempty"`
}

type ConnectionReqPayload struct {
	Name                        string                 `json:"name"`
	Description                 string                 `json:"description"`
	ServiceID                   string                 `json:"serviceId"`
	SchemaReference             string                 `json:"schemaReference"`
	Environments                []ConnectionReqEnv     `json:"environments"`
	Visibilities                []ConnectionVisibility `json:"visibilities"`
	RequestingServiceVisibility string                 `json:"requestingServiceVisibility"`
	OrgIDInteger                int                    `json:"orgIdInteger"`
}

type CreateDatabaseConnectionReq struct {
	Name            string                                 `json:"name"`
	Description     string                                 `json:"description"`
	ServiceId       string                                 `json:"serviceId"`
	SchemaReference string                                 `json:"schemaReference"`
	Visibilities    []DatabaseConnectionVisibility         `json:"visibilities"`
	EnvMapping      map[string]DatabaseConnectionEnvDetail `json:"envMapping"`
}

type DatabaseConnectionVisibility struct {
	ComponentUuid    string `json:"componentUuid"`
	OrganizationUuid string `json:"organizationUuid"`
	ProjectUuid      string `json:"projectUuid"`
}

type DatabaseConnectionEnvDetail struct {
	ResourceId         string `json:"resourceId"`
	ParameterReference string `json:"parameterReference"`
}

// Third-party connection models
type ThirdPartyConnectionEntry struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	IsFile      bool   `json:"isFile"`
	IsSensitive bool   `json:"isSensitive"`
}

type ThirdPartyConnectionConfig struct {
	EnvironmentUuid string                               `json:"environmentUuid"`
	IsCritical      bool                                 `json:"isCritical"`
	Entries         map[string]ThirdPartyConnectionEntry `json:"entries"`
}

type ThirdPartyConnectionEnvMapping struct {
	ParameterReference string `json:"parameterReference"`
	ResourceId         string `json:"resourceId"`
}

type CreateThirdPartyConnectionReq struct {
	Configurations  map[string]ThirdPartyConnectionConfig     `json:"configurations"`
	Name            string                                    `json:"name"`
	SchemaReference string                                    `json:"schemaReference"`
	Visibilities    []ConnectionVisibility                    `json:"visibilities"`
	ServiceId       string                                    `json:"serviceId"`
	Description     string                                    `json:"description"`
	EnvMapping      map[string]ThirdPartyConnectionEnvMapping `json:"envMapping"`
	ComponentType   string                                    `json:"componentType"`
}
