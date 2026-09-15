package common

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/util/constants"
	"gopkg.in/yaml.v3"
)

const OrgFlagName = "org"

func VerifyIsUserLoggedIn(cmd *cobra.Command, args []string) {
	authVerifySpinner := utils.CreateSpinner(i18n.T(" Verifying user..."), "")
	authVerifySpinner.Start()
	if loggedIn := auth.IsLoggedIn(); !loggedIn {
		authVerifySpinner.Stop()
		utils.HandleErr(api.ErrNotLoggedIn)
	}
	authVerifySpinner.Stop()
}

func ResolveContext(orgFlag, projectFlag string) (orgId string, projectId string, err error) {
	if orgFlag != "" || projectFlag != "" {
		err = fmt.Errorf("skipping context check")
		return
	}

	type CtxModel struct {
		Project string `yaml:"project"`
		Org     string `yaml:"org"`
	}

	repoRootPath, _, err := GetRepoRootAndSubPath()
	if err != nil {
		return
	}

	ctxDPath := filepath.Join(repoRootPath, constants.CTX_DIR)
	ctxFPath := filepath.Join(ctxDPath, constants.CTX_FILE)

	if _, err = os.Stat(ctxFPath); errors.Is(err, os.ErrNotExist) {
		return
	}

	file, err := os.Open(ctxFPath)
	if err != nil {
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return
	}

	var ctxs []CtxModel
	if err = yaml.Unmarshal(content, &ctxs); err != nil {
		return
	}

	if len(ctxs) > 1 {
		fmt.Fprintln(utils.IO.Out, i18n.T("Detected multiple registered contexts"))
		opts := make([]huh.Option[int], 0)

		for i, val := range ctxs {
			opts = append(opts, huh.NewOption(fmt.Sprintf("%s (%s)", val.Project, val.Org), i))
		}

		var value int

		err = prompt.NewPromptSelectMessage(prompt.PromptSelectOpts[int]{
			Message: i18n.T("Select the context you wish to use"),
			Default: opts[0].Value,
			Options: opts,
		}, &value).Prompt()

		if err != nil {
			return
		}

		orgId = ctxs[value].Org
		projectId = ctxs[value].Project

		return
	} else if len(ctxs) == 1 {
		orgId = ctxs[0].Org
		projectId = ctxs[0].Project
	}

	if orgId == "" || projectId == "" {
		err = errors.New("Empty value detected for context")
		return
	}

	return
}

func ResolveTargetOrganization(orgFlagValue string) (*api.Organization, error) {
	var targetOrg *api.Organization

	orgSelectSpinner := utils.CreateSpinner(i18n.T(" Fetching organization information..."), "")
	orgSelectSpinner.Start()
	selectedOrg, err := auth.GetSelectedOrganization()
	if err != nil {
		return nil, err
	}

	if orgFlagValue == "" {
		orgSelectSpinner.Stop()

		targetOrg = selectedOrg
	} else {
		orgs, err := auth.OrgClient.GetOrganizations()
		orgSelectSpinner.Stop()

		if err != nil {
			return nil, err
		}

		found := false
		for _, org := range orgs {
			if org.Handle == orgFlagValue || org.Name == orgFlagValue || org.ID == orgFlagValue {
				targetOrg = &org
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf(i18n.T(" organization with handle %s not found"), orgFlagValue)
		}
	}

	checkAccessSpinner := utils.CreateSpinner(i18n.T("Checking access to organization..."), "")
	checkAccessSpinner.Start()
	hasAccess, err := auth.HasAccessToOrg(targetOrg.ID)
	checkAccessSpinner.Stop()

	if err != nil {
		return nil, err
	} else if !hasAccess {
		return nil, fmt.Errorf(i18n.T("you do not have access to organization %s"), targetOrg.Handle)
	}
	return targetOrg, nil
}

func FetchUserOrgs() (orgList []api.Organization, selectedOrg *api.Organization, err error) {
	loadOrgsSpinner := utils.CreateSpinner(i18n.T(" Fetching organizations..."), "")
	loadOrgsSpinner.Start()

	selectedOrg, err = auth.GetSelectedOrganization()
	if err != nil {
		loadOrgsSpinner.Stop()
		return nil, nil, fmt.Errorf(i18n.T("failed to get active organization: %w"), err)
	}

	orgs, err := auth.OrgClient.GetOrganizations()
	if err != nil {
		loadOrgsSpinner.Stop()
		return nil, nil, fmt.Errorf(i18n.T("failed to get organizations: %w"), err)
	}
	loadOrgsSpinner.Stop()

	return orgs, selectedOrg, nil
}

func AddOrgFlag(cmdFlags *pflag.FlagSet, bindTo *string) {
	// No shorthand: -o is reserved for the --output flag
	cmdFlags.StringVar(bindTo, OrgFlagName, "", i18n.T("organization name, ID, or handle"))
}

func AddGitCredentialFlag(cmdFlags *pflag.FlagSet, bindTo *string) {
	cmdFlags.StringVarP(bindTo, "git-credential", "", "", i18n.T("Git credential name"))
}

func AddDescriptionFlag(cmdFlags *pflag.FlagSet, bindTo *string) {
	cmdFlags.StringVarP(bindTo, "description", "", "", i18n.T("Project description."))
}
