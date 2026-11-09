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

// Package remoteconnect is the client side of the remote-connect tunnel: the wire
// protocol shared with the remote-agent (protocol.go), the resolve request and response
// exchanged with the control plane (resolve.go), the tunnel client (client.go) and its
// dial (dial.go), and a byte pipe (pipe.go).
//
// protocol.go, resolve.go, client.go and pipe.go are copied from OpenChoreo
// v1.4.0-m.2 (see each file's provenance line). dial.go is new: it adds the WebSocket
// transport in front of the pinned TLS connection, for a remote-agent reached through
// the remote-connect bridge rather than directly.
package remoteconnect
