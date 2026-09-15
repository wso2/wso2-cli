package region

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wso2/integration-platform-tools/internal/config"
)

// Non-production environment configuration is supplied at runtime rather than
// compiled in.
//
// The dev, stage and preview environments are internal infrastructure. Their
// hostnames, gateway identifiers and client IDs used to sit in us_config.go and
// eu_config.go, which published a map of that infrastructure once this
// repository became public. Production is unaffected -- those endpoints are the
// ones end users reach anyway.
//
// Nothing about the WSO2IP_ENV feature changes for the people who use it: set
// WSO2IP_ENV=dev|stage as before, and point WSO2IP_ENV_CONFIG at a JSON file
// describing those environments.
const (
	// EnvConfigPathVar names the file describing the non-production environments.
	EnvConfigPathVar = "WSO2IP_ENV_CONFIG"
	// EnvVar selects which of them to use.
	EnvVar = "WSO2IP_ENV"
)

// nonProdEnv mirrors the arguments of BuildEnvConfig, so a config file is a
// direct description of an environment with no separate schema to keep in sync.
type nonProdEnv struct {
	ConsoleHost     string          `json:"consoleHost"`
	APIHost         string          `json:"apiHost"`
	STSHost         string          `json:"stsHost"`
	InsightsHost    string          `json:"insightsHost"`
	AppHost         string          `json:"appHost"`
	MarketplaceHost string          `json:"marketplaceHost"`
	ConnectionsHost string          `json:"connectionsHost"`
	RegionApisPath  string          `json:"regionApisPath"`
	Overrides       RegionOverrides `json:"overrides"`
}

// nonProdFile is keyed by region ("US", "EU") then environment ("dev", "stage").
type nonProdFile map[string]map[string]nonProdEnv

// loadNonProdConfigs reads the file named by WSO2IP_ENV_CONFIG. A missing
// variable is not an error: the overwhelmingly common case is an end user on
// production, who never sets either variable.
func loadNonProdConfigs() (nonProdFile, error) {
	path := os.Getenv(EnvConfigPathVar)
	if path == "" {
		return nil, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s=%q: %w", EnvConfigPathVar, path, err)
	}
	var f nonProdFile
	if err := json.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("parsing %s=%q: %w", EnvConfigPathVar, path, err)
	}
	return f, nil
}

// nonProdConfig returns the configured environment for a region, or nil when
// the file does not describe it. A nil result leaves Region.Dev/Stage nil, and
// GetRegionConfig falls back to production rather than panicking.
func (f nonProdFile) nonProdConfig(region, env string) *config.EnvConfig {
	if f == nil {
		return nil
	}
	envs, ok := f[region]
	if !ok {
		return nil
	}
	e, ok := envs[env]
	if !ok {
		return nil
	}
	cfg := BuildEnvConfig(
		e.ConsoleHost,
		e.APIHost,
		e.STSHost,
		e.InsightsHost,
		e.AppHost,
		e.MarketplaceHost,
		e.ConnectionsHost,
		e.RegionApisPath,
		e.Overrides,
	)
	return &cfg
}

// NonProdConfigError reports a WSO2IP_ENV request that cannot be satisfied.
// Returned rather than logged so the caller decides how loudly to fail; a
// silent fall back to production would point a developer's commands at the
// live platform, which is worse than refusing.
func NonProdConfigError(region, env string) error {
	if os.Getenv(EnvConfigPathVar) == "" {
		return fmt.Errorf(
			"%s=%s requests a non-production environment, but %s is not set. "+
				"Non-production endpoints are not built into this binary; point %s at a JSON file "+
				"describing them (keyed by region, then environment)",
			EnvVar, env, EnvConfigPathVar, EnvConfigPathVar)
	}
	return fmt.Errorf(
		"%s=%s requested, but %s does not describe environment %q for region %q",
		EnvVar, env, EnvConfigPathVar, env, region)
}

// mustNonProd returns cfg, or panics with an actionable message when the
// requested non-production environment was never configured.
//
// Panicking is deliberate. The alternative -- returning the production config
// for someone who asked for dev -- would let a developer create, deploy and
// delete against the live platform while believing they were on a test
// environment. WSO2IP_ENV is a developer-only variable, so no end user reaches
// this path.
func mustNonProd(cfg *config.EnvConfig, region, env string) *config.EnvConfig {
	if cfg != nil {
		return cfg
	}
	panic(NonProdConfigError(region, env))
}
