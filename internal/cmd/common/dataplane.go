package common

import (
	"fmt"
	"strings"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/devops"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

func GetDataPlaneInfo(orgId string, orgUuid string) (cloudDataPlanes []devops.GatewayInfo, dataPlanes []devops.DataPlaneItem, err error) {
	dataPlaneInfo := utils.CreateSpinner(i18n.T(" Fetching data plane information..."), "")
	dataPlaneInfo.Start()

	cloudDataPlanes, err = auth.DevopsClient.GetCloudPlaneClusters(orgUuid, orgId)
	if err != nil {
		return nil, nil, fmt.Errorf(i18n.T("failed to fetch cloud data plane information: %w"), err)
	}

	dataPlanes, err = auth.DevopsClient.GetDataPlaneClusters(orgId)
	if err != nil {
		return nil, nil, fmt.Errorf(i18n.T("failed to fetch data plane information: %w"), err)
	}

	dataPlaneInfo.Stop()

	return cloudDataPlanes, dataPlanes, nil
}

func GetDataPlaneHost(cloudDataPlanes []devops.GatewayInfo, dataPlanes []devops.DataPlaneItem, clusterId string) (selectedDataPlaneHost string, err error) {
	for _, dataPlaneItem := range cloudDataPlanes {
		if strings.ToLower(dataPlaneItem.ID) == strings.ToLower(clusterId) {
			return dataPlaneItem.ExternalGatewayVirtualHost, nil
		}
	}

	for _, dataPlaneItem := range dataPlanes {
		if strings.ToLower(dataPlaneItem.ID) == strings.ToLower(clusterId) {
			return dataPlaneItem.ExternalGatewayVirtualHost, nil
		}
	}

	return "", fmt.Errorf(i18n.T("failed to find data plane host: %w"), err)
}

func ResolveDataPlane(selectedEnv project.ProjectEnvironment, cloudDataPlanes []devops.GatewayInfo, dataPlanes []devops.DataPlaneItem) (selectedDataPlaneHost string, isCilium bool) {
	for _, dataPlaneItem := range cloudDataPlanes {
		if dataPlaneItem.ID == selectedEnv.DPID {
			selectedDataPlaneHost = dataPlaneItem.ExternalGatewayVirtualHost
			isCilium = dataPlaneItem.IsCilium
			break
		}
	}

	if selectedDataPlaneHost == "" {
		for _, dataPlaneItem := range dataPlanes {
			if dataPlaneItem.ID == selectedEnv.DPID {
				selectedDataPlaneHost = dataPlaneItem.ExternalGatewayVirtualHost
				break
			}
		}
	}

	return selectedDataPlaneHost, isCilium
}
