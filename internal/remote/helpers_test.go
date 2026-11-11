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
// Adapted from openchoreo internal/occ/cmd/remote/connect_test.go @ 69f36d373d0e7f9ad3f9dd21cf0c4960be75cd3d (v1.4.0-m.2).

package remote

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeRenewable stands in for the remoteUnit a reporter consults.
type fakeRenewable struct {
	expiry  time.Time
	stopped bool
	nudges  atomic.Int32
}

func (f *fakeRenewable) expiresAt() time.Time { return f.expiry }

func (f *fakeRenewable) renewing() bool { return !f.stopped }

func (f *fakeRenewable) nudge() { f.nudges.Add(1) }

// fakeTunnel stands in for a yamux session to a remote-agent: each OpenStream dials a
// local echo server, so Connect's local plumbing (listeners, env, per-connection
// stream open) can be exercised without a real remote-agent. The occ -> remote-agent ->
// dependency chain's own hops are covered in their own packages.
type fakeTunnel struct {
	addr   string
	onOpen func(key string)
	// values answers Fetch by key. A key not present is answered as a read failure,
	// which is how the control plane's denial reaches occ.
	values   map[string][]byte
	fetchErr error
	fetched  []string
}

func (f *fakeTunnel) OpenStreamWith(key, _ string) (net.Conn, error) {
	if f.onOpen != nil {
		f.onOpen(key)
	}
	return net.Dial("tcp", f.addr)
}

func (f *fakeTunnel) Fetch(key string) ([]byte, error) {
	f.fetched = append(f.fetched, key)
	if f.fetchErr != nil {
		return nil, f.fetchErr
	}
	v, ok := f.values[key]
	if !ok {
		return nil, errors.New("read failed")
	}
	return v, nil
}

func (f *fakeTunnel) Close() error { return nil }

// startEchoServer starts a plain TCP echo server, standing in for the far end of a
// remote-connect stream (in production, this is the tunneled dependency).
func startEchoServer(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) { defer func() { _ = c.Close() }(); _, _ = io.Copy(c, c) }(c)
		}
	}()
	return ln
}

// syncBuf is a writer safe to read while forward's goroutines are still writing.
type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// mintExpiredCapability builds a capability-shaped token that expired an hour
// ago. Only the exp claim matters: the CLI reads it unverified, purely to explain
// a failure.
func mintExpiredCapability(t *testing.T) string {
	t.Helper()
	encode := func(v any) string {
		raw, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return base64.RawURLEncoding.EncodeToString(raw)
	}
	header := encode(map[string]string{"alg": "EdDSA", "typ": "JWT"})
	claims := encode(map[string]any{"exp": time.Now().Add(-time.Hour).Unix()})
	return header + "." + claims + "." + base64.RawURLEncoding.EncodeToString([]byte("signature"))
}

func envToMap(env []string) map[string]string {
	m := map[string]string{}
	for _, kv := range env {
		if k, v, ok := strings.Cut(kv, "="); ok {
			m[k] = v
		}
	}
	return m
}

func tunnelRoundTrip(addr, msg string) error {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.Write([]byte(msg)); err != nil {
		return err
	}
	buf := make([]byte, len(msg))
	if _, err := io.ReadFull(conn, buf); err != nil {
		return err
	}
	if string(buf) != msg {
		return io.ErrUnexpectedEOF
	}
	return nil
}
