# AGENTS.md

## Build and Development Commands

```bash
make build                    # Build CLI binary → bin/wso2-integration-platform (packages translations first)
make compile_ip               # Build branded binary → bin/{OS}-{ARCH}/wso2-integration-platform
make run                      # Run CLI locally via `go run`
make test                     # Run unit tests (excludes integration_tests/, 1200s timeout)
make run_integration_tests    # Run integration tests (1500s timeout)
make run_integration_tests TEST="TestName"  # Run specific integration test
make extract_tools_defs       # Generate tools_summary.csv from MCP tool definitions
make clean                    # Remove build artifacts

# Cross-platform compilation (branded binary)
OS=linux  ARCH=amd64 make compile_ip   # → bin/linux-amd64/wso2-integration-platform
OS=darwin ARCH=arm64 make compile_ip   # → bin/darwin-arm64/wso2-integration-platform
OS=windows ARCH=amd64 make compile_ip  # → bin/windows-amd64/wso2-integration-platform.exe

# mcpb Claude Connector bundles
OS=linux  ARCH=amd64 make bundle_mcpb  # → bin/linux-amd64/wso2-integration-platform-linux-amd64.mcpb
OS=darwin ARCH=arm64 make bundle_mcpb  # → bin/darwin-arm64/wso2-integration-platform-darwin-arm64.mcpb
OS=windows ARCH=amd64 make bundle_mcpb # → bin/windows-amd64/wso2-integration-platform-windows-amd64.mcpb
```

Translation packaging (`make package_translations`) runs automatically before `build`, `run`, and `compile_ip`.
If you add or modify i18n strings, translations must be regenerated with `make gen_translations` first.

