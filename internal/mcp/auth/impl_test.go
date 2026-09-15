package auth

import (
	"errors"
	"os"
	"strings"
	"testing"
)

var errNoCallback = errors.New("no callback arrived")

func withEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for k, v := range kv {
		orig, had := os.LookupEnv(k)
		if v == "" {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, v)
		}
		t.Cleanup(func() {
			if had {
				os.Setenv(k, orig)
			} else {
				os.Unsetenv(k)
			}
		})
	}
}

// The MCP server is a separate process from the editor, so it cannot use the
// editor's auth. In a cloud editor without BASE_URL the OAuth callback falls
// back to localhost, which resolves to the machine running the browser — the
// redirect is delivered to nothing and the flow times out after five minutes
// with no explanation. Detect that state instead of starting a doomed flow.
func TestRemoteWithoutForwardedCallback(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{
			"local: no cloud hints, no BASE_URL",
			map[string]string{"BASE_URL": "", "CLOUD_STS_TOKEN": "", "CLOUD_INITIAL_ORG_ID": "", "CLOUD_INITIAL_PROJECT_ID": ""},
			false,
		},
		{
			"cloud, properly configured: BASE_URL present",
			map[string]string{"BASE_URL": "editor.example.dev", "CLOUD_STS_TOKEN": "t"},
			false,
		},
		{
			"cloud, misconfigured: STS token but no BASE_URL",
			map[string]string{"BASE_URL": "", "CLOUD_STS_TOKEN": "t"},
			true,
		},
		{
			"cloud, misconfigured: org hint but no BASE_URL",
			map[string]string{"BASE_URL": "", "CLOUD_STS_TOKEN": "", "CLOUD_INITIAL_ORG_ID": "org-1"},
			true,
		},
		{
			"cloud, misconfigured: project hint but no BASE_URL",
			map[string]string{"BASE_URL": "", "CLOUD_STS_TOKEN": "", "CLOUD_INITIAL_ORG_ID": "", "CLOUD_INITIAL_PROJECT_ID": "p-1"},
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withEnv(t, tc.env)
			if got := remoteWithoutForwardedCallback(); got != tc.want {
				t.Errorf("remoteWithoutForwardedCallback() = %v, want %v", got, tc.want)
			}
		})
	}
}

// The guidance must point at the supported cloud path rather than suggesting the
// user retry a browser flow that structurally cannot work here.
func TestRemoteAuthGuidance(t *testing.T) {
	for _, want := range []string{"CLOUD_STS_TOKEN", "BASE_URL", "separate process", "mcp.json"} {
		if !strings.Contains(remoteAuthGuidance, want) {
			t.Errorf("guidance must mention %q", want)
		}
	}
}

// check_login_status previously could only say "Not logged in", which is
// indistinguishable from the user simply not having finished signing in yet.
func TestLastLoginFailureReporting(t *testing.T) {
	t.Run("nothing attempted", func(t *testing.T) {
		loginAttempt.Lock()
		loginAttempt.started, loginAttempt.finished, loginAttempt.err = false, false, nil
		loginAttempt.Unlock()
		if got := lastLoginFailure(); got != "" {
			t.Errorf("expected no reason before any attempt, got %q", got)
		}
	})

	t.Run("in progress names the callback URL", func(t *testing.T) {
		recordLoginStarted("http://127.0.0.1:55152/auth-callback")
		got := lastLoginFailure()
		if !strings.Contains(got, "still in progress") {
			t.Errorf("should report an in-flight attempt, got %q", got)
		}
		if !strings.Contains(got, "127.0.0.1:55152") {
			t.Error("should name the callback URL so an unreachable one is self-evident")
		}
	})

	t.Run("failure is surfaced", func(t *testing.T) {
		recordLoginStarted("http://127.0.0.1:55152/auth-callback")
		recordLoginFinished(errNoCallback)
		got := lastLoginFailure()
		if !strings.Contains(got, "previous login attempt failed") || !strings.Contains(got, "no callback") {
			t.Errorf("should surface the underlying error, got %q", got)
		}
	})

	t.Run("success clears the reason", func(t *testing.T) {
		recordLoginStarted("http://127.0.0.1:55152/auth-callback")
		recordLoginFinished(nil)
		if got := lastLoginFailure(); got != "" {
			t.Errorf("expected no reason after success, got %q", got)
		}
	})
}
