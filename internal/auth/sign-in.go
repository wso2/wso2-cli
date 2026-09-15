package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/golang-jwt/jwt/v5"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/browser"
	"github.com/wso2/integration-platform-tools/internal/region"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api"
)

func StartAuthFlow() (int, string, error) {
	port, err := getRandomPort()
	if err != nil {
		return 0, "", fmt.Errorf("error occurred while getting a random port: %v", err)
	}

	localCallbackUrl := getCallbackURL(port)

	regionStrs := []string{}
	clientIds := []string{}
	regions := region.GetValidRegions()
	for _, regionStr := range regions {
		regionStrs = append(regionStrs, regionStr)
		regionConfig := region.GetConfigByRegion(regionStr)
		clientIds = append(clientIds, regionConfig.DevantConfig.AsgardeoClientId)
	}
	link, err := GetAuthUrl(localCallbackUrl, "", "", regionStrs, clientIds)
	if err != nil {
		return 0, "", fmt.Errorf("error occurred while building the auth URL: %w", err)
	}

	return port, link, nil
}

func getSanitizedName(name string, maxLength int) string {
	var sb strings.Builder
	for _, r := range name { // Iterate over the original case
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		}
	}
	sanitized := sb.String()
	if len(sanitized) > 0 && unicode.IsDigit(rune(sanitized[0])) {
		sanitized = "n" + sanitized
	}

	if len(sanitized) > maxLength {
		return sanitized[:maxLength]
	}
	return sanitized
}

func findNewValidOrgName(accessToken string) (string, error) {
	claims := jwt.MapClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(accessToken, claims)
	if err != nil {
		return "", err
	}
	fullName := claims["name"].(string)
	email := claims["email"].(string)
	sanitizedName := getSanitizedName(fullName, 26)
	for i := 0; i < 1000; i++ {
		newOrgName := sanitizedName
		if i > 0 {
			newOrgName = fmt.Sprintf("%s-%d", sanitizedName, i)
		}
		isValid, err := usrMgtClient.IsValidNewOrg(accessToken, newOrgName, email)
		if isValid {
			sanitizedName = newOrgName
			break
		}
		if err != nil {
			return "", err
		}
	}
	return sanitizedName, nil
}

func HandleAuthCode(port int, orgID *string) (info *api.UserInfo, isNewUser bool, err error) {
	var authCode string
	var asgardeoToken *api.AccessToken
	var org *api.Organization
	var token *api.AccessToken
	authCodeSpinner := utils.CreateSpinner(i18n.T(" Waiting for you to complete the login process..."), "")
	authCodeSpinner.Start()

	// wait till user completes the login process
	// and the local server receives the auth code
	authCode, err = ListenForAuthCode(port, time.Minute*5)
	authCodeSpinner.Stop()
	if err != nil {
		return
	}

	tokenExchangeSpinner := utils.CreateSpinner(
		i18n.T(" Authenticating..."),
		i18n.T("Successfully authenticated"),
	)
	tokenExchangeSpinner.Start()

	// exchange the auth code for an asgardeo token
	asgardeoToken, err = authClient.ExchangeAuthCode(authCode, "", "")
	tokenExchangeSpinner.Stop()
	if err != nil {
		return
	}

	validateUserSpinner := utils.CreateSpinner(i18n.T(" Validating user..."), "")
	validateUserSpinner.Start()

	// validate the asgardeo token and get the user info
	userInfo, err := usrMgtClient.ValidateUser(asgardeoToken.AccessToken)
	validateUserSpinner.Stop()
	if err != nil {
		if errors.Is(err, api.ErrNoAccountFound) {
			orgNameGenSpinner := utils.CreateSpinner(i18n.T(" Finding a valid org name..."), "")
			orgNameGenSpinner.Start()
			newOrgName, err := findNewValidOrgName(asgardeoToken.AccessToken)
			orgNameGenSpinner.Stop()
			if err != nil {
				return nil, false, err
			}

			userRegisterSpinner := utils.CreateSpinner(i18n.T(" Registering new user..."), "")
			userRegisterSpinner.Start()
			userInfo, err = usrMgtClient.RegisterUser(asgardeoToken.AccessToken, newOrgName)
			userRegisterSpinner.Stop()
			if err != nil {
				return nil, false, err
			}
			isNewUser = true
		} else {
			return
		}
	}

	// clear auth store now that we have a valid token
	ClearAuthStores()

	if err = usrStore.StoreUserInfo(*userInfo); err != nil {
		return
	}
	fmt.Fprintf(utils.IO.Out, i18n.T("Welcome, %s!\n\n"), utils.CS.Bold(userInfo.UserEmail))

	fetchingOrgsSpinner := utils.CreateSpinner(i18n.T(" Fetching Organizations..."), "")
	fetchingOrgsSpinner.Start()

	// Pick the organization to be used, select previous or default if no orgID is given
	if orgID == nil {
		if org, err = GetSelectedOrganization(); err != nil {
			if !errors.Is(KeyringEntryNotFound, err) {
				fetchingOrgsSpinner.Stop()
				return
			}
		}
	} else {
		for _, cOrg := range userInfo.Organizations {
			if cOrg.ID == *orgID {
				org = &cOrg
				break
			}
		}
		if err = orgStore.SetDefaultOrg(org); err != nil {
			fetchingOrgsSpinner.Stop()
			return
		}
	}
	fetchingOrgsSpinner.Stop()

	// if the org is still nil, this is a newly authenticating user therefore we can rely on the org list obtained
	// through the userInfo object
	if org == nil {
		if len(userInfo.Organizations) == 0 {
			err = fmt.Errorf("no organization info found for the user")
			return
		}

		org = &userInfo.Organizations[0]
		if err = orgStore.SetDefaultOrg(org); err != nil {
			return
		}
	}

	if len(userInfo.Organizations) > 1 {
		fmt.Fprintf(utils.IO.Out, i18n.T("Selected organization '%s'\n\n"), org.Name)
	}

	orgTokenSpinner := utils.CreateSpinner(i18n.T(" Validating Organization access..."), "")
	orgTokenSpinner.Start()
	// exchange the asgardeo token for a VSCode token for the selected org
	if token, err = authClient.ExchangeSTSToken(asgardeoToken.AccessToken, org.Handle); err != nil {
		orgTokenSpinner.Stop()
		return
	}
	if err = tokenStore.StoreToken(region.GetCurrentRegion(), org.ID, token); err != nil {
		orgTokenSpinner.Stop()
		return
	}
	orgTokenSpinner.Stop()

	info = userInfo
	return
}