**Build prerequisite:** `go-bindata` must be installed for translation packaging:
```bash
go install github.com/go-bindata/go-bindata/...@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

**Build verification:** `go build ./...` (or `make compile_ip`) must exit 0 before committing. The binary is ~17–18 MB (stripped via `-s -w` in `LDFLAGS`); if it comes out ~5 MB or less, the wrong entry point was built (use `go build cmd/main.go`, not `go build .`). Also confirm `--version` does **not** print `DEV` — that means `$(LDFLAGS)` was dropped or the `-X` symbol path no longer matches the module path, and `go build` silently ignores an unknown `-X` symbol.

**Regenerating all release artifacts** (note: `OS`/`ARCH` must be set per invocation; a shell loop over `"linux amd64"`-style strings will not word-split under zsh):
```bash
make clean
OS=linux   ARCH=amd64 make compile_ip
OS=linux   ARCH=arm64 make compile_ip
OS=darwin  ARCH=arm64 make compile_ip
OS=darwin  ARCH=amd64 make compile_ip
OS=windows ARCH=amd64 make compile_ip
OS=linux   ARCH=amd64 make bundle_mcpb
OS=darwin  ARCH=arm64 make bundle_mcpb
OS=windows ARCH=amd64 make bundle_mcpb
```

## Architecture Overview

Go CLI for the **WSO2 Integration Platform**. Supports multi-region deployment (US/EU), OAuth2 authentication, PAT authentication, and an MCP server mode for AI assistant integration. The binary is named `wso2-integration-platform`.

### Binary Names

| Target | Binary | Purpose |
|--------|--------|---------|
| `make build` | `bin/wso2-integration-platform` | Local/CI build (version-stamped) |
| `make compile_ip` | `bin/{OS}-{ARCH}/wso2-integration-platform` | Branded release binary |

### Command Structure

Commands follow a **four-file pattern** in `internal/cmd/{domain}/{operation}/`:

| File | Purpose |
|------|---------|
| `cmd.go` | Cobra command definition, flags in `init()`, calls handler from `impl.go` |
| `impl.go` | Business logic handler (e.g., `HandleProjectCreate`) |
| `models.go` | Parameter structs for flag binding |
| `util.go` | Formatting/processing helpers specific to that command |

Top-level commands (`create`, `delete`, `list`, `describe`, `suspend`, `resume`) are defined in `internal/cmd/` and aggregate domain subcommands. Adding/removing root-level commands requires manually updating the help template in `root.go` (see comment on line 173).

**Common flag helpers** in `internal/cmd/common/` ensure consistency:
- `AddOrgFlag()`, `AddProjectFlag()`, `AddDescriptionFlag()`, `AddGitCredentialFlag()`
- `AddGenericHelper()` — custom help formatting for commands
- `VerifyIsUserLoggedIn` — used as `PreRun` hook on authenticated commands

### MCP Server

Hidden command `start-mcp-server` (supports `--http` and `--port` flags) exposes CLI operations as MCP tools using `mark3labs/mcp-go`.

Each MCP domain in `internal/mcp/{domain}/` has three files:
- `tools.go` — tool definitions with `mcp.NewTool()`, including hint annotations (ReadOnly, Destructive, Idempotent)
- `impl.go` — tool handler implementations
- `mcp.go` — `Register{Domain}Tools(s *server.MCPServer)` function

New MCP tools must be registered in `internal/mcp/cmd.go` via the domain's `Register` function.

**MCP stdio transport:** The server uses newline-delimited JSON-RPC. When testing manually, keep stdin open (use a named pipe, not a file) — the server exits when stdin closes (EOF), so responses to long-running tool calls are lost if stdin closes before the API call returns. The MCP protocol requires an `initialize` request followed by a `notifications/initialized` notification before tools can be called.

**mcpb format (Claude Connector bundle):** A zip file containing:
- `manifest.json` — server metadata and launch config
- `server/{binary}` — the platform binary

The `mcp_config.command` field **must** use the `${__dirname}` template variable (not a hardcoded path) — the Claude Desktop host replaces it with the extraction directory at runtime:
```json
"command": "${__dirname}/server/wso2-integration-platform"
```
Platform-specific manifests: `mcpb/manifest.json` (Linux), `mcpb/manifest-darwin.json`, `mcpb/manifest-windows.json`.

### API Client Layer

`pkg/api/` contains GraphQL and REST clients organized by service domain. The shared HTTP client (`IPHTTPClient`) provides retry logic (3 attempts, 3s delay), auth headers, and optional trace logging (`TRACE_ENABLED=true` env var).

`DoForActiveOrg()` on `IPHTTPClient` calls `GetTokenForActiveOrg()` to obtain the bearer token for the stored active org. **Do not call `GetSelectedOrganization()` from inside `GetTokenForActiveOrg()`** — that method calls `OrgClient.GetOrganizations()` → `DoForActiveOrg()` → `GetTokenForActiveOrg()` (infinite loop). Instead use `orgStore.GetDefaultOrg()` (pure keyring read, no API call). This pattern is in `internal/auth/ro-token-store.go`.

### Multi-Region Support

- Region configs: `internal/region/us_config.go`, `eu_config.go`, `shared_config.go`
- Thread-safe region selection with mutex; persisted in keyring
- Tokens keyed by `{region}_{integerOrgId}` (not UUID, not just orgId)
- Each region has dev/stage/prod environments with distinct API endpoints
- `GetCurrentRegion()` checks `WSO2IP_REGION` env var first, then keyring, then defaults to `"US"`
- `GetRegionConfig()` checks `WSO2IP_ENV` env var for dev/stage override; always returns prod for end users
- Non-prod (dev/stage) endpoints are **not compiled in**. `WSO2IP_ENV=dev|stage` also requires
  `WSO2IP_ENV_CONFIG=<path.json>`, keyed by region then environment, whose fields mirror
  `BuildEnvConfig`'s arguments (`internal/region/nonprod_config.go`). Requesting an
  environment the file does not describe panics rather than falling back to prod — a silent
  fall-through would point a developer's commands at the live platform.
- Region selection lives in `internal/region/config.go` and `store.go`; token storage in `internal/auth/token-store.go`

### Authentication

`internal/auth/` implements three authentication flows:

1. **OAuth2 with PKCE** (`sign-in.go`) — browser-based interactive login
2. **PAT login** (`token-login.go: LoginWithToken`) — Personal Access Token; reads `WSO2IP_PAT` env var; calls PAT introspect endpoint then org list to bootstrap session
3. **STS token login** (`token-login.go: LoginWithSTSToken`) — bootstraps from a cloud STS token; parses JWT claims directly (no API call needed for org handle), fetches org details via `makeRequest` (not `OrgClient`), then attempts `ExchangeSTSToken` to get a refreshable token; falls back to storing the STS token directly if exchange fails

Tokens are AES-256 encrypted at rest in `~/.wso2/.auth/` (`CLI_HOME_DIR` = `.wso2`, `CLI_HOME_AUTH_DIR` = `.auth` in `pkg/util/constants/constants.go`), with encryption keys stored in the system keyring via `zalando/go-keyring`.

**Environment variables:**

| Variable | Purpose |
|----------|---------|
| `WSO2IP_PAT` | Personal Access Token — skips interactive login |
| `WSO2IP_REGION` | Force region (`US` or `EU`) |
| `WSO2IP_ENV` | Force environment (`dev` or `stage`); end users never set this |
| `CLOUD_STS_TOKEN` | Bootstrap session from VSCode/cloud extension STS token |
| `CLOUD_INITIAL_ORG_ID` | Pre-select org on first launch (cloud extension context) |
| `CLOUD_INITIAL_PROJECT_ID` | Pre-select project on first launch (cloud extension context) |
| `TRACE_ENABLED` | Set to `true` to log all HTTP requests/responses |
| `OUTPUT_FORMAT` | Default output format (`table` or `json`) |

### I/O and Output

All CLI output uses `utils.IO.Out` and `utils.IO.ErrOut` (from `cli/cli/v2/pkg/iostreams`), not direct `fmt.Print`. This enables output capture in tests. Errors use `utils.HandleErr()`, info messages use `utils.PrintInfo()`, and tables use `utils.CreateTable()`.

MCP tool responses use `utils.NewMCPResponse(data, conclusion, nextSteps)` for success and `utils.NewMCPErrorResponse(err, conclusion)` for errors. Both return a structured JSON payload with `data`, `error`, `conclusion`, and `next_steps` fields.

### i18n

All user-facing strings must be wrapped in `i18n.T("...")`. Translation files live in `i18n/translations/` (JSON format). The build process generates and embeds translation assets via `scripts/translations/`.

## Component vs. Integration

What users see is an **integration**; what the code and the platform API call it
is a **component**. The two never have to agree, and the split is deliberate:

| Layer | Wording | Why |
|---|---|---|
| CLI commands, flags, help, prompts, table headers, error text | integration | What the product calls it |
| MCP tool names and parameters (`create_integration`, `integration_uuid`) | integration | Agent-facing surface |
| Go identifiers, package paths (`internal/cmd/component`, `pkg/api/component`, `models.Component`) | component | Matches the GraphQL/REST schema it maps to |
| JSON/YAML tags, URL paths, GraphQL queries, `component.yaml` | component | Wire format and files on disk |

So `common.AddComponentFlag` registers a flag named `--integration`, and
`internal/cmd/component/list` implements `list integrations`. Renaming a Go
identifier to `Integration` only makes sense if the API type it mirrors is
renamed too.

The pre-rename spellings stay reachable: `create`/`list`/`describe`/`delete`
accept `component` and `components` as hidden subcommands registered through
`common.LegacyAliasCommand`, and `--component`/`-c` is a hidden deprecated alias
of `--integration`/`-i`. Both print a deprecation notice on stderr. Two places
keep the old word on purpose — the `components` key in `describe project
--output=json`, and the bridge component `connect` creates on the platform.

New user-facing strings say "integration". When you change one, remember the
matching entry in `i18n/translations/all.en_US.json` (id and translation both);
a missing entry falls back to the English source string, so a stale file is not
a build failure, just drift.

## Brand and Identifier Policy

This codebase is open-sourced as the **WSO2 Integration Platform CLI**. Go identifiers (variable names, struct fields, package names, function names, comments) must not contain Choreo brand references. Backend-enforced strings are intentionally left unchanged.

**Renamed (Go identifiers):**
- Package `choreoGit` → `platformgit` (in `pkg/api/platformgit/`)
- `AuthClientConfig` fields: `ChoreoSTSClientId/TokenUrl/Scopes` → `STSClientId/TokenUrl/Scopes`
- `ExchangeChoreoSTSToken` → `ExchangeSTSToken`
- `ChoreoPlatformHostname` → `PlatformHostname`
- Region config vars: `DEFAULT_CHOREO_ENV_CONFIG` → `DEFAULT_ENV_CONFIG`, `EU_CHOREO_*` → `EU_*`
- Model fields: `ChoreoEnv` → `Env` (JSON tags preserved)
- `GetChoreoSamples` → `GetSamples`; `ChoreoConsoleUrl` → `ConsoleUrl`
- Local config directory: `.choreo/` → `.wso2/`

**Also renamed in the September 2026 pass** (no backend dependency, so these were ours to change):
- Test env vars `CHOREO_CLI_TEST_*` → `WSO2IP_TEST_*` (code and `.github/actions/e2e-test-run`)
- npm keyword `choreo` → `devant` (`scripts/mcp-npm-package/package.json`)
- Stale i18n exclusions `vscode.choreo.ext` and `CHOREO_ENV` (the code emits `vscode.wso2ip.ext` and reads `WSO2IP_ENV`)
- `.gitignore` guards for the pre-rename `/choreo` binary
- Documentation links repointed to `https://wso2.com/integration-platform/docs/`. Not a base
  swap: the new site has a different taxonomy and every old `develop-components/*` path 404s
  under it, while the old choreo paths still resolve — so a host-only rewrite silently breaks
  them. The WSO2 Cloud pages live under `manage/cloud/`, found via the site's `sitemap.xml`
  (1076 pages), which is the way to locate a replacement rather than guessing from the nav:
  endpoints → `manage/cloud/configurations/endpoint-configurations`, git credentials →
  `manage/cloud/cicd/connect-git-repository`, connections →
  `develop/integration-artifacts/supporting/connections`. Verify any new link returns 200.

