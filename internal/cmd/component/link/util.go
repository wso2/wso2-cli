package link

import (
	"os"
	"path/filepath"

	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/pkg/util/constants"
	"gopkg.in/yaml.v3"
)

func CreateComponentLink(componentDir string, projectHandle string, orgHandle string, componentHandle string) error {
	projectLinkData := common.LinkData{
		Project:   projectHandle,
		Org:       orgHandle,
		Component: componentHandle,
	}

	data, err := yaml.Marshal(&projectLinkData)
	if err != nil {
		return err
	}

	if fileStat, err := os.Stat(filepath.Join(componentDir, constants.PROJECT_META_DIR_NAME)); os.IsNotExist(err) || !fileStat.IsDir() {
		err = os.MkdirAll(filepath.Join(componentDir, constants.PROJECT_META_DIR_NAME), 0755)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(filepath.Join(componentDir, constants.PROJECT_META_DIR_NAME, constants.COMPONENT_LINK_FILE), data, 0644)
	if err != nil {
		return err
	}

	return nil
}
