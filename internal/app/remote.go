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

package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/wso2/wso2-cli/internal/auth"
	"github.com/wso2/wso2-cli/internal/auth/trustedhttp"
	"github.com/wso2/wso2-cli/internal/contexts"
	"github.com/wso2/wso2-cli/internal/modules"
	"github.com/wso2/wso2-cli/internal/remote"
	"github.com/wso2/wso2-cli/internal/version"
	"github.com/wso2/wso2-cli/sdk/problem"
)

const (
	remoteConnectUsage = "Run wso2 remote connect <workload.yaml> --env <environment> [--print-env] [--no-secrets | --show-secrets]."
	remoteRecovery     = "Run wso2 remote connect <workload.yaml> --env <environment> to tunnel a workload's dependencies to this machine."

	// remoteConnectScope and remoteReadSecretsScope are the two permissions the
	// broker checks. The code holds the how; the context holds the where.
	remoteConnectScope     = "cloud:remote-connect:connect"
	remoteReadSecretsScope = "cloud:remote-connect:read-secrets"

	// remoteCallTimeout bounds one call to the broker.
	remoteCallTimeout = 30 * time.Second
)

// remoteScopes are what every exchange for the remote product asks for.
func remoteScopes() []string { return []string{remoteConnectScope, remoteReadSecretsScope} }

// remoteCommand builds the wso2 remote tree. It is a verification client for
// remote connect, not a product module: a module cannot hold a terminal, an
// open tunnel or a renewing token (docs: research 05, option A).
func (s Shell) remoteCommand() *cobra.Command {
	command := &cobra.Command{
		Use:                   "remote <subcommand>",
		Short:                 "Reach a deployed environment's dependencies from this machine.",
		Long:                  remoteRecovery,
		DisableFlagsInUseLine: true,
		RunE: func(command *cobra.Command, args []string) error {
			if len(args) == 0 {
				return helpForBareFamily(command)
			}
			return problem.New(problem.CategoryUsage, "shell.unknown_command",
				fmt.Sprintf("%q is not a wso2 remote subcommand", args[0])).
				WithRecovery(remoteRecovery)
		},
	}
	command.AddCommand(s.remoteConnectCommand())
	return command
}

func (s Shell) remoteConnectCommand() *cobra.Command {
	var params remote.ConnectParams
	command := &cobra.Command{
		Use:   "connect <workload.yaml>",
		Short: "Tunnel a workload's dependencies to this machine and start a subshell.",
		Long: "Reads the workload file, merges its dependencies over the ones the environment's deployed release " +
			"already binds (the file wins), asks the broker to resolve them, opens a local port for each remote " +
			"dependency, and starts a subshell whose environment points at those ports. The session renews itself. " +
			"The project and component come from the file's spec.owner and the organization from your login.\n\n" +
			"The context needs a remote product: an exchanged product at the audience " + contexts.RemoteConnectAudience +
			" with the " + remoteConnectScope + " and " + remoteReadSecretsScope + " scopes, whose url is the broker's base URL. " +
			"Run wso2 login again after adding it, because the login must ask for those scopes.",
		Args: exactlyOneArgument("one workload file", remoteConnectUsage),
		RunE: func(command *cobra.Command, args []string) error {
			params.WorkloadPath = args[0]
			contextName := ""
			if flag := shellFlag(command, contextFlag); flag != nil {
				contextName = flag.Value.String()
			}
			return s.remoteConnect(contextName, params)
		},
	}
	f := command.Flags()
	f.StringVar(&params.Environment, "env", "", "The environment to resolve dependencies for.")
	f.BoolVar(&params.PrintEnv, "print-env", false,
		"Print the resolved env bindings and hold the tunnels open instead of starting a subshell.")
	f.BoolVar(&params.ShowSecrets, "show-secrets", false,
		"With --print-env, print fetched secret and configmap values in full instead of redacting them.")
	f.BoolVar(&params.NoSecrets, "no-secrets", false,
		"Fetch no secret- or configmap-backed value, so no credential enters the local process.")
	declareContextFlag(f)
	_ = command.MarkFlagRequired("env")
	command.MarkFlagsMutuallyExclusive("show-secrets", "no-secrets")
	return command
}

