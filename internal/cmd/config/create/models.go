package create

import "github.com/wso2/integration-platform-tools/internal/cmd/common"

type ConfigCreateOptions struct {
	orgFlag              string
	projectFlag          string
	componentFlag        string
	envFlag              string
	deploymentTrackFlag  string
	nameFlag             string
	typeFlag             string
	mountTypeFlag        string
	envVars              []common.KeyValOpt
	fileMountPathFlag    string
	fileMountContentFlag string
}
