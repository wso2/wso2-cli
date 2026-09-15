package auth

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

func handleLogout(cmd *cobra.Command, args []string) error {

	// Logout
	logoutSpinner := utils.CreateSpinner(" Logging out...", "")
	logoutSpinner.Start()
	err := auth.SignOut()
	if err != nil {
		logoutSpinner.Stop()
		return fmt.Errorf("failed to logout, %w", err)
	}
	logoutSpinner.Stop()

	fmt.Fprint(utils.IO.Out, heredoc.Docf(`
		You have been logged out.

		To log back in :
			%s
	`, "$ wso2-integration-platform login"))
	return nil
}

var LogoutCmd = &cobra.Command{
	Use:     "logout [flags]",
	Aliases: []string{"signout"},
	Short:   "logout from the Integration Platform",
	Long:    "Log out of your Integration Platform account, ending your current session and ensuring secure access control. ",
	Run: func(cmd *cobra.Command, args []string) {
		err := handleLogout(cmd, args)
		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddGenericHelper(LogoutCmd)
}
