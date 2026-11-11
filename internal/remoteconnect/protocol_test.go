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
	"reflect"
	"testing"

	"github.com/hashicorp/yamux"
)

func TestMessageRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	want := StreamOpen{Key: "ep/proj/comp/http", Capability: "cap-token"}
	if err := WriteMessage(&buf, want); err != nil {
		t.Fatal(err)
	}
	var got StreamOpen
	if err := ReadMessage(&buf, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip = %+v, want %+v", got, want)
	}
}

func TestReadMessageRefusesOversizedLengthPrefix(t *testing.T) {
	// 0x00300000 = 3 MiB, above the 2 MiB ceiling.
	err := ReadMessage(bytes.NewReader([]byte{0x00, 0x30, 0x00, 0x00}), &Hello{})
	if !errors.Is(err, ErrMessageTooLarge) {
		t.Fatalf("err = %v, want ErrMessageTooLarge", err)
	}
}

// fakeAgent answers the Hello handshake, layers a yamux server, and serves each stream
// with serve. It stands in for the remote-agent on the far side of conn.
func fakeAgent(t *testing.T, conn net.Conn, serve func(stream net.Conn)) {
	t.Helper()
	go func() {
		var hello Hello
		if err := ReadMessage(conn, &hello); err != nil {
			return
		}
		if err := WriteMessage(conn, HelloResult{OK: hello.ProtocolVersion == ProtocolVersion}); err != nil {
			return
		}
		session, err := yamux.Server(conn, nil)
		if err != nil {
			return
		}
		for {
			stream, aerr := session.AcceptStream()
			if aerr != nil {
				return
			}
			go serve(stream)
		}
	}()
}

func TestOpenStreamPresentsKeyAndCapabilityThenPipes(t *testing.T) {
	clientConn, agentConn := net.Pipe()
	var opened StreamOpen
	fakeAgent(t, agentConn, func(stream net.Conn) {
		defer func() { _ = stream.Close() }()
		if err := ReadMessage(stream, &opened); err != nil {
			return
		}
		_ = WriteMessage(stream, StreamResult{OK: true})
		_, _ = io.Copy(stream, stream) // echo
	})

	client, err := NewTunnelClient(clientConn, func() string { return "cap-1" })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	stream, err := client.OpenStream("res/db/client")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stream.Close() }()
	if _, err := stream.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 4)
	if _, err := io.ReadFull(stream, buf); err != nil {
		t.Fatal(err)
	}
	if string(buf) != "ping" {
		t.Fatalf("echo = %q", buf)
	}
	if opened.Key != "res/db/client" || opened.Capability != "cap-1" {
		t.Fatalf("agent saw %+v", opened)
	}
}

func TestOpenStreamReportsAgentRefusal(t *testing.T) {
	clientConn, agentConn := net.Pipe()
	fakeAgent(t, agentConn, func(stream net.Conn) {
		defer func() { _ = stream.Close() }()
		var so StreamOpen
		_ = ReadMessage(stream, &so)
		_ = WriteMessage(stream, StreamResult{OK: false, Error: "not authorized"})
	})
	client, err := NewTunnelClient(clientConn, func() string { return "cap" })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	if _, err := client.OpenStream("ep/x"); err == nil {
		t.Fatal("expected the refusal to surface")
	}
}
