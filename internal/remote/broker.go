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

package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/wso2/wso2-cli/internal/remoteconnect"
)

// CallTimeout bounds one call to the broker; the shell builds its HTTP client with it too.
const CallTimeout = 30 * time.Second

// responseLimit bounds what is read of a broker answer.
const responseLimit = 4 << 20

// Resolver turns a workload's declared dependencies into connection targets and
// a signed capability. Connect and renew are the same call.
type Resolver interface {
	Resolve(ctx context.Context, req remoteconnect.ResolveRequest) (*remoteconnect.ResolveResponse, error)
}

// Broker is the client of the cloud remote-connect broker: the resolve call and
// the links read that feeds the dependency merge. It is the only thing here that
// knows a URL, and the URL is the context's (products.remote.url).
type Broker struct {
	// BaseURL is the Resources API base the broker is published under.
	BaseURL string
	// Token returns the bearer for one call. A renewing session outlives any
	// one token, so it is asked per call.
	Token func(ctx context.Context) (string, error)
	// HTTPClient is the transport; nil means a client with CallTimeout.
	HTTPClient *http.Client
}

// resolveBody is the broker's request: the org comes from the token, the
// project from the path, and the purpose is the broker's own to set.
type resolveBody struct {
	Component          string                      `json:"component"`
	Environment        string                      `json:"environment"`
	Endpoints          []remoteconnect.EndpointDep `json:"endpoints,omitempty"`
	Resources          []remoteconnect.ResourceDep `json:"resources,omitempty"`
	SkipSecrets        bool                        `json:"skipSecrets,omitempty"`
	PreviousCapability string                      `json:"previousCapability,omitempty"`
}

// Resolve calls the broker's resolve operation for req.Project.
func (b *Broker) Resolve(ctx context.Context, req remoteconnect.ResolveRequest) (*remoteconnect.ResolveResponse, error) {
	payload, err := json.Marshal(resolveBody{
		Component: req.Component, Environment: req.Environment,
		Endpoints: req.Endpoints, Resources: req.Resources,
		SkipSecrets: req.SkipSecrets, PreviousCapability: req.PreviousCapability,
	})
	if err != nil {
		return nil, fmt.Errorf("encode the resolve request: %w", err)
	}
	var out remoteconnect.ResolveResponse
	err = b.call(ctx, http.MethodPost, b.endpoint("projects", req.Project, "remote-connect", "resolve"), payload, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// linksResponse is the part of the links read this uses.
type linksResponse struct {
	Environments []struct {
		Name     string      `json:"name"`
		Deployed bool        `json:"deployed"`
		Release  LinkTargets `json:"release"`
	} `json:"environments"`
}

// DeployedRelease reads what the environment's deployed release binds for the
// component. found is false for a component the platform has not seen (404) and
// for an environment with no deployed release; both mean the workload file is
// the whole dependency list.
func (b *Broker) DeployedRelease(ctx context.Context, project, component, environment string) (LinkTargets, bool, error) {
	var links linksResponse
	err := b.call(ctx, http.MethodGet, b.endpoint("links", project, component), nil, &links)
	var refused *BrokerError
	if errors.As(err, &refused) && refused.Status == http.StatusNotFound {
		return LinkTargets{}, false, nil
	}
	if err != nil {
		return LinkTargets{}, false, err
	}
	for _, env := range links.Environments {
		if env.Name == environment && env.Deployed {
			return env.Release, len(env.Release.Env)+len(env.Release.Files) > 0, nil
		}
	}
	return LinkTargets{}, false, nil
}

func (b *Broker) endpoint(segments ...string) string {
	escaped := make([]string, len(segments))
	for i, segment := range segments {
		escaped[i] = url.PathEscape(segment)
	}
	return strings.TrimRight(b.BaseURL, "/") + "/" + strings.Join(escaped, "/")
}

func (b *Broker) call(ctx context.Context, method, target string, payload []byte, into any) error {
	token, err := b.Token(ctx)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, method, target, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	client := b.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: CallTimeout}
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("call the remote-connect broker: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(response.Body, responseLimit))
	if err != nil {
		return fmt.Errorf("read the broker's answer: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return newBrokerError(response.StatusCode, body)
	}
	if err := json.Unmarshal(body, into); err != nil {
		return fmt.Errorf("decode the broker's answer: %w", err)
	}
	return nil
}
