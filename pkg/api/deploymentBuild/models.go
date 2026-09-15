package deploymentbuild

type BuildKind struct {
	ApiVersion string        `json:"apiVersion"` // core.choreo.dev/v1alpha1
	Kind       string        `json:"kind"`       // Build
	Metadata   BuildMetadata `json:"metadata"`
	Spec       BuildSpec     `json:"spec"`
	Status     *BuildStatus  `json:"status,omitempty"`
}

type BuildStatusCommit struct {
	Message string `json:"message"`
	Author  string `json:"author"`
	Date    string `json:"date"`
	Email   string `json:"email"`
}

type BuildStatusImage struct {
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type BuildStatus struct {
	RunID       int                `json:"runId"`
	Conclusion  string             `json:"conclusion"`
	Status      string             `json:"status"`
	StartedAt   string             `json:"startedAt"`
	CompletedAt string             `json:"completedAt"`
	Images      []BuildStatusImage `json:"images,omitempty"`
	GitCommit   BuildStatusCommit  `json:"gitCommit"`
	ClusterId   string             `json:"clusterId"`
	BuildRef    string             `json:"buildRef"`
}

type BuildMetadata struct {
	Name          string `json:"name"`
	ComponentName string `json:"componentName"` // component name
	ProjectName   string `json:"projectName"`   // project handle
}

type BuildSpec struct {
	Revision string `json:"revision"` // commit hash
}

type DeployKind struct {
	ApiVersion string         `json:"apiVersion"` // core.choreo.dev/v1alpha1
	Kind       string         `json:"kind"`       // Deployment
	Metadata   DeployMetadata `json:"metadata"`
	Spec       DeploySpec     `json:"spec"`
}

type DeployMetadata struct {
	Name              string `json:"name"`
	DeploymentTrackId string `json:"deploymentTrackId"`
	Environment       string `json:"environment"`
	ComponentName     string `json:"componentName"` // component name
	ProjectName       string `json:"projectName"`   // project handle
}

type DeploySpec struct {
	BuildRef string `json:"buildRef"`
	DeploySpecOpts
}

type ProxyTargetEndpoint struct {
	ProductionEndpoint string  `json:"productionEndpoint"`
	SandboxEndpoint    *string `json:"sandboxEndpoint,omitempty"`
}

type ProxyDeploymentConfig struct {
	Keys       *ProxyTargetEndpoint `json:"keys,omitempty"`
	AccessMode *string              `json:"accessMode,omitempty"`
}

type DeploySpecOpts struct {
	ScheduleExp string                 `json:"scheduleExpression,omitempty"`
	ScheduleTZ  string                 `json:"scheduleTimezone,omitempty"`
	ProxyConf   *ProxyDeploymentConfig `json:"proxyConfig,omitempty"`
}

type DeploymentTrackResponse struct {
	DeployDeploymentTrack string `json:"deployDeploymentTrack"`
}

type ProxyDeployment struct {
	APIID       string `json:"apiId"`
	Environment struct {
		Env  string `json:"choreoEnv"`
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"environment"`
	LifecycleStatus string `json:"lifecycleStatus"`
	Version         string `json:"version"`
	InvokeURL       string `json:"invokeUrl"`
	Endpoint        string `json:"endpoint"`
	SandboxEndpoint string `json:"sandboxEndpoint"`
	APIRevision     struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		CreatedTime int64  `json:"createdTime"`
	} `json:"apiRevision"`
	Build struct {
		ID                 string `json:"id"`
		BaseRevisionID     string `json:"baseRevisionId"`
		DeployedRevisionID string `json:"deployedRevisionId"`
	} `json:"build"`
	DeployedTime        int64 `json:"deployedTime"`
	SuccessDeployedTime int64 `json:"successDeployedTime"`
}

type ProxyInsightsResponse struct {
	GetTotalTrafficByAPI   int `json:"getTotalTrafficByAPI"`
	GetOverallLatencyByAPI struct {
		Response float64 `json:"response"`
		Typename string  `json:"__typename"`
	} `json:"getOverallLatencyByAPI"`
	GetTotalErrorsByAPI struct {
		Proxy    int    `json:"proxy"`
		Typename string `json:"__typename"`
	} `json:"getTotalErrorsByAPI"`
}

type ProxyPromoteResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
