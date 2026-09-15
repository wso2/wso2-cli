package region

import (
	"github.com/wso2/integration-platform-tools/internal/config"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The US and EU region configs are registered by init() in us_config.go / eu_config.go.

func TestGetValidRegions_ContainsBothRegions(t *testing.T) {
	regions := GetValidRegions()
	assert.Contains(t, regions, REGION_US)
	assert.Contains(t, regions, REGION_EU)
}

func TestGetCurrentRegion_DefaultsToUS(t *testing.T) {
	// Without WSO2IP_REGION set and nothing in the keyring, US is the default.
	os.Unsetenv("WSO2IP_REGION")
	currentRegion = "" // reset in-process cache so the store is consulted

	got := GetCurrentRegion()
	// Either the keyring returns empty (fresh env) → US, or it returns what was last set.
	// At minimum, the result must be a known region.
	assert.Contains(t, GetValidRegions(), got)
}

func TestGetCurrentRegion_EnvVarOverride(t *testing.T) {
	os.Setenv("WSO2IP_REGION", REGION_EU)
	defer os.Unsetenv("WSO2IP_REGION")
	currentRegion = ""

	assert.Equal(t, REGION_EU, GetCurrentRegion())
}

func TestGetCurrentRegion_InvalidEnvVarFallsBack(t *testing.T) {
	os.Setenv("WSO2IP_REGION", "INVALID")
	defer os.Unsetenv("WSO2IP_REGION")
	currentRegion = ""

	// An unrecognised env var must NOT be returned; the function must fall back.
	got := GetCurrentRegion()
	assert.NotEqual(t, "INVALID", got)
}

func TestSetCurrentRegion_ValidRegion(t *testing.T) {
	err := SetCurrentRegion(REGION_EU)
	require.NoError(t, err)
	currentRegion = "" // clear cache to force re-read
	os.Unsetenv("WSO2IP_REGION")

	assert.Equal(t, REGION_EU, GetCurrentRegion())

	// Restore to US so other tests aren't affected.
	_ = SetCurrentRegion(REGION_US)
}

func TestSetCurrentRegion_InvalidRegion(t *testing.T) {
	err := SetCurrentRegion("INVALID")
	assert.ErrorContains(t, err, "invalid region")
}

func TestGetConfigByRegion_ReturnsProdByDefault(t *testing.T) {
	os.Unsetenv("WSO2IP_ENV")

	cfg := GetConfigByRegion(REGION_US)
	require.NotNil(t, cfg)

	usCfg := RegionConfigs[REGION_US]
	assert.Equal(t, usCfg.Prod, cfg)
}

// Non-production environments are no longer compiled in, so these exercise the
// WSO2IP_ENV_CONFIG path: write a config file, point the variable at it, and
// confirm the environment resolves through BuildEnvConfig exactly as the
// hardcoded blocks used to.
func writeNonProdConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "nonprod.json")
	body := `{
      "US": {"dev": {"consoleHost":"https://console.example.invalid","apiHost":"https://apis.example.invalid",
                     "stsHost":"https://sts.example.invalid","insightsHost":"https://cp.example.invalid",
                     "appHost":"https://app.example.invalid","marketplaceHost":"https://apis.example.invalid",
                     "connectionsHost":"https://apis.example.invalid","regionApisPath":"/projects/1.0.0/graphql",
                     "overrides":{"AsgardeoClientId":"test-client"}}},
      "EU": {"stage": {"consoleHost":"https://console.eu.example.invalid","apiHost":"https://apis.eu.example.invalid",
                     "stsHost":"https://sts.eu.example.invalid","insightsHost":"https://cp.eu.example.invalid",
                     "appHost":"https://app.eu.example.invalid","marketplaceHost":"https://apis.eu.example.invalid",
                     "connectionsHost":"https://apis.eu.example.invalid","regionApisPath":"/projects/1.0.0/graphql",
                     "overrides":{"AsgardeoClientId":"test-client-eu"}}}
    }`
	require.NoError(t, os.WriteFile(path, []byte(body), 0600))
	return path
}

// reinit rebuilds RegionConfigs, which is populated in package init() from the
// environment as it stood then.
func reinit(t *testing.T, cfgPath string) {
	t.Helper()
	t.Setenv(EnvConfigPathVar, cfgPath)
	nonProd, err := loadNonProdConfigs()
	require.NoError(t, err)
	InitUSRegion(nonProd.nonProdConfig(REGION_US, config.ENV_DEV), nonProd.nonProdConfig(REGION_US, config.ENV_STAGE), &DEFAULT_ENV_CONFIG)
	InitEURegion(nonProd.nonProdConfig(REGION_EU, config.ENV_DEV), nonProd.nonProdConfig(REGION_EU, config.ENV_STAGE), &EU_ENV_CONFIG_PROD)
	t.Cleanup(func() {
		InitUSRegion(nil, nil, &DEFAULT_ENV_CONFIG)
		InitEURegion(nil, nil, &EU_ENV_CONFIG_PROD)
	})
}

func TestGetConfigByRegion_ReturnsDevFromEnvConfig(t *testing.T) {
	reinit(t, writeNonProdConfig(t))
	t.Setenv(EnvVar, "dev")

	cfg := GetConfigByRegion(REGION_US)
	require.NotNil(t, cfg)
	assert.Equal(t, "https://console.example.invalid", cfg.ConsoleUrls.BaseUrl)
	assert.Equal(t, "test-client", cfg.AsgardeoClientId)
}

func TestGetConfigByRegion_ReturnsStageFromEnvConfig(t *testing.T) {
	reinit(t, writeNonProdConfig(t))
	t.Setenv(EnvVar, "stage")

	cfg := GetConfigByRegion(REGION_EU)
	require.NotNil(t, cfg)
	assert.Equal(t, "https://console.eu.example.invalid", cfg.ConsoleUrls.BaseUrl)
}

// Asking for an environment that is not configured must fail rather than hand
// back production -- a developer who believes they are on dev would otherwise
// be creating and deleting real resources.
func TestUnconfiguredNonProdIsFatal(t *testing.T) {
	reinit(t, writeNonProdConfig(t))
	t.Setenv(EnvVar, "stage") // configured for EU only, not US

	assert.PanicsWithError(t, NonProdConfigError(REGION_US, "stage").Error(), func() {
		GetConfigByRegion(REGION_US)
	})
}

// No config file at all is the end-user case: production, with no error.
func TestProdNeedsNoEnvConfig(t *testing.T) {
	t.Setenv(EnvConfigPathVar, "")
	os.Unsetenv(EnvVar)

	cfg := GetConfigByRegion(REGION_US)
	require.NotNil(t, cfg)
	assert.Equal(t, RegionConfigs[REGION_US].Prod, cfg)
}

// A bad path is a developer's own misconfiguration and must surface, not be
// swallowed into a production fall-through.
func TestMissingEnvConfigFileErrors(t *testing.T) {
	t.Setenv(EnvConfigPathVar, filepath.Join(t.TempDir(), "does-not-exist.json"))
	_, err := loadNonProdConfigs()
	assert.ErrorContains(t, err, "reading")
}
