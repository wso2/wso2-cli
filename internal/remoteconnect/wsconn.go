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
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// closeFrameTimeout bounds how long Close waits to send the close frame before it
// drops the connection anyway.
const closeFrameTimeout = time.Second

// webSocketConn presents a WebSocket as a byte-stream net.Conn. Only binary messages
// carry data. Message boundaries are invisible to the reader, and each Write is sent
// as one binary message.
//
// Read must not be called concurrently with itself, and neither must Write; both
// match the net.Conn use of tls.Conn.
type webSocketConn struct {
	ws *websocket.Conn

	reader io.Reader // the message being read; nil between messages

	writeMu sync.Mutex
}

func newWebSocketConn(ws *websocket.Conn) *webSocketConn {
	return &webSocketConn{ws: ws}
}

func (c *webSocketConn) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	for {
		if c.reader == nil {
			typ, r, err := c.ws.NextReader()
			if err != nil {
				return 0, normalizeReadError(err)
			}
			if typ != websocket.BinaryMessage {
				return 0, fmt.Errorf("remoteconnect: websocket message type %d, want binary", typ)
			}
			c.reader = r
		}
		n, err := c.reader.Read(p)
		if errors.Is(err, io.EOF) {
			c.reader = nil
			err = nil
		}
		if n > 0 || err != nil {
			return n, err
		}
	}
}

// normalizeReadError reports a clean close from the peer as io.EOF.
func normalizeReadError(err error) error {
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
		return io.EOF
	}
	return err
}

func (c *webSocketConn) Write(p []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if err := c.ws.WriteMessage(websocket.BinaryMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close sends a normal-closure frame, then closes the connection. The frame is best
// effort: the connection is closed whether or not it was delivered.
func (c *webSocketConn) Close() error {
	_ = c.ws.WriteControl(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(closeFrameTimeout))
	return c.ws.Close()
}

func (c *webSocketConn) LocalAddr() net.Addr  { return c.ws.LocalAddr() }
func (c *webSocketConn) RemoteAddr() net.Addr { return c.ws.RemoteAddr() }

func (c *webSocketConn) SetDeadline(t time.Time) error {
	return errors.Join(c.ws.SetReadDeadline(t), c.ws.SetWriteDeadline(t))
}
func (c *webSocketConn) SetReadDeadline(t time.Time) error  { return c.ws.SetReadDeadline(t) }
func (c *webSocketConn) SetWriteDeadline(t time.Time) error { return c.ws.SetWriteDeadline(t) }
