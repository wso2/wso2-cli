package create

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/devops"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"gopkg.in/yaml.v3"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	commonCmd "github.com/wso2/integration-platform-tools/internal/cmd/common"
	common "github.com/wso2/integration-platform-tools/pkg/util/common"
	"github.com/wso2/integration-platform-tools/pkg/util/constants"
)

func getConfigKeys(componentType string, buildPackType string, hasConfigFile bool, repoDirPath string) []string {
	var buildConfigInputs []string
	// For API (service) integrations, prompt for port and visibility when no config file exists in the repo
	if componentType == component.ComponentTypeService && !hasConfigFile && repoDirPath != "" {
		buildConfigInputs = append(buildConfigInputs, PORT, VISIBILITY)
	}
	return buildConfigInputs
}

func promptBuildConfigurationInput(configList []string, params *CreateComponentParams, buildPack *devops.BuildPack) error {
	for _, item := range configList {
		switch MapBuildPackPromptTypes[item] {
		case "input":
			defaultVal := MapBuildPackDefaultConfigs[item]

			var inputOpts = prompt.PromptInputOpts{
				Message:  MapBuildPackPrompts[item],
				Default:  defaultVal,
				Validate: prompt.ValidateNotEmpty,
			}

			if fnV, hasValidation := MapBuildPackPromptValidation[item]; hasValidation {
				inputOpts.Validate = fnV
			}

			if MapBuildPackPromptOptional[item] {
				inputOpts.Message = fmt.Sprintf("%s (optional)", inputOpts.Message)
			}

			var input string
			err := prompt.NewPromptInputMessage(
				inputOpts,
				&input,
			).Prompt()
			if err != nil {
				return err
			}

			params.BuildPackConfigs[item] = input
		case "select":
			opts := MapBuildPackPromptOptions[item]

			var selected string
			err := prompt.NewPromptSelectMessage[string](
				prompt.PromptSelectOpts[string]{Message: MapBuildPackPrompts[item], Values: opts},
				&selected,
			).Prompt()

			if err != nil {
				return err
			}

			params.BuildPackConfigs[item] = selected
		}
	}

	return nil
}

func verifyConfigList(configList []string, params *CreateComponentParams) error {
	for _, item := range configList {
		val, ok := (params.BuildPackConfigs)[item]
		if !ok && !MapBuildPackPromptOptional[item] {
			return fmt.Errorf("'%s' is required for '%s' build configuration", item, params.BuildPack)
		}

		if fn, hasValidation := MapBuildPackPromptValidation[item]; hasValidation {
			return fn(val)
		}
	}

	return nil
}

func promptToSelectBuildPack(buildPacks []devops.BuildPack) (string, error) {
	var selectedBuildPackName string
	var buildPackNames []string

	for _, item := range buildPacks {
		buildPackNames = append(buildPackNames, item.Language)
	}

	err := prompt.NewPromptSelectMessage[string](
		prompt.PromptSelectOpts[string]{Message: "Build-pack:", Values: buildPackNames},
		&selectedBuildPackName,
	).Prompt()

	if err != nil {
		return "", err
	}

	return selectedBuildPackName, nil
}

func isSupportedBuildpack(bp devops.BuildPack) bool {
	lang := strings.ToLower(bp.Language)
	disp := strings.ToLower(bp.DisplayName)
	isBal := strings.Contains(lang, "ballerina") || strings.Contains(disp, "ballerina")
	isMI := strings.Contains(lang, "wso2") || strings.Contains(disp, "wso2") ||
		strings.Contains(lang, "micro integrator") || strings.Contains(disp, "micro integrator")
	return isBal || isMI
}

