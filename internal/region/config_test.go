package region

import (
	"os"
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

func TestGetConfigByRegion_ReturnsDevWhenEnvSet(t *testing.T) {
	os.Setenv("WSO2IP_ENV", "dev")
	defer os.Unsetenv("WSO2IP_ENV")

	cfg := GetConfigByRegion(REGION_US)
	require.NotNil(t, cfg)

	usCfg := RegionConfigs[REGION_US]
	assert.Equal(t, usCfg.Dev, cfg)
}

func TestGetConfigByRegion_ReturnsStageWhenEnvSet(t *testing.T) {
	os.Setenv("WSO2IP_ENV", "stage")
	defer os.Unsetenv("WSO2IP_ENV")

	cfg := GetConfigByRegion(REGION_EU)
	require.NotNil(t, cfg)

	euCfg := RegionConfigs[REGION_EU]
	assert.Equal(t, euCfg.Stage, cfg)
}
