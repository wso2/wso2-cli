package create

import "github.com/wso2/integration-platform-tools/internal/cmd/common"

type DeployComponentParams struct {
	EnvFlag             string
	ComponentFlag       string
	ProjectFlag         string
	OrgFlag             string
	DeploymentTrackFlag string
	EnvVars             []common.KeyValOpt
	cronDeployOpts      CronTaskConfigurations
	proxyDeployOpts     ProxyConfigurations
	byoiDeployOpts      ByoiDeployConfigurations
	RunIdFlag           string
}

type CronTaskConfigurations struct {
	Expression string
	TimeZone   string
}

type ProxyConfigurations struct {
	TargetEp  string
	SandboxEp string
}

type ByoiDeployConfigurations struct {
	ImageWithTag  string
	ApiSchemaFile []string
	EndpointsFile string
}