func resolveBuildPackInput(org *api.Organization, params *CreateComponentParams) (devops.BuildPack, error) {
	buildPackSpinner := utils.CreateSpinner(i18n.T(" Fetching build-packs..."), "")
	buildPackSpinner.Start()
	allBuildpacks, err := auth.DevopsClient.GetBuildPackOptions(org.UUID, org.ID, params.ComponentType)
	buildPackSpinner.Stop()
	if err != nil {
		return devops.BuildPack{}, err
	}

	var buildpacks []devops.BuildPack
	for _, bp := range allBuildpacks {
		if isSupportedBuildpack(bp) {
			buildpacks = append(buildpacks, bp)
		}
	}

	if len(buildpacks) == 0 {
		return devops.BuildPack{}, errors.New(i18n.T("no supported buildpacks found"))
	}

	// get the build pack input
	if params.BuildPack == "" {
		selectedBuildpackName, err := promptToSelectBuildPack(buildpacks)
		if err != nil {
			return devops.BuildPack{}, err
		}
		params.BuildPack = selectedBuildpackName

		for _, item := range buildpacks {
			if item.Language == selectedBuildpackName {
				return item, nil
			}
		}
	} else {
		for _, item := range buildpacks {
			if item.Language == params.BuildPack {
				return item, nil
			}
		}
	}

	return devops.BuildPack{}, errors.New(i18n.T("invalid build-pack selection"))
}

func ConfigFileExists(repoRoot string, params *CreateComponentParams) (bool, error) {
	configPath := filepath.Join(repoRoot, params.Subpath, constants.PLATFORM_META_DIR_NAME, constants.COMPONENT_YAML_NAME)
	if _, err := os.Stat(configPath); err == nil {
		return true, nil
	}

	configPath = filepath.Join(repoRoot, params.Subpath, constants.PLATFORM_META_DIR_NAME, constants.COMPONENT_CONFIG_YAML_NAME)
	if _, err := os.Stat(configPath); err == nil {
		return true, nil
	}

	configPath = filepath.Join(repoRoot, params.Subpath, constants.PLATFORM_META_DIR_NAME, constants.ENDPOINT_YAML_NAME)
	if _, err := os.Stat(configPath); err == nil {
		return true, nil
	}

	return false, nil
}

func CreateComponentConfigFile(componentPath string, componentType string, inbounds []models.ComponentYamlEndpoint, outbound *models.ComponentYamlConnectionReference) (string, error) {
	componentConfig := &models.ComponentYamlConfig{
		SchemaVersion: 1.2,
	}

	if componentType == component.ComponentTypeService && len(inbounds) > 0 {
		componentConfig.Endpoints = &inbounds
	}

	if outbound != nil {
		componentConfig.Dependencies = &models.ComponentYamlDependencies{
			ConnectionReferences: []models.ComponentYamlConnectionReference{},
		}
	}

	if fileStat, err := os.Stat(filepath.Join(componentPath, constants.PLATFORM_META_DIR_NAME)); os.IsNotExist(err) || !fileStat.IsDir() {
		err = os.MkdirAll(filepath.Join(componentPath, constants.PLATFORM_META_DIR_NAME), 0755)
		if err != nil {
			return "", err
		}
	}

	data, err := yaml.Marshal(componentConfig)
	if err != nil {
		return "", err
	}

	componentYmlFilePath := filepath.Join(componentPath, constants.PLATFORM_META_DIR_NAME, constants.COMPONENT_YAML_NAME)
	err = os.WriteFile(componentYmlFilePath, data, 0644)
	if err != nil {
		return "", err
	}

	return componentYmlFilePath, nil
}

func CreateComponentYamlFile(componentPath string, componentType string, model models.ComponentYamlConfig) (string, error) {
	dirPath := filepath.Join(componentPath, constants.PLATFORM_META_DIR_NAME)
	filepath := filepath.Join(dirPath, constants.COMPONENT_YAML_NAME)
	if dirStat, err := os.Stat(dirPath); os.IsNotExist(err) || !dirStat.IsDir() {
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			return "", err
		}
	}

	if file, err := os.Open(filepath); os.IsNotExist(err) {
		data, err := yaml.Marshal(model)
		if err != nil {
			return "", err
		}

		err = os.WriteFile(filepath, data, 0644)
		if err != nil {
			return "", err
		}
	} else {
		existingData, err := io.ReadAll(file)
		if err != nil {
			return "", err
		}

		var currD models.ComponentYamlConfig
		if err = yaml.Unmarshal(existingData, &currD); err != nil {
			return "", err
		}

		if model.Proxy != nil {
			currD.Proxy = model.Proxy
		}

		updatedData, err := yaml.Marshal(currD)
		if err != nil {
			return "", err
		}

		err = os.WriteFile(filepath, updatedData, 0644)
		if err != nil {
			return "", err
		}
	}

	return filepath, nil
}

