package region

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/wso2/integration-platform-tools/internal/config"
)

const (
	REGION_US = "US"
	REGION_EU = "EU"
)

// Region represents a WSO2 Integration Platform deployment region
type Region struct {
	Name  string
	Dev   *config.EnvConfig
	Stage *config.EnvConfig
	Prod  *config.EnvConfig
}

var (
	currentRegion     string
	currentRegionLock sync.RWMutex
	RegionConfigs     = make(map[string]Region)
	regionStore       *RegionStore
)

func init() {
	regionStore = NewRegionStore()
}

// GetCurrentRegion returns the currently selected region
func GetCurrentRegion() string {
	currentRegionLock.RLock()
	defer currentRegionLock.RUnlock()

	// Check for environment variable override first
	if envRegion := os.Getenv("WSO2IP_REGION"); envRegion != "" {
		if _, exists := RegionConfigs[envRegion]; exists {
			return envRegion
		}
		log.Printf("Warning: WSO2IP_REGION environment variable is set to '%s' but this is not a valid region. Valid regions are: %v", envRegion, GetValidRegions())
	}

	if currentRegion == "" {
		// Try to get from persistent store
		region, err := regionStore.GetRegion()
		if err != nil || region == "" {
			return REGION_US // Default to US region
		}
		currentRegion = region
	}
	return currentRegion
}

// GetValidRegions returns a slice of valid region names
func GetValidRegions() []string {
	regions := make([]string, 0, len(RegionConfigs))
	for region := range RegionConfigs {
		regions = append(regions, region)
	}
	return regions
}

// SetCurrentRegion sets the current region
func SetCurrentRegion(region string) error {
	if _, exists := RegionConfigs[region]; !exists {
		return fmt.Errorf("invalid region: %s", region)
	}

	currentRegionLock.Lock()
	defer currentRegionLock.Unlock()

	// Store in persistent storage
	err := regionStore.SetRegion(region)
	if err != nil {
		return fmt.Errorf("failed to persist region: %v", err)
	}

	currentRegion = region
	return nil
}

// GetRegionConfig returns the appropriate configuration based on region and environment
func GetRegionConfig() *config.EnvConfig {
	region := GetCurrentRegion()
	regionConfig := RegionConfigs[region]

	// For developers: Allow env override for non-prod environments
	if env := os.Getenv("WSO2IP_ENV"); env != "" {
		switch env {
		case config.ENV_DEV:
			return regionConfig.Dev
		case config.ENV_STAGE:
			return regionConfig.Stage
		}
	}

	// For end users: Always use prod of the selected region
	return regionConfig.Prod
}

// Get config for given region name
func GetConfigByRegion(regionStr string) *config.EnvConfig {
	regionConfig := RegionConfigs[regionStr]

	// For developers: Allow env override for non-prod environments
	if env := os.Getenv("WSO2IP_ENV"); env != "" {
		switch env {
		case config.ENV_DEV:
			return regionConfig.Dev
		case config.ENV_STAGE:
			return regionConfig.Stage
		}
	}

	// For end users: Always use prod of the selected region
	return regionConfig.Prod
}

// Initialize US region configurations
func InitUSRegion(dev, stage, prod *config.EnvConfig) {
	RegionConfigs[REGION_US] = Region{
		Name:  REGION_US,
		Dev:   dev,
		Stage: stage,
		Prod:  prod,
	}
}

// Initialize EU region configurations
func InitEURegion(dev, stage, prod *config.EnvConfig) {
	RegionConfigs[REGION_EU] = Region{
		Name:  REGION_EU,
		Dev:   dev,
		Stage: stage,
		Prod:  prod,
	}
}
