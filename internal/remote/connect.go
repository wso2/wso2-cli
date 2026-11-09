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

// Package remote implements ws remote connect: tunnel one workload's
// dependencies to the developer's machine.
//
// The flow is adapted from the OpenChoreo CLI's `occ remote` (see the notices
// in the files that carry its code). What differs is the control plane: the
// broker answers the resolve call, the dependency list is the merge of the
// environment's deployed release and the workload file, and every call and the
// tunnel itself carry a token exchanged for the cloud audience.
package remote

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/wso2/wso2-cli/internal/remoteconnect"
)

// ConnectParams are the inputs to ws remote connect.
type ConnectParams struct {
	// WorkloadPath is the workload file. Its spec.owner names the project and
	// component; the organization comes from the token.
	WorkloadPath string
	// Environment is the environment to resolve dependencies for.
	Environment string
	// PrintEnv prints the resolved bindings and holds the tunnels open instead
	// of starting a subshell.
	PrintEnv bool
	// ShowSecrets prints fetched values in full under PrintEnv instead of
	// redacting them.
	ShowSecrets bool
	// NoSecrets fetches no secret- or configmap-backed value, so no credential
	// enters the local process.
	NoSecrets bool
}

// LinkReader reads what an environment's deployed release of a component binds.
type LinkReader interface {
	DeployedRelease(ctx context.Context, project, component, environment string) (LinkTargets, bool, error)
}

// Remote implements the connect flow.
type Remote struct {
	// Links feeds the dependency merge.
	Links LinkReader
	// Resolver answers the connect resolve.
	Resolver Resolver
	// RenewResolver answers each renewal. It is the same broker as Resolver,
	// behind a fresh exchange; nil means Resolver.
	RenewResolver Resolver
	// DialAgent opens one tunnel to a remote-agent, authorized by capability.
	DialAgent func(ctx context.Context, agent remoteconnect.AgentEndpoint, capability func() string) (tunnel, error)
	// RunShell starts the subshell with the given environment.
	RunShell func(ctx context.Context, env []string) error
	// ConfirmSecrets asks whether --print-env may print the named fetched
	// values in full.
	ConfirmSecrets func(out io.Writer, names []string) bool
}

func (d *Remote) renewResolver() Resolver {
	if d.RenewResolver != nil {
		return d.RenewResolver
	}
	return d.Resolver
}

// Connect resolves the workload's dependencies, opens a local listener per
// tunnellable target, renders the env bindings, and starts a subshell (or
// prints the env with PrintEnv). Tunnels live until the subshell exits or ctx
// is cancelled.
func (d *Remote) Connect(ctx context.Context, p ConnectParams, out io.Writer) error {
	if p.Environment == "" {
		return errors.New("--env is required")
	}
	data, err := os.ReadFile(p.WorkloadPath)
	if err != nil {
		return fmt.Errorf("read the workload file: %w", err)
	}
	workload, err := ParseWorkload(data)
	if err != nil {
		return fmt.Errorf("%s: %w", p.WorkloadPath, err)
	}

	release, found, err := d.Links.DeployedRelease(ctx, workload.Project, workload.Component, p.Environment)
	if err != nil {
		return fmt.Errorf("read the deployed links of %s/%s: %w", workload.Project, workload.Component, err)
	}
	if !found {
		fmt.Fprintf(out, "No deployed release of %s/%s in %s; using the dependencies in %s only.\n",
			workload.Project, workload.Component, p.Environment, p.WorkloadPath)
	}
	req := remoteconnect.ResolveRequest{
		Project:     workload.Project,
		Component:   workload.Component,
		Environment: p.Environment,
		Endpoints:   workload.Endpoints,
		Resources:   MergeLinks(release, workload.Resources),
		Purpose:     remoteconnect.PurposeConnect,
		SkipSecrets: p.NoSecrets,
	}
	if len(req.Endpoints) == 0 && len(req.Resources) == 0 {
		return errors.New("the workload declares no dependencies, and its deployed release binds none; nothing to connect")
	}

	s := &session{
		overrides: map[string]string{},
		sensitive: map[string]bool{},
		files:     newFileStore(),
	}
	// Runs on every return path, including the cancelled one Ctrl-C takes, so
	// credentials written to disk do not outlive the tunnels that fetched them.
	defer func() {
		for _, ln := range s.listeners {
			_ = ln.Close()
		}
		for _, tn := range s.tunnels {
			_ = tn.Close()
		}
		// Tunnels a renewal superseded stay open for the streams still running
		// on them, so the session closes them here.
		if s.unit != nil {
			for _, tn := range s.unit.retiredTunnels() {
				_ = tn.Close()
			}
		}
		s.files.cleanup()
	}()

	fmt.Fprintf(out, "Connecting to %s/%s (%s)...\n", workload.Project, workload.Component, p.Environment)
	if err := d.connectRemote(ctx, req, s, out, p.NoSecrets); err != nil {
		return err
	}
	if len(s.listeners) == 0 && len(s.overrides) == 0 {
		return errors.New("none of the declared dependencies could be connected; no session was started")
	}

	// Started once the workload is up, so a renewal cannot overlap the resolve
	// that established the session.
	renewCtx, stopRenewals := context.WithCancel(ctx)
	defer stopRenewals()
	if s.unit.needsRenewal() {
		go s.unit.renew(renewCtx, d.renewResolver())
	}

	if p.PrintEnv {
		// Explicit --show-secrets needs no prompt; without it, ask rather than
		// leaving the developer to guess why a binding has no value.
		show := p.ShowSecrets
		if names := sortedKeys(s.sensitive); !show && len(names) > 0 && d.ConfirmSecrets != nil {
			show = d.ConfirmSecrets(out, names)
		}
		printEnvBindings(out, s.overrides, s.sensitive, show)
		fmt.Fprintln(out, "\nTunnels open. The session renews automatically; press Ctrl-C to disconnect.")
		<-ctx.Done()
		return nil
	}

	fmt.Fprintln(out, "\nTunnels open. The session renews automatically; exit the shell to disconnect.")
	return d.RunShell(ctx, mergeEnv(os.Environ(), s.overrides))
}

