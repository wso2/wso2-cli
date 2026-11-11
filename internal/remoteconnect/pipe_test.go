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
// Copied from openchoreo internal/remoteconnect/pipe_test.go @ 69f36d373d0e7f9ad3f9dd21cf0c4960be75cd3d (v1.4.0-m.2).

package remoteconnect

import (
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

// upstreamEchoOnEOF answers only once its peer has finished writing, which is what a
// request/response protocol does.
func upstreamEchoOnEOF(t *testing.T, reply string) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		conn, aerr := ln.Accept()
		if aerr != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		_, _ = io.ReadAll(conn)
		_, _ = conn.Write([]byte(reply))
	}()
	return ln
}

// TestPipeDeliversReplyAfterHalfClose: a client that signals end-of-request with
// CloseWrite must still receive the response.
func TestPipeDeliversReplyAfterHalfClose(t *testing.T) {
	upstream := upstreamEchoOnEOF(t, "pong")

	local, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = local.Close() })
	go func() {
		down, aerr := local.Accept()
		if aerr != nil {
			return
		}
		up, derr := net.Dial("tcp", upstream.Addr().String())
		if derr != nil {
			return
		}
		Pipe(down, up)
	}()

	client, err := net.Dial("tcp", local.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	if _, err := client.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	if err := client.(*net.TCPConn).CloseWrite(); err != nil {
		t.Fatal(err)
	}

	_ = client.SetReadDeadline(time.Now().Add(5 * time.Second))
	got, err := io.ReadAll(client)
	if err != nil {
		t.Fatalf("read reply: %v", err)
	}
	if string(got) != "pong" {
		t.Fatalf("reply = %q, want %q", got, "pong")
	}
}

// trackedConn records whether Close was called.
type trackedConn struct {
	io.ReadWriteCloser
	mu     sync.Mutex
	closed bool
}

func (c *trackedConn) Close() error {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()
	return c.ReadWriteCloser.Close()
}

func (c *trackedConn) isClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

// TestPipeClosesBothWhenBothDirectionsEnd guards against leaking either endpoint.
func TestPipeClosesBothWhenBothDirectionsEnd(t *testing.T) {
	a1, a2 := net.Pipe()
	b1, b2 := net.Pipe()
	ta, tb := &trackedConn{ReadWriteCloser: a2}, &trackedConn{ReadWriteCloser: b1}

	done := make(chan struct{})
	go func() { Pipe(ta, tb); close(done) }()

	_ = a1.Close()
	_ = b2.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Pipe did not return after both directions ended")
	}
	if !ta.isClosed() || !tb.isClosed() {
		t.Errorf("endpoints left open: a=%v b=%v", ta.isClosed(), tb.isClosed())
	}
}
