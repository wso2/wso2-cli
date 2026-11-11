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

package remoteconnect

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// webSocketScheme marks an endpoint that is a WebSocket bridge in front of the
	// remote-agent router. Any other endpoint is a "host:port" dialed over plain TLS.
	webSocketScheme = "wss://"

	defaultRetryInitial = 2 * time.Second
	defaultRetryMax     = 30 * time.Second
)

// DialOptions tunes how DialTunnel reaches the endpoint. The zero value dials with the
// defaults and no WebSocket credentials.
type DialOptions struct {
	// BearerToken is sent as "Authorization: Bearer" on a WebSocket upgrade. It is
	// unused for a plain TLS endpoint, where there is no HTTP request to carry it.
	BearerToken string
	// UserAgent is sent on a WebSocket upgrade. Empty leaves the library default.
	UserAgent string
	// HTTPClient configures the WebSocket upgrade: its *http.Transport supplies the
	// TLS settings, proxy and dialer. Nil means http.DefaultTransport,
	// which holds the CLI's configured trust and proxy.
	HTTPClient *http.Client
	// RetryInitial and RetryMax bound the WebSocket dial retry loop: the wait starts at
	// RetryInitial and doubles up to RetryMax. Zero means 2 s and 30 s.
	RetryInitial, RetryMax time.Duration
}

// DialTunnel connects to a remote-agent and constructs a TunnelClient. The endpoint
// decides the transport:
//
//   - "wss://host/path" is a WebSocket bridge. The connection is upgraded with the
//     bearer token, wrapped as a net.Conn, and the agent's TLS runs inside it.
//   - anything else is a "host:port" dialed with TLS directly.
//
// Either way the agent presents its own self-signed certificate rather than one from a
// gateway CA, so the client pins caBundle and verifies against serverName (a fixed SAN
// baked into the cert, decoupled from the runtime address). If caBundle is empty, the
// system roots are used.
func DialTunnel(ctx context.Context, endpoint, caBundle, serverName string, capability func() string, opts DialOptions) (*TunnelClient, error) {
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if caBundle != "" {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(caBundle)) {
			return nil, errors.New("remoteconnect: invalid agent CA bundle")
		}
		tlsCfg.RootCAs = pool
	}
	if serverName != "" {
		tlsCfg.ServerName = serverName
	}

	conn, err := dialAgentConn(ctx, endpoint, tlsCfg, opts)
	if err != nil {
		return nil, err
	}
	c, err := NewTunnelClient(conn, capability)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return c, nil
}

// dialAgentConn returns a TLS connection to the agent, over a WebSocket or directly.
func dialAgentConn(ctx context.Context, endpoint string, tlsCfg *tls.Config, opts DialOptions) (net.Conn, error) {
	if !strings.HasPrefix(endpoint, webSocketScheme) {
		conn, err := (&tls.Dialer{Config: tlsCfg}).DialContext(ctx, "tcp", endpoint)
		if err != nil {
			return nil, fmt.Errorf("remoteconnect: dial remote-agent %s: %w", endpoint, err)
		}
		return conn, nil
	}

	raw, err := dialWebSocket(ctx, endpoint, opts)
	if err != nil {
		return nil, err
	}
	conn := tls.Client(raw, tlsCfg)
	if err := conn.HandshakeContext(ctx); err != nil {
		_ = raw.Close()
		return nil, fmt.Errorf("remoteconnect: tls handshake with remote-agent via %s: %w", endpoint, err)
	}
	return conn, nil
}

// dialWebSocket upgrades endpoint to a WebSocket and returns it as a byte-stream
// net.Conn, retrying transient failures with backoff until ctx ends.
func dialWebSocket(ctx context.Context, endpoint string, opts DialOptions) (net.Conn, error) {
	header := http.Header{}
	if opts.BearerToken != "" {
		header.Set("Authorization", "Bearer "+opts.BearerToken)
	}
	if opts.UserAgent != "" {
		header.Set("User-Agent", opts.UserAgent)
	}
	wait, ceiling := opts.RetryInitial, opts.RetryMax
	if wait <= 0 {
		wait = defaultRetryInitial
	}
	if ceiling <= 0 {
		ceiling = defaultRetryMax
	}

	dialer := websocketDialer(opts.HTTPClient)

	for {
		ws, resp, err := dialer.DialContext(ctx, endpoint, header)
		if err == nil {
			// ctx bounds the dial, not the tunnel: the connection lives until Close.
			return newWebSocketConn(ws), nil
		}
		if resp != nil {
			err = fmt.Errorf("%w (bridge answered %s)", err, resp.Status)
		}
		if !retryableDial(resp) {
			return nil, fmt.Errorf("remoteconnect: dial remote-connect bridge %s: %w", endpoint, err)
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("remoteconnect: dial remote-connect bridge %s: %w (last error: %v)", endpoint, ctx.Err(), err)
		case <-time.After(wait):
		}
		wait = min(wait*2, ceiling)
	}
}

// websocketDialer builds the upgrade dialer from the transport of client, so a caller
// that pins roots or routes through a proxy for HTTP gets the same for the bridge. A
// nil client, or one without a transport, means the CLI's configured default,
// http.DefaultTransport, which carries the trust from WSO2_CA_FILE; gorilla's dialer
// never reads it on its own.
func websocketDialer(client *http.Client) *websocket.Dialer {
	d := &websocket.Dialer{Proxy: http.ProxyFromEnvironment, HandshakeTimeout: 45 * time.Second}
	rt := http.DefaultTransport
	if client != nil && client.Transport != nil {
		rt = client.Transport
	}
	if t, ok := rt.(*http.Transport); ok {
		d.Proxy = t.Proxy
		d.NetDialContext = t.DialContext
		d.TLSClientConfig = httpOneTLSConfig(t.TLSClientConfig)
	}
	return d
}

// httpOneTLSConfig returns a copy of base, nil-safe, that offers only HTTP/1.1 in
// ALPN. net/http adds "h2" to a transport's own config on first use, and gorilla
// speaks only the HTTP/1.1 Upgrade, so a gateway that picked h2 would break the
// handshake. The copy keeps the trust; the shared config is never modified.
func httpOneTLSConfig(base *tls.Config) *tls.Config {
	cfg := base.Clone()
	if cfg == nil {
		cfg = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	cfg.NextProtos = []string{"http/1.1"}
	return cfg
}

// retryableDial reports whether a failed upgrade may succeed on a later attempt. No
// response means the network failed; a 5xx, 408 or 429 means the bridge or gateway is
// not ready or is shedding load. Any other status is an answer, such as a refused
// token, and repeating the request cannot change it.
func retryableDial(resp *http.Response) bool {
	if resp == nil {
		return true
	}
	switch {
	case resp.StatusCode >= 500, resp.StatusCode == http.StatusRequestTimeout, resp.StatusCode == http.StatusTooManyRequests:
		return true
	default:
		return false
	}
}
