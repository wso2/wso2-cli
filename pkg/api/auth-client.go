package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	pkce "github.com/grokify/go-pkce"
)

const (
	AccessTokenError  = "error while exchanging the access token"
	RefreshTokenError = "error while exchanging the refresh token"
	VSCodeTokenError  = "error while exchanging the VSCode token"

	ExchangeGrantType          = "urn:ietf:params:oauth:grant-type:token-exchange"
	JWTTokenType               = "urn:ietf:params:oauth:token-type:jwt"
	RefreshTokenGrantType      = "refresh_token"
	AuthorizationCodeGrantType = "authorization_code"
	// OfflineAccessScope asks the STS to issue a refresh token alongside the
	// access token. Without it the exchange returns an access token only, which
	// leaves the CLI unable to renew a session that has gone fully cold — the
	// only other renewal path is exchanging a still-valid token, so any idle
	// period longer than the token lifetime ends the session.
	//
	// If the STS does not have this scope registered for the client it may reject
	// the request outright, so ExchangeSTSToken retries once without it.
	OfflineAccessScope = "offline_access"

	Scope = "choreo:log_view_prod choreo:log_view_non_prod apim:api_manage apim:subscription_manage apim:tier_manage apim:admin apim:publisher_settings environments:view_prod environments:view_dev choreo:user_manage choreo:role_manage apim:dcr:app_manage choreo:deployment_manage choreo:non_prod_env_manage choreo:prod_env_manage choreo:component_manage choreo:project_manage apim:api_publish apim:document_manage apim:api_settings apim:subscription_view choreo:env_manage choreo:log_view"
)

var (
	CommonReqHeaders = map[string]string{
		"Content-Type": "application/x-www-form-urlencoded; charset=utf8",
		"Accept":       "application/json",
	}
)

type AuthClientConfig struct {
	AsgardeoClientId string
	RedirectUrl      string
	AsgardeoTokenUrl string
	STSClientId      string
	STSTokenUrl      string
	STSScopes        string
	LoginUrl         string
	SignUpUrl        string
}

type AuthClient struct {
	Verifier string

	Config *AuthClientConfig
}

func (c *AuthClient) generatePKCEChallenge() error {
	// Generate a new PKCE verifier
	verifier, err := pkce.NewCodeVerifier(32)
	if err != nil {
		return fmt.Errorf("generating PKCE code verifier: %w", err)
	}
	c.Verifier = verifier
	return nil
}

func (c *AuthClient) getHttpClient() *http.Client {
	client := &http.Client{}
	if os.Getenv("TRACE_ENABLED") == "true" {
		// Auth logger: this client posts and receives tokens, so bodies are
		// suppressed in addition to the usual header redaction.
		client.Transport = CreateAuthLogger()(client.Transport)
	}
	return client
}

func (c *AuthClient) ExchangeAuthCode(authCode string, redirectUrl string, clientId string) (*AccessToken, error) {
	data := url.Values{}
	redirect_uri := redirectUrl
	if redirect_uri == "" {
		redirect_uri = c.Config.RedirectUrl
	}
	client_id := clientId
	if client_id == "" {
		client_id = c.Config.AsgardeoClientId
	}
	data.Set("client_id", client_id)
	data.Set("code", authCode)
	data.Set("grant_type", AuthorizationCodeGrantType)
	data.Set("redirect_uri", redirect_uri)
	data.Set("code_verifier", c.Verifier)

	req, err := http.NewRequest("POST", c.Config.AsgardeoTokenUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	for key, value := range CommonReqHeaders {
		req.Header.Set(key, value)
	}

	client := c.getHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New(AccessTokenError)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected response: auth-code-exchange: %s: status code: %d", string(bodyBytes), resp.StatusCode)
	}

	var result AccessToken
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding auth-code-exchange response: %w", err)
	}

	result.LoginTime = time.Now().Format(time.RFC3339)

	return &result, nil
}