**Left unchanged, and why.** A grep for "choreo" still returns ~119 hits. Every one is in a
category where changing it breaks something; none are Go identifiers (that grep returns zero):

| Category | Why it cannot change |
|---|---|
| Production URLs — `choreo.dev`, `sts.choreo.dev`, `.choreoapps.dev`, `.choreoapis.dev` | The live endpoints the CLI calls |
| OAuth scopes `choreo:*` | Issued and validated by the STS |
| API paths — `/choreo-connections`, `/choreo-apis`, `choreologgingapi`, `choreoobsapi`, `urn:choreosystem:*` | Routed by the gateway |
| JSON tags `choreoEnv` / `choreo_env` | GraphQL/REST wire schema |
| `"choreo-local-bridge"` | Component and image name looked up on the platform |
| `"choreo-cache"` | Value the database API accepts (the agent-facing text around it was softened) |
| `"choreo"`, `cloudType=choreo` | Service identifiers sent to user-mgt and subscriptions |
| `choreoanonymouspullable.azurecr.io` | The actual registry host |
| `.choreo/` in `.gitignore` | Guards a developer's pre-rebrand credential directory from being committed |
| `workspaces/apps/choreo-console/...` in comments | Accurate path into the console repo, cited as the source of truth for integration filtering |

## Versioning