func genComponentConfigFile(params *CreateComponentParams, repoPath string, fileExists bool) error {
	if params.ComponentType == component.ComponentTypeService {
		if repoPath != "" && !fileExists {
			// Create component.yaml only if user is within the repo and the file already does not exist
			componentDir := filepath.Join(repoPath, params.Subpath)

			if _, err := os.Stat(componentDir); !os.IsNotExist(err) {
				port, err := strconv.Atoi(params.BuildPackConfigs[PORT])
				if err != nil {
					return err
				}

				inboundConfigs := []models.ComponentYamlEndpoint{{
					Name:                fmt.Sprintf("%s-endpoint", params.ComponentName),
					Service:             models.ComponentYamlService{Port: port},
					Type:                "REST",
					NetworkVisibilities: []string{params.BuildPackConfigs[VISIBILITY]},
				}}

				componentYmlFilePath, err := CreateComponentConfigFile(componentDir, params.ComponentType, inboundConfigs, nil)
				if err != nil {
					return err
				}

				relPath, err := filepath.Rel(repoPath, componentYmlFilePath)

				if err != nil {
					return err
				}

				fmt.Fprintln(
					utils.IO.Out,
					fmt.Sprintf(
						i18n.T("\nComponent configurations created at %s\n"),
						utils.CS.Bold(relPath),
					),
				)

				fmt.Fprintln(
					utils.IO.Out,
					utils.CS.Yellow(
						i18n.T("Please commit and push it to your remote repository in order to expose the connection configurations"+
							" of your service."),
					),
				)

				fmt.Fprintln(
					utils.IO.Out,
					utils.CS.Yellow(
						fmt.Sprintf(
							i18n.T("More information about configuring connections can be found at:\n%s"),
							"https://wso2.com/choreo/docs/develop-components/configure-endpoints/#apply-advanced-component-connection-configurations",
						),
					),
				)
			}
		} else {
			fmt.Fprintln(
				utils.IO.Out,
				utils.CS.Yellowf(
					i18n.T("\nFor more information on configuring inbound & outbound connections:\n%s"),
					"https://wso2.com/choreo/docs/develop-components/configure-endpoints/#apply-advanced-component-connection-configurations",
				),
			)
		}
	}

	return nil
}

func handleCreateComponent(
	params *CreateComponentParams,
	project models.Project,
	orgId,
	orgUUID,
	remoteUrl string) error {

	componentReqData, err := GetComponentKindForCreate(params, project.Handler, orgId, orgUUID, remoteUrl)
	if err != nil {
		return err
	}

	createCmpSpinner := utils.CreateSpinner(i18n.T(" Creating component..."), "")
	createCmpSpinner.Start()
	_, err = auth.ComponentClient.CreateNewComponent(orgId, project.Handler, *componentReqData)
	createCmpSpinner.Stop()
	if err != nil {
		return err
	}
	return nil
}

