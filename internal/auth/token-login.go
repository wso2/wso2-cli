package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/wso2/integration-platform-tools/internal/region"
	"github.com/wso2/integration-platform-tools/pkg/api"
)

type PatIntrospectResp struct {
	ID               string    `json:"id"`
	OrganizationUUID string    `json:"organizationUuid"`
	UserID           string    `json:"userId"`
	AllowedScopes    []string  `json:"allowedScopes"`
	ValidUntil       time.Time `json:"validUntil"`
	Status           string    `json:"status"`
}

type OrgsResponse []struct {
	ID     string `json:"id"`
	UUID   string `json:"uuid"`
	Handle string `json:"handle"`
	Name   string `json:"name"`
	Owner  struct {
		ID        string    `json:"id"`
		IdpID     string    `json:"idpId"`
		CreatedAt time.Time `json:"createdAt"`
	} `json:"owner"`
}

func makeRequest[T any](method, url, token string) (result *T, err error) {
	c := &http.Client{}

	if os.Getenv("TRACE_ENABLED") == "true" {
		// Carries the PAT/STS bearer token — suppress bodies as well as headers.
		c.Transport = api.CreateAuthLogger()(c.Transport)
	}

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := c.Do(req)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		err = fmt.Errorf("HTTP %d: %s", res.StatusCode, res.Status)
		return nil, err
	}

	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func GetPATInfo(authToken string) (*PatIntrospectResp, error) {
	envConfigs := GetEnvConfig()

	patInfo, err := makeRequest[PatIntrospectResp]("POST", fmt.Sprintf("%s/%s", envConfigs.Apis.APIKeyService, "pat/introspect"), authToken)
	if err != nil {
		return nil, fmt.Errorf("PAT introspection failed: %w", err)
	}
	if patInfo == nil {
		return nil, fmt.Errorf("PAT introspection returned no data")
	}

	return patInfo, nil
}

func LoginWithToken(authToken string) (info *api.UserInfo, err error) {
	envConfigs := GetEnvConfig()

	patInfo, err := GetPATInfo(authToken)
	if err != nil {
		return nil, err
	}

	orgsRes, err := makeRequest[OrgsResponse](
		"GET",
		envConfigs.Apis.OrgsAPI,
		authToken,
	)
	if err != nil {
		return nil, fmt.Errorf("fetching organizations failed: %w", err)
	}
	if orgsRes == nil || len(*orgsRes) == 0 {
		return nil, fmt.Errorf("no organizations found for the provided PAT")
	}

	orgHandle := ""
	for _, org := range *orgsRes {
		if patInfo.OrganizationUUID == org.UUID {
			orgHandle = org.Handle
			break
		}
	}
	if orgHandle == "" {
		return nil, fmt.Errorf("organization matching PAT not found in orgs list")
	}

	org, err := makeRequest[struct {
		Organization api.Organization `json:"organization"`
	}](
		"GET",
		fmt.Sprintf("%s/%s", envConfigs.Apis.OrgsAPI, orgHandle),
		authToken,
	)
	if err != nil {
		return nil, fmt.Errorf("fetching organization details failed: %w", err)
	}
	if org == nil {
		return nil, fmt.Errorf("organization details response is nil")
	}

	var cUser api.UserInfo = api.UserInfo{
		DisplayName: "",
		UserEmail:   "",
		IDPId:       patInfo.UserID,
		Organizations: []api.Organization{
			org.Organization,
		},
		UserID:                "",
		UserCreatedAt:         org.Organization.Owner.CreatedAt,
		UserProfilePictureUrl: "",
		IsPATLogin:            true,
	}

	if err = usrStore.StoreUserInfo(cUser); err != nil {
		return nil, fmt.Errorf("failed to store user info: %w", err)
	}

	if err = orgStore.SetDefaultOrg(&org.Organization); err != nil {
		return nil, fmt.Errorf("failed to set default org: %w", err)
	}

	if err = tokenStore.StoreToken(region.GetCurrentRegion(), org.Organization.ID, &api.AccessToken{
		AccessToken:    authToken,
		LoginTime:      time.Now().Format(time.RFC3339),
		RefreshToken:   "",
		ExpirationTime: 0,
	}); err != nil {
		return nil, fmt.Errorf("failed to store token: %w", err)
	}

	info = &cUser
	return info, nil
}

type stsTokenClaims struct {
	Organization struct {
		Handle string `json:"handle"`
		UUID   string `json:"uuid"`
	} `json:"organization"`
	Sub string `json:"sub"`
}

func parseSTSTokenClaims(token string) (*stsTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}
	payload := parts[1]
	for len(payload)%4 != 0 {
		payload += "="
	}
	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}
	var claims stsTokenClaims
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims: %w", err)
	}
	return &claims, nil
}

// LoginWithSTSToken bootstraps a session from an existing STS token.
// It stores the token, org, and user info so subsequent calls use the
// normal token store path (including expiry detection and refresh).
// Uses makeRequest directly (not OrgClient) to avoid the chicken-and-egg
// problem where OrgClient needs a stored token to function.
func LoginWithSTSToken(stsToken string) error {
	claims, err := parseSTSTokenClaims(stsToken)
	if err != nil {
		return fmt.Errorf("failed to parse STS token: %w", err)
	}

	if claims.Organization.Handle == "" {
		return fmt.Errorf("no organization handle claim found in STS token")
	}

	envConfigs := GetEnvConfig()
	orgURL := fmt.Sprintf("%s/%s", envConfigs.Apis.OrgsAPI, claims.Organization.Handle)

	orgResp, err := makeRequest[struct {
		Organization api.Organization `json:"organization"`
	}](
		"GET",
		orgURL,
		stsToken,
	)
	if err != nil {
		return fmt.Errorf("failed to get organization details: %w", err)
	}
	if orgResp == nil {
		return fmt.Errorf("organization details response is nil")
	}
	org := &orgResp.Organization

	// Try to exchange for a proper platform STS token so the normal
	// expiry/refresh path works. If the exchange endpoint doesn't accept
	// this token type, fall back to storing the STS token directly.
	tokenToStore := &api.AccessToken{
		AccessToken:  stsToken,
		LoginTime:    time.Now().Format(time.RFC3339),
		RefreshToken: "",
	}
	if exchangedToken, exchErr := authClient.ExchangeSTSToken(stsToken, org.Handle); exchErr == nil {
		tokenToStore = exchangedToken
	}

	if err := tokenStore.StoreToken(region.GetCurrentRegion(), org.ID, tokenToStore); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	if err := orgStore.SetDefaultOrg(org); err != nil {
		return fmt.Errorf("failed to set default org: %w", err)
	}

	if err := usrStore.StoreUserInfo(api.UserInfo{
		IDPId:         claims.Sub,
		Organizations: []api.Organization{*org},
		IsPATLogin:    true,
	}); err != nil {
		return fmt.Errorf("failed to store user info: %w", err)
	}

	return nil
}
