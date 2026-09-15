package create

import (
	"fmt"
	"regexp"
	"strings"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api/devops"
)

func validateConfigName(configName string, configs []devops.ConfigItem) error {
	if len(configName) > 50 {
		return fmt.Errorf("config name cannot have more than 50 characters")
	}

	regex, err := regexp.Compile("^[a-z]([-a-z0-9]*[a-z0-9])?$")
	if err != nil {
		return err
	}

	if !regex.MatchString(configName) {
		return fmt.Errorf("config name must consist of only lowercase alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")
	}

	for _, item := range configs {
		if item.Name == configName {
			return fmt.Errorf("config name already exists")
		}
	}

	return nil
}

func resolveConfigName(configName *string, configs []devops.ConfigItem) error {
	if *configName == "" {
		err := prompt.NewPromptInputMessage(
			prompt.PromptInputOpts{
				Message: i18n.T("Config name:"),
				Validate: func(configName string) error {
					if valerr := prompt.ValidateNotEmpty(configName); valerr != nil {
						return valerr
					}

					if valerr := validateConfigName(configName, configs); valerr != nil {
						return valerr
					}

					return nil
				},
			},
			configName,
		).Prompt()

		if err != nil {
			return fmt.Errorf("%s", i18n.T(" failed to get a valid config name"))
		}
	} else {
		return validateConfigName(*configName, configs)
	}
	return nil
}

func validateConfigMountPath(pathStr string) error {
	if !strings.HasPrefix(pathStr, "/") {
		return fmt.Errorf("config mount path must start with '/'")
	}
	return nil
}

func resolveMountType(mountType *string) error {
	if *mountType == "" || !strings.Contains(strings.Join(common.MountTypes, ","), *mountType) {
		err := prompt.NewPromptSelectMessage[string](
			prompt.PromptSelectOpts[string]{Message: i18n.T("Mount Type:"), Values: common.MountTypes},
			mountType,
		).Prompt()

		if err != nil {
			return err
		}
	}

	return nil
}

func resolveConfigType(configType *string) error {
	if *configType == "" || !strings.Contains(strings.Join(common.ConfigTypes, ","), *configType) {
		err := prompt.NewPromptSelectMessage[string](
			prompt.PromptSelectOpts[string]{Message: i18n.T("Config type:"), Values: common.ConfigTypes},
			configType,
		).Prompt()

		if err != nil {
			return err
		}
	}

	return nil
}

func resolveConfigMountPath(configMountPath *string) error {
	if *configMountPath == "" {
		err := prompt.NewPromptInputMessage(
			prompt.PromptInputOpts{
				Message:     i18n.T("Config mount path:"),
				Description: i18n.T("The path where the file will be mounted within the container"),
				Validate: func(configPath string) error {
					if valerr := prompt.ValidateNotEmpty(configPath); valerr != nil {
						return valerr
					}

					if valerr := validateConfigMountPath(configPath); valerr != nil {
						return valerr
					}

					return nil
				},
			},
			configMountPath,
		).Prompt()

		if err != nil {
			return fmt.Errorf("%s", i18n.T(" failed to get a valid config mount path"))
		}
	} else {
		return validateConfigMountPath(*configMountPath)
	}

	return nil
}

func resolveConfigMountContent(configMountContent *string) error {
	if *configMountContent == "" {
		// err := utils.PromptMultiLineInput[string](
		// 	&survey.Multiline{Message: i18n.T("Config content:")},
		// 	configMountContent,
		// 	survey.WithValidator(survey.Required),
		// )

		err := prompt.NewPromptTextInputMessage(
			prompt.PromptTextInputOpts{
				Message:  i18n.T("Config content:"),
				Validate: prompt.ValidateNotEmpty,
			},
			configMountContent,
		)
		if err != nil {
			return fmt.Errorf("%s", i18n.T(" failed to get a valid config content"))
		}
	}
	return nil
}

func resolveEnvValues(envVars []common.KeyValOpt) ([]common.KeyValOpt, error) {
	if len(envVars) == 0 {
		for {
			var envVar common.KeyValOpt

			err := prompt.NewPromptInputMessage(
				prompt.PromptInputOpts{
					Message:  i18n.T("Env variable name:"),
					Validate: prompt.ValidateNotEmpty,
				},
				&envVar.Key,
			).Prompt()

			if err != nil {
				return nil, fmt.Errorf("%s", i18n.T(" error prompting for env name"))
			}

			err = prompt.NewPromptInputMessage(
				prompt.PromptInputOpts{
					Message:  i18n.T("Env variable value:"),
					Validate: prompt.ValidateNotEmpty,
				},
				&envVar.Val,
			).Prompt()

			if err != nil {
				return nil, fmt.Errorf("%s", i18n.T(" error prompting for env value"))
			}

			envVars = append(envVars, envVar)

			var anotherVar bool

			err = prompt.NewPromptConfirmMessage(
				prompt.PromptConfirmOpts{
					Title: i18n.T("Do you want to add another environment variable?"),
				},
				&anotherVar,
			).Prompt()

			if err != nil {
				return nil, fmt.Errorf("%s", i18n.T(" error prompting for anotherVar"))
			}

			if !anotherVar {
				break
			}
		}
	}

	if len(envVars) == 0 {
		return nil, fmt.Errorf("%s", i18n.T("need to provide at least on pair of env variable name and value"))
	}

	return envVars, nil
}
