package list

import (
	"fmt"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api/connections"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

// printConnectionListStructured renders the connection list in a
// machine-readable format (json) and writes it to stdout. When the list is
// empty, a human-readable hint is written to stderr instead of stdout, so a
// redirected file or a pipe (e.g. `-o json > connections.json` or `| jq`)
// still receives only valid JSON.
func printConnectionListStructured(conList []connections.Connection, format common.OutputFormat, currentProject models.Project) error {
	// Normalize a nil slice to an empty slice so JSON renders `[]` rather
	// than `null` when there are no connections.
	if conList == nil {
		conList = []connections.Connection{}
	}

	rendered, err := common.RenderStructured(format, conList)
	if err != nil {
		return err
	}

	// The JSON payload always goes to stdout — even when it is just `[]`.
	fmt.Fprintln(utils.IO.Out, rendered)

	// On an empty result, guide the user via stderr (kept out of stdout so
	// the JSON stays machine-parseable).
	if len(conList) == 0 {
		fmt.Fprintf(utils.IO.ErrOut, i18n.T("No connections found in project %s.\n"),
			utils.CS.Bold(currentProject.Name))
	}
	return nil
}

func printConnectionList(conList []connections.Connection) error {
	if len(conList) == 0 {
		fmt.Fprintf(utils.IO.ErrOut, i18n.T("%s No connections found\n"), utils.CS.Yellow("!"))
		return nil
	} else {
		data := [][]string{}
		fmt.Println()

		for _, conItem := range conList {
			data = append(data, []string{conItem.Name, conItem.ServiceName})
		}

		table := utils.CreateTable(data, []string{i18n.T("NAME"), i18n.T("CONNECTION TO")}, "")
		fmt.Println(table)
	}
	return nil
}
