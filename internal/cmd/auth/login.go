package auth

import (
	"bufio"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
)

var sio = &SignInOpt{}

type LoginMode int

const (
	ByBrowser LoginMode = iota
	ByToken
)

func handleLogin(cmd *cobra.Command, args []string) error {
	// display login message and open browser
	if !sio.withToken {
		loginOpts := []huh.Option[LoginMode]{
			huh.NewOption("Login with a web browser", ByBrowser),
			huh.NewOption("Login with a Token", ByToken),
		}

		var loginMode LoginMode

		err := prompt.NewPromptSelectMessage[LoginMode](
			prompt.PromptSelectOpts[LoginMode]{
				Message: "How would you like to login?",
				Options: loginOpts,
			},
			&loginMode,
		).Prompt()

		if err != nil {
			return err
		}

		switch loginMode {
		case ByBrowser:
			_, _, err = auth.SignIn(nil)
			if err != nil {
				return fmt.Errorf("Failed to authenticate: %w", err)
			}

			auth.SetInitialRegionOfOrg()

			fmt.Fprint(utils.IO.Out, heredoc.Docf(i18n.T(`
                To change the current organization :
                    %s
            `),
				"$ wso2-integration-platform change-org",
			))
		case ByToken:
			var tInput string

			err = prompt.NewPromptInputMessage(
				prompt.PromptInputOpts{
					Message:  i18n.T("Paste your authentication token:"),
					Default:  "",
					Validate: prompt.ValidateNotEmpty,
					IsSecure: true,
				},
				&tInput,
			).Prompt()

			if err != nil {
				return err
			}

			authCodeSpinner := utils.CreateSpinner(i18n.T(" Verifying token..."), "")
			authCodeSpinner.Start()
			_, err := auth.LoginWithToken(tInput)
			authCodeSpinner.Stop()
			if err != nil {
				return fmt.Errorf("Failed to authenticate with token: %w", err)
			}

			fmt.Fprintln(utils.IO.Out, heredoc.Docf(i18n.T(`Signed in with the provided token`)))
		}

		return nil
	} else {
		scanner := bufio.NewScanner(utils.IO.In)
		input := []byte{}

		for scanner.Scan() {
			input = append(input, scanner.Bytes()...)
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("Failed to read token: %w", err)
		}

		authCodeSpinner := utils.CreateSpinner(i18n.T(" Verifying token..."), "")
		authCodeSpinner.Start()
		_, err := auth.LoginWithToken(string(input))
		authCodeSpinner.Stop()
		if err != nil {
			return fmt.Errorf("Failed to authenticate with token: %w", err)
		}

		fmt.Fprintln(utils.IO.Out, heredoc.Docf(i18n.T(`Signed in with the provided token`)))

		return nil
	}
}

var LoginCmd = &cobra.Command{
	Use:     "login [flags]",
	Aliases: []string{"signin"},
	Short:   i18n.T("login to the Integration Platform"),
	Long:    i18n.T("Authenticate with your Integration Platform account to start working on your projects."),
	Run: func(cmd *cobra.Command, args []string) {
		err := handleLogin(cmd, args)
		if err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddGenericHelper(LoginCmd)
	LoginCmd.Flags().BoolVar(
		&sio.withToken,
		"with-token",
		false,
		i18n.T("Authenticate with a token, token will be read from the standard input"),
	)

	LoginCmd.Flags().StringVar(
		&sio.orgHandle,
		"org",
		"",
		i18n.T("Organization handle"),
	)
	LoginCmd.Flags().StringVar(
		&sio.userEmail,
		"user",
		"",
		i18n.T("User email for the login"),
	)

}
