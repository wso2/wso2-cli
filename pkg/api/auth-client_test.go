package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// The STS is asked for offline_access so it returns a refresh token. A server
// that does not have that scope registered rejects the whole request rather than
// downgrading, so the retry must make this strictly no worse than not asking.
func TestExchangeSTSTokenOfflineAccessFallback(t *testing.T) {
	tests := []struct {
		name            string
		rejectOffline   bool
		wantAttempts    int
		wantOfflineSent []bool // per attempt
	}{
		{"STS accepts offline_access", false, 1, []bool{true}},
		{"STS rejects it, retry without", true, 2, []bool{true, false}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var mu sync.Mutex
			var offlineSeen []bool

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.ParseForm()
				hasOffline := strings.Contains(r.Form.Get("scope"), OfflineAccessScope)
				mu.Lock()
				offlineSeen = append(offlineSeen, hasOffline)
				mu.Unlock()

				if tc.rejectOffline && hasOffline {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprint(w, `{"error":"invalid_scope","error_description":"scope offline_access not allowed"}`)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				resp := map[string]any{"access_token": "at", "expires_in": 3600}
				if hasOffline {
					resp["refresh_token"] = "rt"
				}
				json.NewEncoder(w).Encode(resp)
			}))
			defer srv.Close()

			c := &AuthClient{Config: &AuthClientConfig{STSTokenUrl: srv.URL, STSClientId: "cid"}}
			tok, err := c.ExchangeSTSToken("subject", "acme")
			if err != nil {
				t.Fatalf("exchange failed: %v", err)
			}

			mu.Lock()
			defer mu.Unlock()
			if len(offlineSeen) != tc.wantAttempts {
				t.Fatalf("attempts = %d, want %d", len(offlineSeen), tc.wantAttempts)
			}
			for i, want := range tc.wantOfflineSent {
				if offlineSeen[i] != want {
					t.Errorf("attempt %d sent offline_access=%v, want %v", i+1, offlineSeen[i], want)
				}
			}
			// When the STS accepts it we must actually keep the refresh token.
			if !tc.rejectOffline && tok.RefreshToken == "" {
				t.Error("refresh token was not captured from the response")
			}
		})
	}
}

// A non-scope failure must not be retried — doubling the latency of every real
// auth failure would be worse than the problem being solved.
func TestExchangeSTSTokenDoesNotRetryNonScopeFailures(t *testing.T) {
	var attempts int
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		attempts++
		mu.Unlock()
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":"invalid_client"}`)
	}))
	defer srv.Close()

	c := &AuthClient{Config: &AuthClientConfig{STSTokenUrl: srv.URL, STSClientId: "cid"}}
	if _, err := c.ExchangeSTSToken("subject", "acme"); err == nil {
		t.Fatal("expected an error")
	}
	mu.Lock()
	defer mu.Unlock()
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1 — a 401 is not a scope rejection", attempts)
	}
}

func TestIsLikelyScopeRejection(t *testing.T) {
	yes := []string{
		`unexpected response: sts-token-exchange: {"error":"invalid_scope"}: status code: 400`,
		`{"error":"invalid_request"}`,
		`something: status code: 400`,
	}
	no := []string{
		`unexpected response: {"error":"invalid_client"}: status code: 401`,
		`dial tcp: connection refused`,
		`status code: 500`,
	}
	for _, m := range yes {
		if !isLikelyScopeRejection(fmt.Errorf("%s", m)) {
			t.Errorf("should be treated as a scope rejection: %q", m)
		}
	}
	for _, m := range no {
		if isLikelyScopeRejection(fmt.Errorf("%s", m)) {
			t.Errorf("should NOT be retried: %q", m)
		}
	}
}
