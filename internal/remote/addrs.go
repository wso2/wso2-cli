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
// Adapted from openchoreo internal/occ/cmd/remote/connect.go @ 69f36d373d0e7f9ad3f9dd21cf0c4960be75cd3d (v1.4.0-m.2).

package remote

import (
	"fmt"
	"io"
	"net"
	"strings"
)

// localHost is where every tunnel terminates, which is what lets one host binding
// serve several addresses of the same resource.
const localHost = "127.0.0.1"

// addrSwap pairs the in-cluster address a declared address resolved to with the local
// listener that replaces it, split into parts so a composed value carrying the host
// and port at separate positions can be re-pointed too.
type addrSwap struct {
	remote     string // in-cluster "host:port"
	local      string // "127.0.0.1:<local port>"
	remoteHost string
	remotePort string
	localPort  string
}

// newAddrSwap splits an in-cluster address so both the fused pair and its
// parts can be substituted. A remote address that will not split yields a swap that
// only ever matches as a whole.
func newAddrSwap(remote, localPort string) addrSwap {
	sw := addrSwap{
		remote:    remote,
		local:     net.JoinHostPort(localHost, localPort),
		localPort: localPort,
	}
	if host, port, err := net.SplitHostPort(remote); err == nil {
		sw.remoteHost, sw.remotePort = host, port
	}
	return sw
}

// splittable reports whether this swap can be applied to a host and a port held at
// separate positions in one value.
func (s addrSwap) splittable() bool {
	return s.remoteHost != "" && s.remotePort != ""
}

// rewriteAddrs re-points a resource's composed bindings at its tunnels, so a connection
// URL or driver connection string resolved for the cluster works from the developer's
// machine. Two shapes are handled, in order:
//
// Fused -- "redis://SVC:6379" -- has the host and port adjacent, so the pair is
// substituted as one string. Being that specific is what keeps the rewrite safe: it
// cannot match a host named with some other port, nor a bare hostname.
//
// Split -- "host=SVC,port=6379,password=x" -- holds the two apart, so they are
// substituted individually. That is only attempted for an address whose fused pair is
// absent from the value, and only when the value carries both the host AND that
// address's port as a delimited token. A value naming the host alone (a TLS server
// name) or with a different port (an admin URL) satisfies neither condition and is left
// as resolved, and reported. Each split rewrite is reported too, since substituting a
// bare host is the weaker inference of the two.
func rewriteAddrs(env map[string]string, swaps []addrSwap, ref string, out io.Writer) map[string]string {
	if len(env) == 0 || len(swaps) == 0 {
		return env
	}
	// Local listener ports, so a split rewrite never consumes a port an earlier rewrite
	// just produced. Ephemeral ports are assigned by the OS and could coincide with
	// another address's in-cluster port; substituting into that would silently undo the
	// earlier one's rewrite.
	localPorts := make(map[string]bool, len(swaps))
	for _, sw := range swaps {
		localPorts[sw.localPort] = true
	}

	rewritten := make(map[string]string, len(env))
	for k, v := range env {
		got := v
		for _, sw := range swaps {
			got = strings.ReplaceAll(got, sw.remote, sw.local)
		}
		// Decided against the original value, not the partially rewritten one: several
		// addresses of a resource share a host output, so the first rewrite would
		// otherwise hide the host from the addresses still to be applied.
		for _, sw := range swaps {
			if !sw.splittable() || strings.Contains(v, sw.remote) {
				continue
			}
			if !strings.Contains(v, sw.remoteHost) || !containsPortToken(v, sw.remotePort) {
				continue
			}
			if localPorts[sw.remotePort] {
				// This in-cluster port is some address's local port, so a
				// substitution here cannot be told apart from one already applied. Left
				// alone, and reported below as still pointing into the cluster.
				continue
			}
			got = strings.ReplaceAll(got, sw.remoteHost, localHost)
			got = replacePortToken(got, sw.remotePort, sw.localPort)
			fmt.Fprintf(out, "  ~ %s: %s had its host and port re-pointed separately\n", ref, k)
		}
		if got == v {
			// Unchanged, but mentions a tunneled host: the value embeds the host with
			// some other port, so it still points into the cluster.
			for _, sw := range swaps {
				if sw.remoteHost != "" && strings.Contains(v, sw.remoteHost) {
					fmt.Fprintf(out, "  ! %s: %s still points at %s and was not re-pointed at a tunnel\n", ref, k, sw.remoteHost)
					break
				}
			}
		}
		rewritten[k] = got
	}
	return rewritten
}

// containsPortToken reports whether port appears in s as a standalone token.
func containsPortToken(s, port string) bool {
	return findPortToken(s, port, 0) >= 0
}

// replacePortToken replaces every standalone occurrence of port in s with local.
func replacePortToken(s, port, local string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		j := findPortToken(s, port, i)
		if j < 0 {
			b.WriteString(s[i:])
			break
		}
		b.WriteString(s[i:j])
		b.WriteString(local)
		i = j + len(port)
	}
	return b.String()
}

// findPortToken returns the index of the first standalone occurrence of port in s at or
// after from, or -1. Standalone means not flanked by a character that could belong to a
// hostname or a longer number, so the 6379 in "cache6379.ns" or "63790" is not a match.
func findPortToken(s, port string, from int) int {
	for i := from; ; {
		j := strings.Index(s[i:], port)
		if j < 0 {
			return -1
		}
		j += i
		if !isAddrChar(s, j-1) && !isAddrChar(s, j+len(port)) {
			return j
		}
		i = j + 1
	}
}

// isAddrChar reports whether s[i] is a character a hostname or number can be made of.
// An index outside s is a delimiter, not a character.
func isAddrChar(s string, i int) bool {
	if i < 0 || i >= len(s) {
		return false
	}
	c := s[i]
	return c == '.' || c == '-' || c == '_' ||
		('0' <= c && c <= '9') || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}
