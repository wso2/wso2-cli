package auth

// a local server that listens for the oauth callback
import (
	"context"
	"errors"

	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/wso2/integration-platform-tools/internal/region"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

const PATH = "/auth-callback"

type AuthResult struct {
	Code string
	Err  error
}

func ListenForAuthCode(port int, timeout time.Duration) (string, error) {
	mux := http.NewServeMux()
	server := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: mux}

	authCodeChan := make(chan AuthResult)

	mux.HandleFunc(PATH, func(w http.ResponseWriter, r *http.Request) {
		authCode := r.URL.Query().Get("code")
		regionParam := r.URL.Query().Get("region")
		if regionParam != "" {
			err := region.SetCurrentRegion(regionParam)
			if err != nil {
				fmt.Fprintln(utils.IO.ErrOut, "Failed to set region from callback:", err)
			}
			ReInit()
		}
		if authCode == "" {
			authCodeChan <- AuthResult{"", errors.New("no code in oauth callback")}
		} else {
			authCodeChan <- AuthResult{authCode, nil}
		}
		writeSuccessHtml(w)
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			authCodeChan <- AuthResult{"", fmt.Errorf("ListenAndServe(): %v", err)}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	defer func() {
		if err := server.Shutdown(ctx); err != nil {
			// handle err
			fmt.Fprintln(utils.IO.ErrOut, "Server shutdown error:", err)
		}
	}()

	select {
	case result := <-authCodeChan:
		return result.Code, result.Err
	case <-time.After(timeout):
		return "", errors.New("timeout waiting for auth code")
	}
}

// CallbackURL exposes the OAuth redirect URI the auth flow will listen on, so
// callers can show it to the user. Whether it is reachable from the user's
// browser depends on where that browser runs: the localhost form only works when
// the browser is on the same machine as this process.
func CallbackURL(port int) string {
	return getCallbackURL(port)
}

// IsCallbackPortForwarded reports whether the callback URL is the port-forwarded
// public form rather than plain localhost. When false and the browser is remote,
// the redirect cannot reach this process and the flow will silently time out.
func IsCallbackPortForwarded() bool {
	return os.Getenv("BASE_URL") != ""
}

// getCallbackURL returns the OAuth redirect URI for the given port.
// In a remote VS Code environment (BASE_URL set), it uses the port-forwarded
// public URL; otherwise falls back to localhost.
func getCallbackURL(port int) string {
	if base := os.Getenv("BASE_URL"); base != "" {
		return fmt.Sprintf("https://%d-%s%s", port, base, PATH)
	}
	return fmt.Sprintf("http://127.0.0.1:%d%s", port, PATH)
}

func getRandomPort() (int, error) {
	var minPort, maxPort int = 55152, 55252

	for port := minPort; port <= maxPort; port++ {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			defer listener.Close()
			return port, nil
		}
	}

	return -1, errors.New("no available ports")
}

func writeSuccessHtml(w http.ResponseWriter) {
	// set the content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, oauthSuccessPage)
}
