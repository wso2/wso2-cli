package common

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api/devops"
)

const (
	ConfigTypeConfigMap = "config-map"
	ConfigTypeSecret    = "secret"
	MountTypeFile       = "file mount"
	MountTypeEnv        = "env variable"
)

func AddEnvVarsFlag(cmdFlags *pflag.FlagSet) {
	cmdFlags.String("env-vars", "", i18n.T(`list of env variables (eg: --env-vars="key1=val1,key2=val2")`))
}

var ConfigTypes = []string{ConfigTypeConfigMap, ConfigTypeSecret}

var MountTypes = []string{MountTypeEnv, MountTypeFile}

func getConfigMaps(orgId string, orgUuid, envId string, projectId string, appEnvironmentID string) ([]devops.ConfigItem, error) {
	configMapSpinner := utils.CreateSpinner(i18n.T(" Fetching config-maps..."), "")
	configMapSpinner.Start()
	configMaps, err := auth.DevopsClient.GetConfigMapList(orgId, orgUuid, envId, projectId)
	configMapSpinner.Stop()

	if err != nil {
		return nil, fmt.Errorf(i18n.T("failed to fetch config-map list: %w"), err)
	}

	filteredConfigs := make([]devops.ConfigItem, 0)

	for _, config := range configMaps {
		if config.AppEnvironmentID == appEnvironmentID {
			filteredConfigs = append(filteredConfigs, config)
		}
	}

	return filteredConfigs, nil
}

func getSecrets(orgId string, orgUuid, envId string, projectId string, appEnvironmentID string) ([]devops.ConfigItem, error) {
	secretsSpinner := utils.CreateSpinner(i18n.T(" Fetching secrets..."), "")
	secretsSpinner.Start()
	secrets, err := auth.DevopsClient.GetSecretsList(orgId, orgUuid, envId, projectId)
	secretsSpinner.Stop()

	if err != nil {
		return nil, fmt.Errorf(i18n.T("failed to fetch secret list: %w"), err)
	}

	filteredSecrets := make([]devops.ConfigItem, 0)

	for _, config := range secrets {
		if config.AppEnvironmentID == appEnvironmentID {
			filteredSecrets = append(filteredSecrets, config)
		}
	}

	return filteredSecrets, nil
}

func GetReleaseContainers(orgId string, orgUuid string, componentId string, appEnvironmentID string, projectId string) ([]devops.ConfigReleaseContainer, error) {
	releaseContainersSpinner := utils.CreateSpinner(i18n.T(" Fetching release containers..."), "")
	releaseContainersSpinner.Start()
	releaseContainers, err := auth.DevopsClient.GetReleaseContainers(orgId, orgUuid, componentId, appEnvironmentID, projectId)
	releaseContainersSpinner.Stop()

	if err != nil {
		return nil, fmt.Errorf(i18n.T("failed to fetch release container list: %w"), err)
	}

	return releaseContainers, nil
}

func GetConfigMounts(orgId string, orgUuid string, componentId string, appEnvironmentID string, projectId string) ([]devops.ConfigMountData, error) {
	releaseContainers, err := GetReleaseContainers(orgId, orgUuid, componentId, appEnvironmentID, projectId)
	if err != nil {
		return nil, err
	}

	allConfigMounts := make([]devops.ConfigMountData, 0)

	configMountsSpinner := utils.CreateSpinner(i18n.T(" Fetching config mounts..."), "")
	configMountsSpinner.Start()

	for _, container := range releaseContainers {
		configMounts, err := auth.DevopsClient.GetConfigMounts(orgId, orgUuid, componentId, appEnvironmentID, container.ID, projectId)
		if err != nil {
			configMountsSpinner.Stop()
			return nil, fmt.Errorf(i18n.T("failed to fetch config mount list: %w"), err)
		}
		allConfigMounts = append(allConfigMounts, configMounts...)
	}

	configMountsSpinner.Stop()

	return allConfigMounts, nil
}

func GetConfigsOfComponent(orgId string, orgUuid, envId string, projectId string, appEnvironmentID string, componentId string, allConfigMounts []devops.ConfigMountData) ([]devops.ConfigItem, error) {
	configMaps, err := getConfigMaps(orgId, orgUuid, envId, projectId, appEnvironmentID)
	if err != nil {
		return nil, err
	}

	secrets, err := getSecrets(orgId, orgUuid, envId, projectId, appEnvironmentID)
	if err != nil {
		return nil, err
	}

	mergedConfigs := make([]devops.ConfigItem, 0)

	for _, mountItem := range allConfigMounts {
		if mountItem.ConfigMapID != "" {
			for _, configMapItem := range configMaps {
				if configMapItem.ID == mountItem.ConfigMapID {
					mergedConfigs = append(mergedConfigs, configMapItem)
					break
				}
			}
		}

		if mountItem.SecretID != "" {
			for _, secretItem := range secrets {
				if secretItem.ID == mountItem.SecretID {
					mergedConfigs = append(mergedConfigs, secretItem)
					break
				}
			}
		}
	}

	return mergedConfigs, nil
}

func GetMountForConfig(configItem devops.ConfigItem, configMounts []devops.ConfigMountData) (devops.ConfigMountData, error) {
	for _, mountItem := range configMounts {
		if configItem.SecretType == "" && mountItem.ConfigMapID == configItem.ID {
			return mountItem, nil
		} else if mountItem.SecretID == configItem.ID {
			return mountItem, nil
		}
	}
	return devops.ConfigMountData{}, errors.New("unable to find mount for config")
}

