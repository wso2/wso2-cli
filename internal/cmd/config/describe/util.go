package describe

import (
	"fmt"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/devops"
)

func getConfigMapData(orgId string, orgUuid, envId string, projectId string, configId string) (devops.ConfigItem, error) {
	secretsSpinner := utils.CreateSpinner(i18n.T(" Fetching config-map data..."), "")
	secretsSpinner.Start()
	configMap, err := auth.DevopsClient.GetConfigMapDetails(orgId, orgUuid, envId, configId, projectId)
	secretsSpinner.Stop()

	if err != nil {
		return devops.ConfigItem{}, fmt.Errorf("%s", i18n.T("failed to fetch config-map data"))
	}

	return configMap, nil
}

func getSecretData(orgId string, orgUuid, envId string, projectId string, configId string) (devops.ConfigItem, error) {
	secretsSpinner := utils.CreateSpinner(i18n.T(" Fetching secret config data..."), "")
	secretsSpinner.Start()
	configMap, err := auth.DevopsClient.GetSecretDetails(orgId, orgUuid, envId, configId, projectId)
	secretsSpinner.Stop()

	if err != nil {
		return devops.ConfigItem{}, fmt.Errorf("%s", i18n.T("failed to fetch secret data"))
	}

	return configMap, nil
}

// emitConfigJSON renders the resolved config as a JSON object. It is only
// called when --output=json. UX affordances (the "to add a config" / "to list
// configs" hints, the 1s sleep) are intentionally skipped: the JSON payload
// describes the resource, not the CLI's suggestions.
//
// Secret values are redacted exactly as the text path redacts them — for a
// secret, only the key names are emitted (with redacted values), never the
// secret material. The JSON path must not become a way to exfiltrate secrets.
func emitConfigJSON(configItem devops.ConfigItem, mountItems []devops.ConfigMountData, format common.OutputFormat) error {
	out := ConfigDescribeOutput{
		ID:         configItem.ID,
		Name:       configItem.Name,
		ConfigType: configItem.ConfigType,
		Version:    configItem.Version,
		CreatedAt:  configItem.CreatedAt.String(),
		UpdatedAt:  configItem.UpdatedAt.String(),
		IsSecret:   configItem.SecretType != "",
		Data:       map[string]string{},
	}

	if out.IsSecret {
		out.Type = "secret"
		if configItem.ConfigType == "File" {
			out.Data["data"] = "*** (Data Redacted) ***"
		} else {
			for _, key := range configItem.Keys {
				out.Data[key] = "*****"
			}
		}
	} else {
		out.Type = "config-map"
		for key, value := range configItem.Data {
			out.Data[key] = value
		}
	}

	if configItem.ConfigType == "File" {
		mountItem, err := common.GetMountForConfig(configItem, mountItems)
		if err != nil {
			return err
		}
		out.MountPath = mountItem.MountPath
	}

	rendered, err := common.RenderStructured(format, out)
	if err != nil {
		return err
	}
	fmt.Fprintln(utils.IO.Out, rendered)
	return nil
}

func printConfigDetails(configItem devops.ConfigItem, mountItems []devops.ConfigMountData) error {
	dataKey, data := "", ""

	configType := "Config-map"
	if configItem.SecretType != "" {
		configType = "Secret"

		if configItem.ConfigType == "File" {
			data = "*** (Data Redacted) ***"
		} else {
			for _, key := range configItem.Keys {
				data += fmt.Sprintf("%s: *****\n", key)
			}
		}
	} else {
		if configItem.ConfigType == "File" {
			data = configItem.Data["data"]
		} else {
			for key, value := range configItem.Data {
				data += fmt.Sprintf("%s: %s\n", key, value)
			}
		}
	}

	mountPath := "N/A"
	if configItem.ConfigType == "File" {
		dataKey = "File Content"
		mountItem, err := common.GetMountForConfig(configItem, mountItems)
		if err != nil {
			return err
		}

		mountPath = mountItem.MountPath
	} else {
		dataKey = "Environment Variables"
	}

	utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

			Config details:
			
			Version:        %s
			ID:             %s
			Name:           %s
			Type:         	%s
			Mount path:     %s
			Created:     	%s
			Last updated:   %s

			%s:
			%s

		`),
		strconv.Itoa(configItem.Version),
		configItem.ID,
		configItem.Name,
		configType,
		mountPath,
		utils.TimeAgo(configItem.CreatedAt),
		utils.TimeAgo(configItem.UpdatedAt),
		dataKey,
		data,
	))

	return nil
}