// remoteConnect runs one session against the selected context.
func (s Shell) remoteConnect(contextName string, params remote.ConnectParams) error {
	selected, err := s.selection(contextName)
	if err != nil {
		return err
	}
	product, err := remoteProduct(selected)
	if err != nil {
		return err
	}
	root, err := s.stateRoot()
	if err != nil {
		return err
	}

	tokens := &remote.Tokens{Mint: func(context.Context) (remote.Grant, error) {
		// A broker grants once, so each exchange gets its own: the one at
		// connect, and one per renewal.
		broker := &auth.Broker{
			Namespace: contexts.RemoteProduct,
			Capabilities: modules.Capabilities{
				AuthAudiences: []string{contexts.RemoteConnectAudience},
				AuthScopes:    remoteScopes(),
			},
			Selection: selected,
			StateRoot: root,
		}
		grant, err := broker.Acquire(auth.Request{Audience: contexts.RemoteConnectAudience, Scopes: remoteScopes()})
		if err != nil {
			return remote.Grant{}, err
		}
		return remote.Grant{Token: grant.Token, ExpiresAt: grant.ExpiresAt}, nil
	}}
	api := &remote.Broker{
		BaseURL:    product.Endpoint,
		Token:      tokens.Token,
		HTTPClient: trustedhttp.Client(&http.Client{Timeout: remoteCallTimeout}),
	}

	// Tunnels live for the session; Ctrl-C or SIGTERM tears everything down.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	err = remote.New(api, tokens, "wso2-cli/"+version.Shell()).Connect(ctx, params, s.Streams.Out)
	return remoteProblem(err)
}

// remoteProduct returns the context's remote product, refusing a context that
// does not record it as the cloud exchange the broker needs.
func remoteProduct(selected contexts.Selection) (contexts.Product, error) {
	product, recorded := selected.Identity.Products[contexts.RemoteProduct]
	switch {
	case !recorded:
		return contexts.Product{}, remoteProductRefusal(selected,
			"records no remote product")
	case product.Grant == nil || product.Grant.Kind != contexts.GrantExchange:
		return contexts.Product{}, remoteProductRefusal(selected,
			"records its remote product without the exchange grant")
	case product.Audience != contexts.RemoteConnectAudience:
		return contexts.Product{}, remoteProductRefusal(selected,
			fmt.Sprintf("records its remote product at an audience other than %s", contexts.RemoteConnectAudience))
	}
	return product, nil
}

func remoteProductRefusal(selected contexts.Selection, what string) error {
	return problem.New(problem.CategoryUsage, "remote.product_not_configured",
		fmt.Sprintf("the %q context %s", selected.Context.Name, what)).
		WithRecovery(fmt.Sprintf("Add products.remote to the context with wso2 context edit: url (the broker's base URL), "+
			"audience %s, scopes %s and %s, and grant kind exchange. Then run wso2 login again, because the login "+
			"must ask for those scopes.", contexts.RemoteConnectAudience, remoteConnectScope, remoteReadSecretsScope))
}

// remoteProblem turns a failure into the typed problem the shell reports: a
// refused exchange reads as the auth refusal it is, and a broker refusal names
// its next step.
func remoteProblem(err error) error {
	if err == nil {
		return nil
	}
	var denial auth.Denial
	if errors.As(err, &denial) {
		return denial.Reported()
	}
	var refused *remote.ResolveError
	if errors.As(err, &refused) {
		return problem.New(problem.CategoryProductService, "remote.broker_refused", refused.Error()).
			WithRecovery(refused.Reason())
	}
	return err
}
