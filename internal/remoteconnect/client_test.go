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
// Copied from openchoreo internal/remoteconnect/client_test.go @ 69f36d373d0e7f9ad3f9dd21cf0c4960be75cd3d (v1.4.0-m.2).

package remoteconnect

import (
	"errors"
	"net"
	"strings"
	"testing"
)

// A refusal is reported as HandshakeRejectedError so callers can tell the agent's answer
// apart from a connection that has not come up yet.
func TestNewTunnelClientReportsRefusalAsTypedError(t *testing.T) {
	client, server := net.Pipe()
	t.Cleanup(func() { _ = client.Close(); _ = server.Close() })

	const reason = "this remote-agent is already serving its maximum number of sessions"
	go func() {
		var hello Hello
		_ = ReadMessage(server, &hello)
		_ = WriteMessage(server, HelloResult{OK: false, Error: reason})
	}()

	_, err := NewTunnelClient(client, func() string { return "capability" })
	if err == nil {
		t.Fatal("expected the handshake to be rejected")
	}
	var rejected *HandshakeRejectedError
	if !errors.As(err, &rejected) {
		t.Fatalf("error %v is not a *HandshakeRejectedError", err)
	}
	if rejected.Reason != reason {
		t.Errorf("reason = %q, want %q", rejected.Reason, reason)
	}
	if !strings.Contains(err.Error(), reason) {
		t.Errorf("error text %q does not carry the reason", err.Error())
	}
}
