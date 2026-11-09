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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"

	"github.com/wso2/wso2-cli/internal/remoteconnect"
)

// tunnel is one yamux session to a project+env remote-agent; each accepted local
// connection opens a stream on it for a resolved target key.
type tunnel interface {
	// OpenStreamWith authorizes the stream with the given capability rather than the
	// tunnel's current one, so a renewal cannot repoint a stream already being opened.
	OpenStreamWith(key, capability string) (net.Conn, error)
	// Fetch resolves one value-fetch key to the bytes the remote-agent read from its own
	// namespace. Separate from OpenStream because a fetch stream is a single
	// request/response, not a byte pipe.
	Fetch(key string) ([]byte, error)
	Close() error
}

// mergeOverrides copies src into dst, warning when a key is already set to a different
// value by an earlier workload in this invocation.
func mergeOverrides(dst map[string]string, out io.Writer, src map[string]string) {
	for k, v := range src {
		if existing, ok := dst[k]; ok && existing != v {
			fmt.Fprintf(out, "  ! warning: %s set by multiple workloads (%q vs %q); using %q\n", k, existing, v, v)
		}
		dst[k] = v
	}
}

// forward accepts local connections and pipes each over a fresh yamux stream to the
// remote-agent (opened via open).
func forward(ln net.Listener, key string, open func() (net.Conn, error), report func(string, error)) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return // listener closed
		}
		go func(local net.Conn) {
			stream, serr := open()
			if serr != nil {
				// Without this the app saw only a connection reset, which made an expired
				// session look like the dependency had gone away.
				report(key, serr)
				_ = local.Close()
				return
			}
			remoteconnect.Pipe(local, stream)
		}(conn)
	}
}

// renewable is the part of a remoteUnit a reporter consults.
type renewable interface {
	expiresAt() time.Time
	renewing() bool
	nudge()
}

// streamErrorReporter prints the first failure for each dependency. A dependency that
// fails once usually fails for every subsequent connection, so repeating it would bury
// the session in noise.
type streamErrorReporter struct {
	out io.Writer
	// unit is read per failure rather than captured, because renewal moves the expiry.
	unit    renewable
	mu      sync.Mutex
	printed map[string]bool
}

func newStreamErrorReporter(out io.Writer, unit renewable) *streamErrorReporter {
	return &streamErrorReporter{out: out, unit: unit, printed: map[string]bool{}}
}

func (r *streamErrorReporter) report(key string, err error) {
	// Every failure nudges, not only the first one printed; the renewal loop decides
	// whether the capability is what is failing.
	r.unit.nudge()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.printed[key] {
		return
	}
	r.printed[key] = true
	// The renewal that dropped the key already reported the reason.
	if errors.Is(err, errRevoked) {
		fmt.Fprintf(r.out, "  ! %s: %v\n", key, err)
		return
	}
	if exp := r.unit.expiresAt(); !exp.IsZero() && time.Now().After(exp) {
		// A lapsed capability with a renewal still pending needs no action.
		if r.unit.renewing() {
			fmt.Fprintf(r.out, "  ! %s: the session expired at %s and is being renewed now — "+
				"retry in a moment\n", key, exp.Local().Format(time.Kitchen))
			return
		}
		fmt.Fprintf(r.out, "  ! %s: the session expired at %s and is no longer being renewed — "+
			"exit and re-run `wso2 remote connect` to reconnect\n", key, exp.Local().Format(time.Kitchen))
		return
	}
	fmt.Fprintf(r.out, "  ! %s: %v\n", key, err)
}

func targetKind(t remoteconnect.ResolvedTarget) string {
	if t.Resource != nil {
		return "resource/" + t.Resource.Address
	}
	return "endpoint"
}

// mergeEnv overlays overrides onto a base environment ("KEY=VALUE" slice).
func mergeEnv(base []string, overrides map[string]string) []string {
	merged := make(map[string]string, len(base)+len(overrides))
	for _, kv := range base {
		if k, v, ok := strings.Cut(kv, "="); ok {
			merged[k] = v
		}
	}
	maps.Copy(merged, overrides)
	out := make([]string, 0, len(merged))
	for k, v := range merged {
		out = append(out, k+"="+v)
	}
	sort.Strings(out)
	return out
}

// printEnvBindings lists the resolved bindings for --print-env. Values fetched from the
// data plane are redacted unless showSecrets: they are the same credentials the cluster
// keeps in a Secret, and this output lands in a terminal, scrollback, and often a pasted
// bug report. A redacted binding still shows its name, so the developer can see it
// resolved.
func printEnvBindings(out io.Writer, env map[string]string, sensitive map[string]bool, showSecrets bool) {
	fmt.Fprintln(out, "\nEnvironment bindings:")
	for _, k := range sortedKeys(env) {
		if sensitive[k] && !showSecrets {
			fmt.Fprintf(out, "  export %s=<hidden; pass --show-secrets to print it>\n", k)
			continue
		}
		fmt.Fprintf(out, "  export %s=%s\n", k, env[k])
	}
}

// confirmShowSecrets asks whether to print the named fetched values in full. It answers
// no unless stdin is a terminal: --print-env is routinely piped, and a pipe must not
// block on a prompt nothing will answer.
func confirmShowSecrets(out io.Writer, names []string) bool {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false
	}
	fmt.Fprintf(out, "\nRead from a Secret or ConfigMap: %s\n", strings.Join(names, ", "))
	fmt.Fprint(out, "Print these values in full? They stay in this terminal's scrollback. [y/N]: ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	}
	return false
}

// sortedKeys returns m's keys in sorted order, so printed output is stable.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func runInteractiveShell(ctx context.Context, env []string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := exec.CommandContext(ctx, shell)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Put the shell in its own process group and make it the terminal's foreground group
	// (Unix) so it can read stdin without blocking on SIGTTIN. ws does not reap the
	// shell's background jobs — that lifecycle belongs to the shell, which warns on exit
	// if jobs are still running and SIGHUPs them.
	setSubshellProcessGroup(cmd)
	err := cmd.Run()
	// The subshell held the terminal foreground; reclaim it before ws returns so the
	// parent shell isn't handed a background terminal. Background jobs the user started
	// are the shell's to manage — zsh/bash SIGHUP their jobs on exit.
	restoreTerminalForeground()
	return err
}
