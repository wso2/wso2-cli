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
	"context"
	"sync"
	"time"

	"github.com/wso2/wso2-cli/internal/remoteconnect"
)

// tokenSkew is how long before its expiry a cached token is treated as spent.
const tokenSkew = 30 * time.Second

// Grant is one minted access token and when it stops working.
type Grant struct {
	Token     string
	ExpiresAt time.Time
}

// Tokens holds the token a session presents to the broker and the bridge.
//
// One exchange serves a whole connect (the links read, the resolve and every
// tunnel dial), and the renewal loop invalidates it before each renewal so a
// renewal always runs on a fresh exchange. The same token goes to the broker
// and to the wss:// bridge.
type Tokens struct {
	// Mint performs one RFC 8693 exchange for the cloud audience.
	Mint func(ctx context.Context) (Grant, error)
	// Now reads the clock; nil means time.Now.
	Now func() time.Time

	mu      sync.Mutex
	current Grant
}

// Token returns the current token, minting one when there is none or it is
// about to lapse.
func (t *Tokens) Token(ctx context.Context) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current.Token != "" && t.now().Add(tokenSkew).Before(t.current.ExpiresAt) {
		return t.current.Token, nil
	}
	grant, err := t.Mint(ctx)
	if err != nil {
		return "", err
	}
	t.current = grant
	return grant.Token, nil
}

// Invalidate drops the current token, so the next Token call exchanges again.
func (t *Tokens) Invalidate() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.current = Grant{}
}

func (t *Tokens) now() time.Time {
	if t.Now == nil {
		return time.Now()
	}
	return t.Now()
}

// renewalResolver exchanges afresh before each resolve it carries. It is what
// the renewal loop resolves through; connect resolves through the plain one.
type renewalResolver struct {
	inner  Resolver
	tokens *Tokens
}

func (r renewalResolver) Resolve(ctx context.Context, req remoteconnect.ResolveRequest) (*remoteconnect.ResolveResponse, error) {
	r.tokens.Invalidate()
	return r.inner.Resolve(ctx, req)
}
