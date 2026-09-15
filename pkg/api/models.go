package api

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

const (
	BUFFER_TIME      = 1 * time.Minute
	LONG_BUFFER_TIME = 20 * time.Minute
)

type Owner struct {
	ID        string    `json:"id"`
	IDPId     string    `json:"idpId"`
	CreatedAt time.Time `json:"createdAt"`
}

type Organization struct {
	ID     string  `json:"id"`
	UUID   string  `json:"uuid"`
	Handle string  `json:"handle"`
	Name   string  `json:"name"`
	Owner  Owner   `json:"owner,omitempty"`
	Region *string `json:"region,omitempty"`
}

// return a json string representation of the organization
func (o *Organization) String() string {
	organizationJson, err := json.Marshal(o)
	if err != nil {
		return ""
	}
	return string(organizationJson)
}

type UserInfo struct {
	DisplayName           string         `json:"displayName"`
	UserEmail             string         `json:"userEmail"`
	UserProfilePictureUrl string         `json:"userProfilePictureUrl"`
	IDPId                 string         `json:"idpId"`
	Organizations         []Organization `json:"organizations"`
	UserID                string         `json:"userId"`
	UserCreatedAt         time.Time      `json:"userCreatedAt"`
	IsPATLogin            bool           `json:"isPatLogin,omitempty"`
}

type ValidateUserFailure struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
	Cause   string `json:"cause"`
}

// return a json string representation of the user info
func (u *UserInfo) String() string {
	userInfoJson, err := json.Marshal(u)
	if err != nil {
		return ""
	}
	return string(userInfoJson)
}

type AccessToken struct {
	AccessToken    string `json:"access_token"`
	RefreshToken   string `json:"refresh_token"`
	LoginTime      string
	ExpirationTime int `json:"expires_in"`
}

type JwtToken struct {
	Exp       int64  `json:"exp"`
	Sub       string `json:"sub"`
	IdpClaims struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"idp_claims"`
	Organization struct {
		Handle string `json:"handle"`
		Uuid   string `json:"uuid"`
	} `json:"organization"`
}

// return a json string representation of the token
func (t *AccessToken) String() (string, error) {
	tokenJson, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(tokenJson), nil
}

// isExpired returns true if the token has expired
func (t *AccessToken) IsExpired() bool {
	claims := t.GetDecodedToken()
	if claims == nil {
		// will be unable to decode a PAT and therefore, we need to assume PAT has not expired yet
		return false
	}
	bufferTime := BUFFER_TIME
	if t.RefreshToken == "" {
		// have a long buffer time if refresh token does not exists
		bufferTime = LONG_BUFFER_TIME
	}
	expTime := time.Unix(claims.Exp, 0).Add(-bufferTime)
	isExpired := time.Now().After(expTime)
	return isExpired
}

func (t *AccessToken) GetDecodedToken() *JwtToken {
	// Try to decode the JWT access token and check the "exp" claim
	parts := strings.Split(t.AccessToken, ".")
	if len(parts) != 3 {
		return nil
	}
	// JWT uses base64url encoding without padding
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}
	var claims JwtToken
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil
	}

	return &claims
}

type ForbiddenApiResponse struct {
	NetworkError struct {
		StatusCode string `json:"statusCode"`
	} `json:"networkError,omitempty"`
	Message  string `json:"message,omitempty"`
	Metadata struct {
		ErrorCode      string `json:"errorCode"`
		AdditionalData string `json:"additionalData"`
	} `json:"metadata,omitempty"`
}

type ForbiddenDeclarativeResponse struct {
	StatusCode string `json:"statusCode"`
	Message    string `json:"message,omitempty"`
}

type ErrorResponse struct {
	StatusCode string `json:"statusCode"`
	Message    string `json:"message"`
}
