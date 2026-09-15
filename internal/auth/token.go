package auth

import (
	"fmt"
	"os"

	"github.com/wso2/integration-platform-tools/internal/utils"

	"github.com/wso2/integration-platform-tools/internal/region"
	"github.com/wso2/integration-platform-tools/pkg/api"
)

// get a token for the given orgID, authenticating the user if necessary or refreshing the token if it has expired
// if a token for the given org ID is not found, it will try to find a token for a different org of the same user
// and then exchange it for a token for the given org ID
func (r *ReadOnlyTokenStore) getToken(orgID string, forceSignIn bool, orgsToCheck []api.Organization) (string, error) {

	token, err := tokenStore.RetrieveToken(region.GetCurrentRegion(), orgID)
	if err != nil {

		// check if a token is available for a different org of the same user
		userInfo, usrErr := usrStore.RetrieveUserInfo()
		if usrErr != nil {
			return "", usrErr
		}

		if userInfo.IsPATLogin {
			return "", fmt.Errorf("Switching orgs is not allowed with Personal Access Tokens")
		}

		if userInfo == nil && forceSignIn {
			if _, _, err = SignIn(&orgID); err != nil {
				return "", err
			}
			if token, err = tokenStore.RetrieveToken(region.GetCurrentRegion(), orgID); err != nil {
				return "", err
			}
		} else if userInfo != nil {
			var otherOrgToken string
			var targetOrg api.Organization
			// iterate through the orgs of the user and check if a token is available for any of them
			if orgsToCheck == nil {
				orgsToCheck = []api.Organization{}

				for _, org := range userInfo.Organizations {
					if org.ID != orgID {
						orgsToCheck = append(orgsToCheck, org)
					}
				}
			}

			for _, org := range userInfo.Organizations {
				if org.ID == orgID {
					targetOrg = org
					break
				}
			}
			// check if a token is available for a different org of the same user
			for i := 0; i < len(orgsToCheck); i++ {
				otherOrgToken, err = r.getToken(orgsToCheck[i].ID, forceSignIn, orgsToCheck[i+1:])
				if err == nil {
					break // Found a valid token
				}
			}

			if err == nil && otherOrgToken != "" && orgsToCheck != nil {
				token, exchangeErr := authClient.ExchangeSTSToken(otherOrgToken, targetOrg.Handle)
				if exchangeErr != nil {
					return "", exchangeErr
				}

				// Store the exchanged token
				if err = tokenStore.StoreToken(region.GetCurrentRegion(), orgID, token); err != nil {
					return "", err
				}

				return token.AccessToken, nil
			} else {
				err = fmt.Errorf(" unable to retrieve a token for org %s", orgID)
				return "", err
			}
		}
	} else if token.IsExpired() {
		// If the token has expired, refresh it.
		refreshedAccessToken := token.AccessToken
		// Carry the refresh token across the whole cycle. The STS token-exchange
		// response below does not reliably return one, so without this the
		// refresh token would survive exactly one renewal and the session would
		// be back to having no way to recover from going cold.
		carriedRefreshToken := token.RefreshToken

		if token.RefreshToken != "" {
			refreshedToken, refreshErr := authClient.ExchangeRefreshToken(token.RefreshToken)
			if refreshErr != nil {
				// Deliberately not fatal. The access token may still be inside its
				// expiry buffer and therefore usable as the exchange subject below,
				// which is exactly how renewal worked before refresh tokens were
				// requested at all. Failing here instead would make sessions with a
				// refresh token *less* recoverable than those without one.
				if os.Getenv("TRACE_ENABLED") == "true" {
					fmt.Fprintln(utils.IO.ErrOut,
						"refresh token exchange failed, falling back to exchanging the current token:", refreshErr)
				}
			} else {
				refreshedAccessToken = refreshedToken.AccessToken
				// Honour rotation: if the STS issued a new refresh token, the old
				// one may now be invalid.
				if refreshedToken.RefreshToken != "" {
					carriedRefreshToken = refreshedToken.RefreshToken
				}
			}
		}

		org, err := GetOrg(orgID)
		if err != nil {
			return "", err
		}
		// Exchange for an org-scoped access token from the WSO2 STS
		// FIXME: This is a temporary solution. We need to talk to the WSO2 Integration Platform STS team to get a better solution
		// by sending all the claims in the refresh token itself without having to exchange it for a new access token
		// again.
		token, err := authClient.ExchangeSTSToken(refreshedAccessToken, org.Handle)
		if err != nil {
			return "", err
		}

		// Preserve the refresh token if the exchange did not return one, so the
		// next renewal still has something to work with.
		if token.RefreshToken == "" {
			token.RefreshToken = carriedRefreshToken
		}

		// Store the refreshed token
		if err = tokenStore.StoreToken(region.GetCurrentRegion(), orgID, token); err != nil {
			return "", err
		}
		return token.AccessToken, nil
	}
	return token.AccessToken, nil
}
