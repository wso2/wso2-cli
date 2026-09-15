package change

import (
	"fmt"

	"github.com/charmbracelet/huh"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api"
)

func promptToSelectOrg(orgs []api.Organization) (*api.Organization, error) {

	selectedOrg, err := auth.GetSelectedOrganization()
	if err != nil {
		return nil, fmt.Errorf(i18n.T("failed to get active organization: %w"), err)
	}

	orgNames := make([]string, len(orgs))
	for i, org := range orgs {
		if selectedOrg != nil && selectedOrg.ID == org.ID {
			orgNames[i] = fmt.Sprintf("%s	%s %s", org.Name, utils.CS.Gray(org.Handle), "(active)")
		} else {
			orgNames[i] = fmt.Sprintf("%s	%s", org.Name, utils.CS.Gray(org.Handle))
		}
	}

	var orgIndex int
	opts := make([]huh.Option[int], len(orgs))

	for i := range orgs {
		opts[i] = huh.NewOption(orgNames[i], i)
	}

	err = prompt.NewPromptSelectMessage(
		prompt.PromptSelectOpts[int]{
			Message: i18n.T("Choose an org:"),
			Options: opts,
		},
		&orgIndex,
	).Prompt()

	if err != nil {
		return nil, err
	}
	if (orgIndex < 0) || (orgIndex >= len(orgs)) {
		return nil, fmt.Errorf("%s", i18n.T("invalid organization index"))
	}
	return &orgs[orgIndex], nil
}
