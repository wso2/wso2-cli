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
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// wsPair returns the two ends of a WebSocket as net.Conns: the dialing end and the
// accepted end.
func wsPair(t *testing.T) (client, server net.Conn) {
	t.Helper()
	accepted := make(chan net.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		accepted <- newWebSocketConn(ws)
	}))
	t.Cleanup(srv.Close)
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	client, server = newWebSocketConn(ws), <-accepted
	t.Cleanup(func() { _ = client.Close(); _ = server.Close() })
	return client, server
}

func TestWebSocketConnReadSpansMessageBoundaries(t *testing.T) {
	c, s := wsPair(t)
	for _, part := range []string{"hello ", "brave ", "new world"} {
		if _, err := c.Write([]byte(part)); err != nil {
			t.Fatal(err)
		}
	}
	got := make([]byte, len("hello brave new world"))
	if _, err := io.ReadFull(s, got); err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello brave new world" {
		t.Errorf("read %q", got)
	}
}

func TestWebSocketConnSmallReadsDrainOneMessage(t *testing.T) {
	c, s := wsPair(t)
	if _, err := c.Write([]byte("abcdef")); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	buf := make([]byte, 4)
	for out.Len() < 6 {
		n, err := s.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		out.Write(buf[:n])
	}
	if out.String() != "abcdef" {
		t.Errorf("read %q", out.String())
	}
}

func TestWebSocketConnConcurrentWritesStayWholeMessages(t *testing.T) {
	c, s := wsPair(t)
	const writers, size = 8, 4096
	var wg sync.WaitGroup
	for i := range writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.Write(bytes.Repeat([]byte{byte('a' + i)}, size)); err != nil {
				t.Error(err)
			}
		}()
	}
	got := make([]byte, writers*size)
	if _, err := io.ReadFull(s, got); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	for off := 0; off < len(got); off += size {
		if len(bytes.Trim(got[off:off+size], string(got[off:off+1]))) != 0 {
			t.Fatalf("message at %d interleaved with another writer", off)
		}
	}
}

func TestWebSocketConnCloseEndsThePeerRead(t *testing.T) {
	c, s := wsPair(t)
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Read(make([]byte, 1)); err != io.EOF {
		t.Errorf("peer read after close = %v, want io.EOF", err)
	}
	if _, err := c.Write([]byte("x")); err == nil {
		t.Error("write after close succeeded")
	}
}

func TestWebSocketConnReadDeadline(t *testing.T) {
	c, _ := wsPair(t)
	if err := c.SetReadDeadline(time.Now().Add(20 * time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	_, err := c.Read(make([]byte, 1))
	var ne net.Error
	if !errors.As(err, &ne) || !ne.Timeout() {
		t.Errorf("read = %v, want a timeout", err)
	}
}

func TestWebSocketConnAddresses(t *testing.T) {
	c, s := wsPair(t)
	if c.RemoteAddr().String() != s.LocalAddr().String() {
		t.Errorf("client remote %v != server local %v", c.RemoteAddr(), s.LocalAddr())
	}
}