// session is the mutable state one invocation accumulates.
type session struct {
	overrides map[string]string
	// sensitive names the env vars whose values were fetched from the data
	// plane, so --print-env can redact them.
	sensitive map[string]bool
	listeners []net.Listener
	tunnels   []tunnel
	files     *fileStore
	// unit holds the capability the streams are authorized by.
	unit *remoteUnit
}

// connectRemote resolves the dependencies and opens a local listener for each
// tunnellable target.
func (d *Remote) connectRemote(ctx context.Context, req remoteconnect.ResolveRequest, s *session, out io.Writer, noSecrets bool) error {
	resp, err := d.Resolver.Resolve(ctx, req)
	if err != nil {
		return err
	}

	// Per resource, the in-cluster address each of its addresses resolved to and
	// the local listener that now stands in for it.
	localAddrs := map[string][]addrSwap{}
	staged := map[string]string{}
	stagedSensitive := map[string]bool{}

	// Built before the tunnels, because each tunnel is handed unit.cap.
	unit := newRemoteUnit(req, resp, out, d.DialAgent)
	s.unit = unit
	agentTunnels := make(map[string]tunnel, len(resp.Agents))
	for id, agent := range resp.Agents {
		tn, terr := d.DialAgent(ctx, agent, unit.cap)
		if terr != nil {
			return terr
		}
		s.tunnels = append(s.tunnels, tn)
		agentTunnels[id] = tn
		unit.installTunnel(id, tn)
	}
	if len(resp.Targets) > 0 {
		// Each target's streams go to its own agent; same-namespace dependencies
		// share a tunnel. The unit resolves key -> tunnel per connection, so the
		// listener does not care which agent currently serves it.
		reporter := newStreamErrorReporter(out, unit)
		for _, t := range resp.Targets {
			if _, ok := agentTunnels[t.AgentID]; !ok {
				return fmt.Errorf("resolve returned no remote-agent %q for target %s", t.AgentID, t.Key)
			}
			ln, lerr := net.Listen("tcp", net.JoinHostPort(localHost, "0"))
			if lerr != nil {
				return fmt.Errorf("open local listener for %s: %w", t.Key, lerr)
			}
			s.listeners = append(s.listeners, ln)
			port := ln.Addr().(*net.TCPAddr).Port

			key := t.Key
			open := func() (net.Conn, error) { return unit.openStream(key) }
			go forward(ln, key, open, reporter.report)

			mergeOverrides(staged, out, remoteconnect.RenderEnv(t, localHost, port))
			if t.Resource != nil && t.Resource.RemoteAddr != "" {
				localAddrs[t.Resource.Ref] = append(localAddrs[t.Resource.Ref],
					newAddrSwap(t.Resource.RemoteAddr, strconv.Itoa(port)))
			}
			fmt.Fprintf(out, "  %-28s -> %s:%d  (%s)\n", t.Key, localHost, port, targetKind(t))
		}
	}
	applyResourceBindings(staged, out, resp, localAddrs)
	// After the tunnels: a fetched value travels over one.
	materialized := fetchBindings(staged, stagedSensitive, out, resp, agentTunnels, s.files, localAddrs, noSecrets)
	// After both merges: an env var naming a mount path may have come from
	// StaticEnv or from a fetch, so the repoint must see the finished map.
	repointFilePaths(staged, out, s.files, materialized)
	for _, u := range resp.Unconnectable {
		fmt.Fprintf(out, "  ! %s: %s\n", u.Ref, u.Reason)
	}

	mergeOverrides(s.overrides, out, staged)
	for name := range stagedSensitive {
		s.sensitive[name] = true
	}
	return nil
}