`internal/cmd.Version` is stamped at build time via `LDFLAGS`. It defaults to
`"DEV"` in source (`internal/cmd/root.go`), and a `"DEV"` value also disables
the update checker.

| Build | VERSION | Stamped as |
|-------|---------|-----------|
| Release (`release.yml` passes the tag) | `v1.0.0` | `v1.0.0` — verbatim |
| Local / dev (VERSION unset) | derived | `<git describe>-<YYYYMMDD>` |

Release builds must stamp the bare tag. Do not re-append the date in `LDFLAGS`
— that was the reason `updatechk` had to strip a suffix back off.

`bundle_mcpb` substitutes `MCPB_VERSION` (the tag without its leading `v`) into
the manifest's `version` field. The value checked into `mcpb/manifest*.json` is
a placeholder; do not rely on it. The MCP host uses that field to detect
connector updates, so a bundle built without a `VERSION=` will advertise the
placeholder.

**Version baseline:** the project reset to `v1.0.0` when it was renamed to
`integration-platform-tools` and open-sourced. The 232 pre-reset tags used a
scheme that concatenated a version and a timestamp with no separator
(`v1.2.21` + `2606121500` → `v1.2.212606121500`), which parses as
patch `212606121500` and numerically outranks any `v1.x`. The reset is therefore
only valid on fresh history — carrying those tags forward would make every
`v1.0.0` user see a bogus "update available", and would block npm publishes
below `1.2.x`. `internal/cmd/updatechk/checker_test.go` pins this behaviour.