// convert key1=val1,key2=val2 to slice of EnvVar
func ConvertStrToEnvVarSlice(envFlagsStr string) []KeyValOpt {
	envVars := []KeyValOpt{}
	envPairs := strings.Split(envFlagsStr, ",")

	for _, pair := range envPairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			envVars = append(envVars, KeyValOpt{Key: parts[0], Val: parts[1]})
		}
	}

	return envVars
}

func CreateConfig(params CreateConfigParams) error {
	configId := ""
	var resp devops.ConfigItem
	var err error

	createConfigSpinner := utils.CreateSpinner(i18n.T(" Creating config map..."), "")
	createConfigSpinner.Start()
	if params.MountType == MountTypeEnv {
		envVarsData := make(map[string]string)
		for _, item := range params.EnvVars {
			envVarsData[item.Key] = item.Val
		}

		if params.ConfigType == ConfigTypeSecret {
			resp, err = auth.DevopsClient.CreateConfigSecret(
				params.OrgId,
				params.OrgUuid,
				params.ProjectId,
				params.EnvId,
				"VariableList",
				params.ConfigName,
				params.AppEnvId,
				envVarsData,
			)
		} else {
			resp, err = auth.DevopsClient.CreateConfigMap(
				params.OrgId,
				params.OrgUuid,
				params.ProjectId,
				params.EnvId,
				"VariableList",
				params.ConfigName,
				params.AppEnvId,
				envVarsData,
			)
		}

	} else {
		type ConfigData struct {
			Data string "json:\"data\""
		}
		if params.ConfigType == ConfigTypeSecret {
			resp, err = auth.DevopsClient.CreateConfigSecret(
				params.OrgId,
				params.OrgUuid,
				params.ProjectId,
				params.EnvId,
				"File",
				params.ConfigName,
				params.AppEnvId,
				ConfigData{Data: params.FileMountContent},
			)
		} else {
			resp, err = auth.DevopsClient.CreateConfigMap(
				params.OrgId,
				params.OrgUuid,
				params.ProjectId,
				params.EnvId,
				"File",
				params.ConfigName,
				params.AppEnvId,
				ConfigData{Data: params.FileMountContent},
			)
		}
	}
	createConfigSpinner.Stop()

	if err != nil {
		return err
	}
	configId = resp.ID

	releaseContainers, err := GetReleaseContainers(params.OrgId, params.OrgUuid, params.ComponentId, params.AppEnvId, params.ProjectId)
	if err != nil {
		return err
	}

	createMountSpinner := utils.CreateSpinner(i18n.T(" Creating config mount..."), "")
	createMountSpinner.Start()
	for _, container := range releaseContainers {
		var configMapId, secretId *string
		if params.ConfigType == ConfigTypeSecret {
			secretId = &configId
		} else {
			configMapId = &configId
		}

		if params.MountType == MountTypeEnv {
			err = auth.DevopsClient.CreateConfigMount(
				params.OrgId,
				params.OrgUuid,
				params.ProjectId,
				params.ComponentId,
				container.ID,
				params.AppEnvId,
				configMapId,
				"",
				"0000",
				"ENVFile",
				"",
				secretId,
			)
		} else {
			err = auth.DevopsClient.CreateConfigMount(
				params.OrgId,
				params.OrgUuid,
				params.ProjectId,
				params.ComponentId,
				container.ID,
				params.AppEnvId,
				configMapId,
				params.FileMountPath,
				"0644",
				"File",
				"data",
				secretId,
			)
		}

		if err != nil {
			createMountSpinner.Stop()
			return err
		}
	}
	createMountSpinner.Stop()

	return nil
}

func promptToSelectConfigMap(configs []devops.ConfigItem) (string, error) {
	var selectedConfigName string
	var configNames []string

	for _, item := range configs {
		configNames = append(configNames, item.Name)
	}

	// itemNamePrompt := &survey.Select{Message: "Config:", Options: configNames}
	// err := utils.PromptSelection(
	// 	itemNamePrompt,
	// 	&selectedConfigName,
	// 	survey.WithValidator(survey.Required),
	// )

	err := prompt.NewPromptSelectMessage[string](
		prompt.PromptSelectOpts[string]{Message: "Config:", Values: configNames},
		&selectedConfigName,
	).Prompt()

	if err != nil {
		return "", err
	}

	return selectedConfigName, nil
}

func ResolveTargetConfig(configs []devops.ConfigItem, configFlag string) (devops.ConfigItem, error) {
	var selectedConfig *devops.ConfigItem
	if configFlag != "" {
		for _, configItem := range configs {
			if configItem.Name == configFlag {
				selectedConfig = &configItem
				break
			}
		}
	} else {
		selectedConfigName, err := promptToSelectConfigMap(configs)
		if err != nil {
			if errors.Is(err, internal.ErrNonInteractive) {
				return devops.ConfigItem{}, utils.CreateNonInteractiveError("config selection", "name")
			}
			return devops.ConfigItem{}, fmt.Errorf("%s", i18n.T(" failed to select the config"))
		}
		for _, configItem := range configs {
			if configItem.Name == selectedConfigName {
				selectedConfig = &configItem
				break
			}
		}
	}

	if selectedConfig == nil {
		return devops.ConfigItem{}, fmt.Errorf("%s", i18n.T(" invalid config selection"))
	}

	return *selectedConfig, nil
}

func DeleteConfigMount(orgId string, orgUuid string, componentId string, releaseId string, containerId string, configMountId string, projectId string) error {
	deleteConfigSpinner := utils.CreateSpinner(i18n.T(" Deleting config mount..."), "")
	deleteConfigSpinner.Start()
	err := auth.DevopsClient.DeleteConfigMount(orgId, orgUuid, componentId, releaseId, containerId, configMountId, projectId)
	deleteConfigSpinner.Stop()

	if err != nil {
		return errors.New(i18n.T("failed to delete config mount"))
	}

	return nil
}
