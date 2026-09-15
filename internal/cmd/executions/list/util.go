package list

import (
	"fmt"
	"strconv"
	"time"

	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/textstyle"
	"github.com/wso2/integration-platform-tools/pkg/api/logs"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

// printExecutionsStructured renders the execution list in a machine-readable
// format (json) and writes it to stdout. When the list is empty, a
// human-readable hint is written to stderr instead of stdout, so a redirected
// file or a pipe (e.g. `-o json > executions.json` or `| jq`) still receives
// only valid JSON.
func printExecutionsStructured(executions []logs.ExecutionListItemV2, format common.OutputFormat, sComponent models.Component, sProject models.Project) error {
	// Normalize a nil slice to an empty slice so JSON renders `[]` rather
	// than `null` when there are no executions.
	if executions == nil {
		executions = []logs.ExecutionListItemV2{}
	}

	rendered, err := common.RenderStructured(format, executions)
	if err != nil {
		return err
	}

	// The JSON payload always goes to stdout — even when it is just `[]`.
	fmt.Fprintln(utils.IO.Out, rendered)

	// On an empty result, guide the user via stderr (kept out of stdout so
	// the JSON stays machine-parseable).
	if len(executions) == 0 {
		fmt.Fprintf(utils.IO.ErrOut, i18n.T("No executions found for the component %s of %s project.\n"),
			utils.CS.Bold(sComponent.Name),
			utils.CS.Bold(sProject.Name))
	}
	return nil
}

func printExecutionsTable(executions []logs.ExecutionListItemV2, sComponent models.Component, sProject models.Project) {
	dataTable := getDataTableContent(executions)
	boldStyle := textstyle.CreateTextStyle().AddTextStyle(textstyle.BOLD)

	table := utils.CreateTable(
		dataTable,
		[]string{
			"RunId",
			"Revision",
			"Status",
			"Duration",
			"Started at",
		},
		" ",
	)
	utils.PrintInfo(
		"\nList of executions for the component %s of %s\n\n",
		boldStyle.Text(sComponent.Name),
		boldStyle.Text(sProject.Name),
	)

	utils.PrintInfo("%s", table)
}

func getDataTableContent(executions []logs.ExecutionListItemV2) [][]string {
	var dataTable = make([][]string, 0)

	for _, exec := range executions {
		row := make([]string, 0)
		status := "in-progress"
		duration := "-"

		stVal, err := strconv.Atoi(exec.StartTime)

		if err != nil {
			fmt.Printf("Error parsing StartTime for execution %s: %v\n", exec.ID, err)
			stVal = 0
		}

		if exec.CompletionTime != "" {
			ctVal, err := strconv.Atoi(exec.CompletionTime)

			if err != nil {
				fmt.Printf("Error parsing StartTime for execution %s: %v\n", exec.ID, err)
			} else {
				duration = time.Unix(int64(ctVal), 0).Sub(time.Unix(int64(stVal), 0)).String()
			}
		}

		switch exec.Status {
		case "Failed":
			status = "failure"
		case "Succeeded":
			status = "success"
		default:
			status = "in-progress"
		}

		row = append(row, exec.ID)
		row = append(row, shortRevision(exec.RevisionID))
		row = append(row, getRunStatus(status))
		row = append(row, duration)
		row = append(row, utils.TimeAgo(time.Unix(int64(stVal), 0)))

		dataTable = append(dataTable, row)
	}

	return dataTable
}

// shortRevision returns the first 8 characters of a revision string, or the
// full string if it is shorter. Some executions report short or empty
// RevisionIDs, so a naive `s[:8]` slice would panic.
func shortRevision(revisionID string) string {
	if len(revisionID) < 8 {
		return revisionID
	}
	return revisionID[:8]
}

func getRunStatus(status string) string {
	successT := textstyle.CreateTextStyle().SetTextColor(textstyle.GREEN)
	failureT := textstyle.CreateTextStyle().SetTextColor(textstyle.RED)
	midT := textstyle.CreateTextStyle().SetTextColor(textstyle.BLUE)

	switch status {
	case "success":
		return successT.Text(status)
	case "failure":
		return failureT.Text(status)
	default:
		return midT.Text(status)
	}
}
