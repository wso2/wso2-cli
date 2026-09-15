package list

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	deploymentbuild "github.com/wso2/integration-platform-tools/pkg/api/deploymentBuild"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

// printBuildListStructured renders the build list in a machine-readable
// format (json) and writes it to stdout. When the list is empty, a
// human-readable hint is written to stderr instead of stdout, so a redirected
// file or a pipe (e.g. `-o json > builds.json` or `| jq`) still receives only
// valid JSON.
func printBuildListStructured(builds []deploymentbuild.BuildKind, format common.OutputFormat, componentInfo models.Component, currentProject models.Project) error {
	// Normalize a nil slice to an empty slice so JSON renders `[]` rather
	// than `null` when there are no builds.
	if builds == nil {
		builds = []deploymentbuild.BuildKind{}
	}

	rendered, err := common.RenderStructured(format, builds)
	if err != nil {
		return err
	}

	// The JSON payload always goes to stdout — even when it is just `[]`.
	fmt.Fprintln(utils.IO.Out, rendered)

	// On an empty result, guide the user via stderr (kept out of stdout so
	// the JSON stays machine-parseable).
	if len(builds) == 0 {
		fmt.Fprintf(utils.IO.ErrOut, i18n.T("No builds found for the component %s of %s project.\n"),
			utils.CS.Bold(componentInfo.Name),
			utils.CS.Bold(currentProject.Name))
	}
	return nil
}

func printBuildListTable(builds []deploymentbuild.BuildKind, componentInfo models.Component, currentProject models.Project) {
	if len(builds) == 0 {
		fmt.Fprintln(
			utils.IO.Out,
			fmt.Sprintf(
				i18n.T(`No builds found for the component %s of %s project`),
				utils.CS.Bold(componentInfo.Name),
				utils.CS.Bold(currentProject.Name),
			),
		)
		return
	}

	fmt.Fprintln(
		utils.IO.Out,
		fmt.Sprintf(
			i18n.T(heredoc.Doc(`
                        List of available builds of the component %s of %s project:
                    `)),
			utils.CS.Bold(componentInfo.Name),
			utils.CS.Bold(currentProject.Name),
		),
	)

	tableRows := make([][]string, 0)
	for _, build := range builds {
		tableRows = append(tableRows, []string{
			fmt.Sprintf("%d", build.Status.RunID),
			common.GetBuildStatusStr(build),
			fmt.Sprintf("%s ago", utils.GetRelativeTime(build.Status.StartedAt)),
			shortCommitId(build.Spec.Revision),
			build.Status.GitCommit.Message,
		})
	}

	fmt.Fprintln(
		utils.IO.Out,
		utils.CreateTable(tableRows, []string{"Build ID", "Status", "Started", "Commit ID", "Commit Message"}, ""),
	)
}

func shortCommitId(commitId string) string {
	return commitId[:8]
}
