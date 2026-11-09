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

// Copyright 2026 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0
//
// Copied from openchoreo internal/remoteconnect/client.go @ 69f36d373d0e7f9ad3f9dd21cf0c4960be75cd3d (v1.4.0-m.2).

package remoteconnect

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/hashicorp/yamux"
)

// TunnelClient is the occ-side endpoint of a remote-connect tunnel: it performs the
// capability handshake over a single connection to a project+env remote-agent and
// multiplexes one stream per accepted local connection with yamux.
type TunnelClient struct {
	conn    net.Conn
	session *yamux.Session
	// capability supplies the capability to authorize each stream with. A function, so a
	// renewal takes effect on the next stream without touching this client.
	capability func() string
}

// HandshakeRejectedError reports that the remote-agent answered Hello with a refusal,
// such as the session limit or an unsupported protocol version. The agent is reachable
// and has given its answer, so redialing cannot change it.
type HandshakeRejectedError struct {
	Reason string
}

func (e *HandshakeRejectedError) Error() string {
	return "remoteconnect: tunnel handshake rejected: " + e.Reason
}

// NewTunnelClient runs the Hello/HelloResult handshake over an already-established
// connection (TLS in production; plain in tests) and layers a yamux client session.
// capability is called once per stream opened on this tunnel; it must not be nil.
func NewTunnelClient(conn net.Conn, capability func() string) (*TunnelClient, error) {
	if capability == nil {
		return nil, errors.New("remoteconnect: capability provider is required")
	}
	if err := WriteMessage(conn, Hello{ProtocolVersion: ProtocolVersion}); err != nil {
		return nil, fmt.Errorf("remoteconnect: send hello: %w", err)
	}
	var res HelloResult
	if err := ReadMessage(conn, &res); err != nil {
		return nil, fmt.Errorf("remoteconnect: read hello result: %w", err)
	}
	if !res.OK {
		return nil, &HandshakeRejectedError{Reason: res.Error}
	}

	ycfg := yamux.DefaultConfig()
	ycfg.LogOutput = io.Discard // quiet yamux; occ surfaces its own errors
	session, err := yamux.Client(conn, ycfg)
	if err != nil {
		return nil, fmt.Errorf("remoteconnect: yamux client: %w", err)
	}
	return &TunnelClient{conn: conn, session: session, capability: capability}, nil
}

// OpenStream opens a multiplexed stream and requests the target identified by key.
// The returned net.Conn is a transparent byte pipe to the dialed dependency.
func (c *TunnelClient) OpenStream(key string) (net.Conn, error) {
	return c.OpenStreamWith(key, c.capability())
}

// OpenStreamWith is OpenStream authorized by an explicit capability.
func (c *TunnelClient) OpenStreamWith(key, capability string) (net.Conn, error) {
	stream, err := c.session.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("remoteconnect: open stream: %w", err)
	}
	if err := WriteMessage(stream, StreamOpen{Key: key, Capability: capability}); err != nil {
		_ = stream.Close()
		return nil, err
	}
	var res StreamResult
	if err := ReadMessage(stream, &res); err != nil {
		_ = stream.Close()
		return nil, err
	}
	if !res.OK {
		_ = stream.Close()
		return nil, fmt.Errorf("remoteconnect: stream to %q rejected: %s", key, res.Error)
	}
	return stream, nil
}

// Fetch opens a stream for a value-fetch key and returns the value the remote-agent read
// from its own namespace. Unlike OpenStream this consumes the whole exchange: a fetch
// stream carries exactly one reply and is closed here, so no byte pipe is handed back.
//
// The value is returned to the caller and never logged, written to disk by this package,
// or included in an error — an error mentioning the value would put secret material into
// occ's terminal output.
func (c *TunnelClient) Fetch(key string) ([]byte, error) {
	if !IsSecretGrantKey(key) {
		return nil, fmt.Errorf("remoteconnect: %q is not a value-fetch key", key)
	}
	stream, err := c.session.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("remoteconnect: open stream: %w", err)
	}
	defer stream.Close()

	if err := WriteMessage(stream, StreamOpen{Key: key, Capability: c.capability()}); err != nil {
		return nil, err
	}
	var res SecretResult
	if err := ReadMessage(stream, &res); err != nil {
		return nil, err
	}
	if !res.OK {
		return nil, fmt.Errorf("%s", res.Error)
	}
	return res.Value, nil
}

// Close tears down the session and underlying connection.
func (c *TunnelClient) Close() error {
	if c.session != nil {
		_ = c.session.Close()
	}
	return c.conn.Close()
}
