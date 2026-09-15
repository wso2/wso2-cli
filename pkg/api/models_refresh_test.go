package api

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func jwtExpiring(in time.Duration) string {
	h := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`))
	b, _ := json.Marshal(map[string]any{"exp": time.Now().Add(in).Unix(), "sub": "u"})
	return h + "." + base64.RawURLEncoding.EncodeToString(b) + ".sig"
}

// Whether a session can survive going idle hinges entirely on this: with a
// refresh token the CLI renews 1 minute before expiry and can recover a fully
// cold session; without one it must renew 20 minutes early by exchanging a
// still-valid token, so any idle period past expiry ends the session.
func TestExpiryBufferDependsOnRefreshToken(t *testing.T) {
	cases := []struct {
		name        string
		remaining   time.Duration
		refresh     string
		wantExpired bool
	}{
		{"no refresh token, 25m left", 25 * time.Minute, "", false},
		{"no refresh token, 15m left", 15 * time.Minute, "", true},
		{"has refresh token, 15m left", 15 * time.Minute, "rt", false},
		{"has refresh token, 2m left", 2 * time.Minute, "rt", false},
		{"has refresh token, 30s left", 30 * time.Second, "rt", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tok := &AccessToken{AccessToken: jwtExpiring(tc.remaining), RefreshToken: tc.refresh}
			if got := tok.IsExpired(); got != tc.wantExpired {
				t.Errorf("IsExpired() = %v, want %v", got, tc.wantExpired)
			}
		})
	}
}

// A PAT cannot be decoded as a JWT and has no refresh token, so it must never be
// treated as expired — there is nothing to refresh it with.
func TestPATNeverExpires(t *testing.T) {
	tok := &AccessToken{AccessToken: "chp_notajwt", RefreshToken: ""}
	if tok.IsExpired() {
		t.Error("a PAT must not be reported as expired")
	}
}