func GetComponentKindForCreate(
	params *CreateComponentParams,
	projectHandle string,
	orgId,
	orgUUID string,
	remoteUrl string) (*component.ComponentKind, error) {
	displayType := component.GetDisplayNameForComponentType(params.ComponentType, params.BuildPack)
	displayName := params.ComponentName
	if params.DisplayName != "" {
		displayName = params.DisplayName
	}
	sanitizedComponentName := strings.ToLower(params.ComponentName)
	regex, err := regexp.Compile(`[^a-z0-9-]`)
	if err != nil {
		return nil, fmt.Errorf("component name validation failed: %w", err)
	}
	sanitizedComponentName = regex.ReplaceAllString(
		strings.ReplaceAll(sanitizedComponentName, " ", "-"),
		"",
	)
	if len(sanitizedComponentName) > 25 {
		sanitizedComponentName = sanitizedComponentName[:25]
	}

	sourceKind := component.ComponentKindSource{}

	gitProvider := component.GitProvider{
		Repository:           remoteUrl,
		Branch:               params.RepoBranch,
		Path:                 params.Subpath,
		IsPublicRepo:         params.IsPublicRepo,
		PullLatestSubmodules: true,
	}

	if params.RepoProvider == GIT_LAB_SERVER {
		sourceKind.Gitlab = &gitProvider
		sourceKind.SecretRef = params.GitCredRef
	} else if params.RepoProvider == BIT_BUCKET {
		sourceKind.Bitbucket = &gitProvider
		sourceKind.SecretRef = params.GitCredRef
	} else {
		sourceKind.Github = &gitProvider
	}

	subType := params.ComponentSubType
	if subType == "fileIntegration" {
		if params.BuildPack == component.ComponentBuildPackBallerina && displayType == component.DisplayTypeBallerinaEventHandler {
			subType = "ballerinaFileIntegration"
		} else if params.BuildPack == component.ComponentBuildPackMI && displayType == component.DisplayTypeMiEventHandler {
			subType = "miFileIntegration"
		}
	}

	if params.OriginCloud == "" {
		params.OriginCloud = "devant"
	}

	componentReqData := component.ComponentKind{
		ApiVersion: "core.choreo.dev/v1alpha1",
		Kind:       "Component",
		Metadata: component.ComponentKindMetadata{
			Name:        sanitizedComponentName,
			DisplayName: displayName,
			ProjectName: projectHandle,
		},
		Spec: component.ComponentKindSpec{
			Type:    displayType,
			SubType: subType,
			Source:  sourceKind,
		},
		OriginCloud: &params.OriginCloud,
	}

	if params.Description != "" {
		componentReqData.Metadata.Description = params.Description
	}

	if params.BuildPack == component.ComponentBuildPackBallerina {
		componentReqData.Spec.Build = component.ComponentKindSpecBuild{
			Ballerina: &component.ComponentKindBuildBallerina{
				EnableCellDiagram: true,
				IsUnitTestEnabled: true,
			},
		}
	}
	// MI: no additional build configuration needed

	componentReqData.Spec.Build.EnableAutoBuild = true
	if params.AutoBuild != nil {
		componentReqData.Spec.Build.EnableAutoBuild = *params.AutoBuild
	}
	componentReqData.Spec.Build.EnableAutoDeploy = true
	if params.AutoDeploy != nil {
		componentReqData.Spec.Build.EnableAutoDeploy = *params.AutoDeploy
	}

	return &componentReqData, nil
}

func genDirList(dirStruct []component.PathEntry, finalList *[]string) {
	for _, item := range dirStruct {
		if item.Type == component.PathTreeType {
			if strings.HasPrefix(item.SubPath, ".") {
				continue
			}

			*finalList = append(*finalList, item.Path)
		}

		genDirList(item.Children, finalList)
	}
}

func validateComponentName(componentName string, existingComps []models.Component) error {
	err := prompt.ValidateNotEmpty(componentName)
	if err != nil {
		return err
	}

	err = commonCmd.ValidateNameText(componentName)
	if err != nil {
		return err
	}

	for _, item := range existingComps {
		if item.Name == componentName {
			return fmt.Errorf("component name already exists")
		}
	}

	return nil
}

func promptComponentName(opts *CreateComponentParams) error {
	if opts.ComponentName == "" {
		defaultCompName := ""
		if !common.StringExistsInSlice(filepath.Base(opts.Subpath), []string{".", "./", "/"}) {
			defaultCompName = filepath.Base(opts.Subpath)
		}

		err := prompt.NewPromptInputMessage(
			prompt.PromptInputOpts{
				Message: i18n.T("Display Name:"),
				Default: defaultCompName,
				Validate: func(s string) error {
					input := strings.Trim(s, " ")

					if err := validateComponentName(input, []models.Component{}); err != nil {
						return err
					}
					return nil
				},
			},
			&opts.ComponentName,
		).Prompt()

		if err != nil {
			return err
		}
	}

	return nil
}
