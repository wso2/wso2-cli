package auth

import (
	cfg "github.com/wso2/integration-platform-tools/internal/config"
	"github.com/wso2/integration-platform-tools/internal/region"
)

// GetEnvConfig is now a wrapper around GetRegionConfig
func GetEnvConfig() *cfg.EnvConfig {
	return region.GetRegionConfig()
}
