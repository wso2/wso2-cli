package auth

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/auth"
)

// loginAttempt records the outcome of the most recent browser login, which
// completes asynchronously after the login tool has already returned.
//
// Without this, both the callback wait and the sign-in ran in a goroutine whose
// errors were discarded, so check_login_status could only ever report
// "Not logged in" with no indication of why — indistinguishable from the user
// simply not having finished yet.
var loginAttempt struct {
	sync.Mutex
	started     bool
	finished    bool
	err         error
	callbackURL string
}

func recordLoginStarted(callbackURL string) {
	loginAttempt.Lock()
	defer loginAttempt.Unlock()
	loginAttempt.started = true
	loginAttempt.finished = false
	loginAttempt.err = nil
	loginAttempt.callbackURL = callbackURL
}

func recordLoginFinished(err error) {
	loginAttempt.Lock()
	defer loginAttempt.Unlock()
	loginAttempt.finished = true
	loginAttempt.err = err
}

// lastLoginFailure returns a human-readable reason the previous login attempt
// failed, or "" if there is nothing useful to report.
func lastLoginFailure() string {
	loginAttempt.Lock()
	defer loginAttempt.Unlock()

	if !loginAttempt.started {
		return ""
	}
	if !loginAttempt.finished {
		return fmt.Sprintf(
			"A login is still in progress — the browser has not completed sign-in yet. "+
				"The redirect must reach %s. If your browser is running on a different machine than this "+
				"server (for example a cloud editor), that address points at the browser's own machine and "+
				"the sign-in can never complete; see the note below.",
			loginAttempt.callbackURL)
	}
	if loginAttempt.err != nil {
		return fmt.Sprintf("The previous login attempt failed: %v", loginAttempt.err)
	}
	return ""
}

// remoteWithoutForwardedCallback reports whether this process looks like it is
// running in a cloud editor but was not given the port-forwarding base URL.
//
// In that state the browser flow cannot work: the callback falls back to
// localhost, which resolves to the user's own machine rather than to this
// process, so the redirect is delivered to nothing and the flow times out
// silently. Detection is best-effort — it relies on at least one CLOUD_* hint
// being present.
func remoteWithoutForwardedCallback() bool {
	if auth.IsCallbackPortForwarded() {
		return false
	}
	for _, v := range []string{"CLOUD_STS_TOKEN", "CLOUD_INITIAL_ORG_ID", "CLOUD_INITIAL_PROJECT_ID"} {
		if os.Getenv(v) != "" {
			return true
		}
	}
	return false
}

const remoteAuthGuidance = "This MCP server is a separate process from the editor, so it cannot use the editor's " +
	"own authentication. It can only complete a browser login if the redirect URL is reachable from your browser. " +
	"In a cloud editor the supported approach is to pass the editor's session to this server instead of running a " +
	"browser flow here: set CLOUD_STS_TOKEN in the MCP server's environment (and BASE_URL, so the callback uses the " +
	"port-forwarded address rather than localhost). In a VS Code mcp.json that is the server's \"env\" block."

func loginToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if auth.IsLoggedIn() {
		return mcp.NewToolResultText("You are already logged in."), nil
	}

	// Refuse rather than hand out a URL that cannot possibly work. Starting the
	// flow here would leave the user staring at a browser page that succeeds
	// while this server waits five minutes for a callback that is being
	// delivered to their laptop.
	if remoteWithoutForwardedCallback() {
		return mcp.NewToolResultError(fmt.Sprintf(
			"Cannot complete a browser login from here. This looks like a cloud editor environment "+
				"(a CLOUD_* variable is set) but BASE_URL is not, so the OAuth redirect would fall back to "+
				"http://127.0.0.1:<port>/auth-callback — an address that resolves to the machine running "+
				"your browser, not to this server. The sign-in would appear to succeed and then never "+
				"complete.\n\n%s",
			remoteAuthGuidance)), nil
	}

	redirectListenPort, link, err := auth.StartAuthFlow()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to start auth flow", err), nil
	}

	callbackURL := auth.CallbackURL(redirectListenPort)
	recordLoginStarted(callbackURL)

	go func() {
		authCode, err := auth.ListenForAuthCode(redirectListenPort, time.Minute*5)
		if err != nil {
			recordLoginFinished(fmt.Errorf(
				"no OAuth callback arrived at %s within 5 minutes (%w). If your browser runs on a different "+
					"machine than this server, that address is not reachable from it and the browser flow "+
					"cannot be used", callbackURL, err))
			return
		}
		if _, _, err := auth.SignInWithAuthCode(authCode, "", "", ""); err != nil {
			recordLoginFinished(fmt.Errorf("exchanging the authorization code failed: %w", err))
			return
		}
		recordLoginFinished(nil)
	}()

	note := ""
	if !auth.IsCallbackPortForwarded() {
		note = fmt.Sprintf(
			"\n\nNote: after sign-in your browser will be redirected to %s. That only works if the browser is "+
				"running on the same machine as this server. If it is not, the login will not complete — %s",
			callbackURL, remoteAuthGuidance)
	}

	return mcp.NewToolResultText(fmt.Sprintf(
		"Open this URL in your browser to authenticate:\n\n%s\n\n"+
			"After completing sign-in in your browser, call `check_login_status` to confirm authentication succeeded.%s",
		link, note,
	)), nil
}

func checkLoginStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !auth.IsLoggedIn() {
		msg := "Not logged in."
		if reason := lastLoginFailure(); reason != "" {
			msg += " " + reason
		} else {
			msg += " Use the `login` tool to authenticate, or run `wso2-integration-platform login` in your " +
				"terminal before starting this MCP server."
		}
		if remoteWithoutForwardedCallback() {
			msg += "\n\n" + remoteAuthGuidance
		}
		return mcp.NewToolResultText(msg), nil
	}

	user, err := auth.GetCurrentUser()
	if err != nil || user == nil {
		return mcp.NewToolResultText("Authenticated, but could not retrieve user info."), nil
	}

	name := user.UserEmail
	if name == "" {
		name = user.DisplayName
	}
	if name == "" {
		name = "the current user"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Logged in as %s", name)), nil
}
