package set

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/util/constants"
	"gopkg.in/yaml.v3"
)

var ctxSetOpts CtxSetOpts

var SetCtxCmd = &cobra.Command{
	Use:   "set-context [flags]",
	Short: i18n.T("set context for the Integration Platform CLI to be used within the repository"),
	Long: i18n.T(heredoc.Doc(`
        Associate the repository directory with a project in WSO2 Integration Platform and use it as a context for subsequent commands
    `)),
	PreRun: common.VerifyIsUserLoggedIn,
	Run: func(cmd *cobra.Command, args []string) {
		ctxSetOpts.ShowMessages = true
		err := HandleSetCtx(ctxSetOpts)
		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func HandleSetCtx(params CtxSetOpts) error {
	repoRootPath, _, err := common.GetRepoRootAndSubPath()
	if err != nil {
		return err
	}

	ctxDPath := filepath.Join(repoRootPath, constants.CTX_DIR)
	ctxFPath := filepath.Join(ctxDPath, constants.CTX_FILE)

	if _, err = os.Stat(ctxFPath); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(ctxDPath, 0755)
		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(ctxFPath, os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var ctxs []CtxModel

	if err := yaml.Unmarshal(content, &ctxs); err != nil {
		return err
	}

	so, err := common.ResolveTargetOrganization(params.Org)
	if err != nil {
		return err
	}

	if params.ShowMessages {
		fmt.Fprintln(utils.IO.Out, fmt.Sprintf(i18n.T("Resolved Org: %s(%s)"), so.Name, so.Handle))
	}

	prj, err := common.ResolveTargetProject(so, params.Project)
	if err != nil {
		return err
	}

	if params.ShowMessages {
		fmt.Fprintln(utils.IO.Out, fmt.Sprintf(i18n.T("Resolved Project: %s(%s)"), prj.Name, prj.Handler))
	}

	found := false

	for _, entry := range ctxs {
		if entry.Org == so.Handle && entry.Project == prj.Handler {
			found = true
			break
		}
	}

	if found {
		if params.ShowMessages {
			fmt.Fprintln(utils.IO.Out, i18n.T("Context already registered in context.yaml"))
		}
	} else {
		ctxs = append(ctxs, CtxModel{Org: so.Handle, Project: prj.Handler})
		outBytes, err := yaml.Marshal(ctxs)
		if err != nil {
			return err
		}
		os.WriteFile(ctxFPath, outBytes, 0644)
		if params.ShowMessages {
			fmt.Fprintln(utils.IO.Out, i18n.T("Context was set successfully"))
		}
	}

	return nil
}

func init() {
	common.AddOrgFlag(SetCtxCmd.Flags(), &ctxSetOpts.Org)
	common.AddProjectFlag(SetCtxCmd.Flags(), &ctxSetOpts.Project)
}
