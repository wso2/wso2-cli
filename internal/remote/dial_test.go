// Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
//
// WSO2 LLC. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package remote

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/wso2/wso2-cli/internal/remoteconnect"
)

// The bridge is dialled with the token the broker is called with and the CLI's
// user agent, and a token that cannot be minted stops the dial.
func TestAgentDialerPresentsTheTokenToTheBridge(t *testing.T) {
	t.Setenv(agentEndpointOverrideEnv, "")
	headers := make(chan http.Header, 1)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers <- r.Header.Clone()
		if conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil); err == nil {
			_ = conn.Close()
		}
	}))
	defer server.Close()

	dialer := AgentDialer{
		Token:      staticToken("bridge-token"),
		UserAgent:  "wso2-cli/9.9.9",
		HTTPClient: server.Client(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// The agent's TLS never answers, so the dial itself fails; the upgrade
	// request is what is under test.
	_, err := dialer.Dial(ctx, remoteconnect.AgentEndpoint{
		Endpoint: "wss" + strings.TrimPrefix(server.URL, "https") + "/connect", ServerName: "agent.remote-connect",
	}, func() string { return "cap" })
	if err == nil {
		t.Fatal("want the dial to fail without an agent behind the bridge")
	}
	header := <-headers
	if got := header.Get("Authorization"); got != "Bearer bridge-token" {
		t.Errorf("Authorization = %q", got)
	}
	if got := header.Get("User-Agent"); got != "wso2-cli/9.9.9" {
		t.Errorf("User-Agent = %q", got)
	}
}

func TestAgentDialerNeedsAnEndpointAndAToken(t *testing.T) {
	t.Setenv(agentEndpointOverrideEnv, "")
	dialer := AgentDialer{Token: staticToken("t")}
	if _, err := dialer.Dial(context.Background(), remoteconnect.AgentEndpoint{}, nil); err == nil {
		t.Error("an empty endpoint: want an error")
	}
	failing := AgentDialer{Token: func(context.Context) (string, error) { return "", context.DeadlineExceeded }}
	if _, err := failing.Dial(context.Background(), remoteconnect.AgentEndpoint{Endpoint: "x:1"}, nil); err == nil {
		t.Error("no token: want an error")
	}
}