func SignIn(orgID *string) (userInfo *api.UserInfo, isNewUser bool, err error) {

	var redirectListenPort int
	var link string

	var loginMode string
	loginOpts := []string{i18n.T("Open link using default browser"), i18n.T("Get a link to open in browser manually")}
	err = prompt.NewPromptSelectMessage[string](
		prompt.PromptSelectOpts[string]{
			Values:  loginOpts,
			Default: loginOpts[0],
			Message: i18n.T("How would you like to sign in?"),
		},
		&loginMode,
	).Prompt()

	if err != nil {
		return
	}

	// start local server and open browser
	if redirectListenPort, link, err = StartAuthFlow(); err != nil {
		return
	}

	switch loginMode {
	case loginOpts[0]:
		brwsr := browser.New("", iostreams.System().Out, iostreams.System().ErrOut)
		brwsr.Browse(link)
	case loginOpts[1]:
		fmt.Fprintln(
			utils.IO.Out,
			fmt.Sprintf(i18n.T("\nCopy the following link and open it in your browser:\n\n%s\n"), link),
		)
	default:
		// This should never happen
	}

	return HandleAuthCode(redirectListenPort, orgID)
}

func GetAuthUrl(callBackUrl string, baseUrl string, clientId string, regions []string, clientIds []string) (string, error) {
	return authClient.GetAuthURL(callBackUrl, baseUrl, clientId, regions, clientIds)
}

func SignInWithAuthCode(authCode string, orgID string, redirectUrl string, clientId string) (info *api.UserInfo, isNewUser bool, err error) {
	var asgardeoToken *api.AccessToken
	var org *api.Organization
	var token *api.AccessToken

	// exchange the auth code for an asgardeo token
	asgardeoToken, err = authClient.ExchangeAuthCode(authCode, redirectUrl, clientId)
	if err != nil {
		return
	}

	// validate the asgardeo token and get the user info
	userInfo, err := usrMgtClient.ValidateUser(asgardeoToken.AccessToken)
	if err != nil {
		if errors.Is(err, api.ErrNoAccountFound) {
			newOrgName, err := findNewValidOrgName(asgardeoToken.AccessToken)
			if err != nil {
				return nil, false, err
			}

			userInfo, err = usrMgtClient.RegisterUser(asgardeoToken.AccessToken, newOrgName)
			if err != nil {
				return nil, false, err
			}
			isNewUser = true
		} else {
			return
		}
	}

	if len(userInfo.Organizations) == 0 {
		err = api.NoOrgsAvailable
		return
	}

	// clear auth store now that we have a valid token
	ClearAuthStores()

	if err = usrStore.StoreUserInfo(*userInfo); err != nil {
		return
	}

	// Pick the organization to be used, select previous or default if no orgID is given
	if orgID == "" {
		if org, err = GetSelectedOrganization(); err != nil {
			if !errors.Is(KeyringEntryNotFound, err) {
				return
			}
		}
	} else {
		for _, cOrg := range userInfo.Organizations {
			if cOrg.ID == orgID {
				org = &cOrg
				break
			}
		}
		if err = orgStore.SetDefaultOrg(org); err != nil {
			return
		}
	}

	// if the org is still nil, this is a newly authenticating user therefore we can rely on the org list obtained
	// through the userInfo object
	if org == nil {
		if len(userInfo.Organizations) == 0 {
			err = api.NoOrgsAvailable
			return
		}

		org = &userInfo.Organizations[0]
		if err = orgStore.SetDefaultOrg(org); err != nil {
			return
		}
	}

	// exchange the asgardeo token for a VSCode token for the selected org
	if token, err = authClient.ExchangeSTSToken(asgardeoToken.AccessToken, org.Handle); err != nil {
		return
	}
	if err = tokenStore.StoreToken(region.GetCurrentRegion(), org.ID, token); err != nil {
		return
	}

	info = userInfo
	return
}

func SetInitialRegionOfOrg() error {
	selectedOrg, err := GetSelectedOrganization()
	if err == nil && selectedOrg != nil {
		deploymentPipelineSpinner := utils.CreateSpinner(i18n.T(" Fetching deployment pipelines..."), "")
		deploymentPipelineSpinner.Start()

		deploymentPipelineData, err := DevopsClient.GetDeploymentPipeline(selectedOrg.UUID, selectedOrg.ID)
		deploymentPipelineSpinner.Stop()
		if err == nil && len(deploymentPipelineData) == 0 {
			initRegionSpinner := utils.CreateSpinner(i18n.T(" Setting initial region..."), "")
			initRegionSpinner.Start()
			DevopsClient.InitOrgRegion(selectedOrg.ID, selectedOrg.UUID)
			initRegionSpinner.Stop()
		}
	}
	return nil
}