Do not reintroduce the concatenated scheme. Use plain semver tags
(`v1.2.3`), with build metadata after a separator if needed (`v1.2.3+2606121500`).

## Development Flow

- Feature branches from `dev` branch
- Bug fixes from `main` branch
- Bi-weekly releases (every second Thursday): `dev` merges to `main`

## Testing

### Unit Tests

Run with `make test`. Located alongside source files (not in a separate directory).

**Current coverage by package:**

| Package | What's tested |
|---------|--------------|
| `internal/auth` | Org/token/user keyring stores; OAuth callback server; PAT env var (invalid/missing); `parseSTSTokenClaims` (7 cases) |
| `internal/cmd/common` | `ValidateNameText`, `ParseKeyValStrPair`, `ParseOutputFormat`, `IsStructured`, `RenderStructured` |
| `internal/cmd/component/create` | `GetComponentKindForCreate` auto-build/deploy defaults |
| `internal/mcp/integration` | `ResolveSubtype` all 6 types; rejection of unsupported values |
| `internal/mcp/utils` | `NewMCPResponse`, `NewMCPErrorResponse` structure and IsError flag |
| `internal/region` | `GetCurrentRegion` env/keyring/default; `SetCurrentRegion` valid/invalid; `GetConfigByRegion` prod/dev/stage |
| `pkg/api/component` | `IsIntegrationDisplayType` all display type variants |

**Not yet covered by unit tests (needs live credentials or complex mocking):**
- `LoginWithToken`, `LoginWithSTSToken` (require real API)
- `GetTokenForActiveOrg` / `DoForActiveOrg` token refresh path
- All API client methods in `pkg/api/`
- All MCP tool handler `impl.go` files
- All CLI command handler `impl.go` files

**Test patterns to follow:**
- Tests in the same package as the source (no `_test` package suffix needed for internal functions)
- Use `github.com/stretchr/testify/assert` and `require`
- Keyring-backed stores work on Linux CI; `internal/auth/store_test.go` has a `//go:build !linux` constraint as an exception
- Use `t.Helper()` in test helper functions
- For pure functions, use table-driven tests with a `tests []struct{...}` slice

### Integration Tests

Located in `integration_tests/`. Require environment variables:
- `WSO2IP_TEST_USER_NAME`, `WSO2IP_TEST_USER_PASS` — test credentials
- `RUN_OS` — platform identifier (appended to test project names)
- `WSO2IP_ENV` — target environment (read at `integration_tests/shared_test.go:38`)
- All tests run in non-interactive mode (`config.NonInteractive = true`)

The E2E flow (`TestE2EMainFlow`) covers: login via browser automation (Chromium headless), project create/delete, component create/list, build trigger + polling until success, deployment trigger.

**PAT-based integration tests** (`integration_tests/pat_env_test.go`) currently only cover invalid/missing token cases. Testing a valid PAT round-trip requires setting `WSO2IP_PAT` to a real token.