// ExchangeSTSToken exchanges a subject token for an org-scoped platform token.
//
// It first asks for offline_access so the response carries a refresh token. An
// STS that does not have that scope registered for this client rejects the whole
// request rather than downgrading, so a scope-shaped failure is retried once
// without it. That keeps this strictly no worse than not asking at all.
func (c *AuthClient) ExchangeSTSToken(apimAccessToken string, orgHandle string) (*AccessToken, error) {
	token, err := c.exchangeSTSToken(apimAccessToken, orgHandle, true)
	if err == nil {
		return token, nil
	}
	if !isLikelyScopeRejection(err) {
		return nil, err
	}
	return c.exchangeSTSToken(apimAccessToken, orgHandle, false)
}

// isLikelyScopeRejection reports whether an exchange failure looks like the STS
// refusing the requested scopes, as opposed to a network or credential problem.
// Retrying those blindly would double the latency of every genuine failure.
func isLikelyScopeRejection(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "scope") ||
		strings.Contains(msg, "invalid_request") ||
		strings.Contains(msg, "status code: 400")
}

func (c *AuthClient) exchangeSTSToken(apimAccessToken string, orgHandle string, requestOfflineAccess bool) (*AccessToken, error) {
	data := url.Values{}
	data.Set("client_id", c.Config.STSClientId)
	data.Set("subject_token", apimAccessToken)
	data.Set("subject_token_type", JWTTokenType)
	data.Set("grant_type", ExchangeGrantType)
	data.Set("requested_token_type", JWTTokenType)
	scopes := Scope + " " + c.Config.STSScopes
	if requestOfflineAccess {
		scopes += " " + OfflineAccessScope
	}
	data.Set("scope", scopes)
	data.Set("orgHandle", orgHandle)

	req, err := http.NewRequest("POST", c.Config.STSTokenUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return &AccessToken{}, err
	}

	for key, value := range CommonReqHeaders {
		req.Header.Set(key, value)
	}

	client := c.getHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return &AccessToken{}, errors.New(VSCodeTokenError)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected response: sts-token-exchange: %s: status code: %d", string(bodyBytes), resp.StatusCode)
	}

	var result AccessToken

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	result.LoginTime = time.Now().Format(time.RFC3339)

	return &result, nil
}

func (c *AuthClient) ExchangeRefreshToken(refreshToken string) (*AccessToken, error) {
	data := url.Values{}
	data.Set("client_id", c.Config.STSClientId)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", RefreshTokenGrantType)

	req, err := http.NewRequest("POST", c.Config.STSTokenUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	for key, value := range CommonReqHeaders {
		req.Header.Set(key, value)
	}

	client := c.getHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrRefreshToken
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, ErrRefreshToken
	}

	var result AccessToken
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding refresh-token response: %w", err)
	}
	result.LoginTime = time.Now().Format(time.RFC3339)

	return &result, nil
}

func (c *AuthClient) GetAuthURL(localCallBackURI string, baseUrl string, clientId string, regions []string, clientIds []string) (string, error) {
	if err := c.generatePKCEChallenge(); err != nil {
		return "", err
	}

	type AuthStateRegions struct {
		Region   string `json:"region"`
		ClientId string `json:"clientId"`
	}

	type AuthState struct {
		Origin      string             `json:"origin"`
		CallbackUri string             `json:"callbackUri"`
		Regions     []AuthStateRegions `json:"regions"`
	}

	state := AuthState{
		Origin:      "vscode.wso2ip.ext",
		CallbackUri: localCallBackURI,
	}

	for index, regionStr := range regions {
		state.Regions = append(state.Regions, AuthStateRegions{Region: regionStr, ClientId: clientIds[index]})
	}

	stateJson, err := json.Marshal(state)
	if err != nil {
		return "", fmt.Errorf("marshalling auth state: %w", err)
	}
	stateBase64 := base64.StdEncoding.EncodeToString(stateJson)
	challenge := pkce.CodeChallengeS256(c.Verifier)
	loginBaseUrl := baseUrl
	if loginBaseUrl == "" {
		loginBaseUrl = c.Config.LoginUrl
	}

	client_id := clientId
	if client_id == "" {
		client_id = c.Config.AsgardeoClientId
	}

	return loginBaseUrl +
		"?profile=vs-code&client_id=" +
		client_id +
		"&code_challenge=" +
		challenge +
		"&code_challenge_method=S256&state=" +
		stateBase64, nil
}

func (c *AuthClient) GetSignUpURL() string {
	return c.Config.SignUpUrl
}
