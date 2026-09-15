package region

import (
	clikeyring "github.com/wso2/integration-platform-tools/pkg/util/keyring"
)

const (
	REGION_ENTRY_KEY = "region_info"
	serviceName      = "wso2-integration-platform"
)

type RegionStore struct{}

func NewRegionStore() *RegionStore {
	return &RegionStore{}
}

func (r *RegionStore) SetRegion(region string) error {
	err := clikeyring.Set(serviceName, REGION_ENTRY_KEY, region)
	if err != nil {
		return err
	}
	return nil
}

func (r *RegionStore) GetRegion() (string, error) {
	region, err := clikeyring.Get(serviceName, REGION_ENTRY_KEY)
	if err != nil {
		return "", err
	}
	return region, nil
}

func (r *RegionStore) RemoveRegion() error {
	return clikeyring.Delete(serviceName, REGION_ENTRY_KEY)
}
