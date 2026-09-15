package list

import (
	"fmt"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/devops"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

// configWithEnv pairs a config item with the human-readable name of the
// environment it belongs to. Configs are listed across one or more
// environments, so the env name is carried alongside each item for display
// (table ENV column) and structured output (the `env` field).
type configWithEnv struct {
	devops.ConfigItem
	EnvName string `json:"env"`
}

// printConfigListStructured renders the config list in a machine-readable
// format (json) and writes it to stdout. When the list is empty, a
// human-readable hint is written to stderr instead of stdout, so a redirected
// file or a pipe (e.g. `-o json > configs.json` or `| jq`) still receives only
// valid JSON.
func printConfigListStructured(configs []configWithEnv, format common.OutputFormat, sComponent models.Component, sProject models.Project) error {
	// Normalize a nil slice to an empty slice so JSON renders `[]` rather
	// than `null` when there are no configs.
	if configs == nil {
		configs = []configWithEnv{}
	}

	rendered, err := common.RenderStructured(format, configs)
	if err != nil {
		return err
	}

	// The JSON payload always goes to stdout — even when it is just `[]`.
	fmt.Fprintln(utils.IO.Out, rendered)

	// On an empty result, guide the user via stderr (kept out of stdout so
	// the JSON stays machine-parseable).
	if len(configs) == 0 {
		fmt.Fprintf(utils.IO.ErrOut, i18n.T("No config-maps or secrets found for the component %s of %s project.\n"),
			utils.CS.Bold(sComponent.Name),
			utils.CS.Bold(sProject.Name))
	}
	return nil
}

func printConfigList(configs []configWithEnv) error {
	if len(configs) == 0 {
		fmt.Fprintf(utils.IO.Out, i18n.T("%s No config-maps or secrets found\n"), utils.CS.Yellow("!"))
		return nil
	} else {
		data := [][]string{}
		fmt.Println()

		for _, configItem := range configs {
			configType := "Config-map"
			if configItem.SecretType != "" {
				configType = "Secret"
			}

			mountType := "Env variables"
			if configItem.ConfigType == "File" {
				mountType = "File"
			}

			data = append(data, []string{configItem.EnvName, configItem.Name, configType, mountType, utils.TimeAgo(configItem.CreatedAt), utils.TimeAgo(configItem.UpdatedAt)})
		}

		table := utils.CreateTable(data, []string{i18n.T("ENV"), i18n.T("NAME"), i18n.T("TYPE"), i18n.T("MOUNT DETAILS"), i18n.T("CREATED"), i18n.T("LAST UPDATED")}, "")
		fmt.Println(table)
	}
	return nil
}
