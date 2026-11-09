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
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/wso2/wso2-cli/internal/remoteconnect"
)

// agentEndpointOverrideEnv, when set, makes the CLI dial this host:port instead
// of the endpoint the resolve call returned. It is for local development where
// the remote-agent's service is not routable from the laptop; the TLS pin (CA
// bundle and server name) still comes from the resolve response, so the agent's
// certificate is verified whatever the dial address.
const agentEndpointOverrideEnv = "WSO2_REMOTE_AGENT_ENDPOINT"

// dialDeadline bounds how long one tunnel dial may retry. It absorbs the window
// where a freshly provisioned bridge or service is still coming up.
const dialDeadline = 30 * time.Second

// AgentDialer opens tunnels the way production does: over the bridge when the
// resolved endpoint is wss://, over plain pinned TLS otherwise.
type AgentDialer struct {
	// Token returns the bearer the bridge is dialled with: the token the broker
	// was called with.
	Token func(ctx context.Context) (string, error)
	// UserAgent is sent to the bridge, as wso2-cli/<version>.
	UserAgent string
}

// Dial opens one tunnel to one remote-agent. capability is read once per
// stream, so a renewal takes effect without re-dialling.
func (a AgentDialer) Dial(ctx context.Context, agent remoteconnect.AgentEndpoint, capability func() string) (tunnel, error) {
	endpoint := agent.Endpoint
	if override := os.Getenv(agentEndpointOverrideEnv); override != "" {
		endpoint = override
	}
	if endpoint == "" {
		return nil, errors.New("resolve returned no remote-agent endpoint")
	}
	token, err := a.Token(ctx)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, dialDeadline)
	defer cancel()
	client, err := remoteconnect.DialTunnel(ctx, endpoint, agent.CABundle, agent.ServerName, capability,
		remoteconnect.DialOptions{BearerToken: token, UserAgent: a.UserAgent})
	if err != nil {
		return nil, fmt.Errorf("connect to remote-agent %s (%s): %w", endpoint, agent.ServerName, err)
	}
	return client, nil
}
