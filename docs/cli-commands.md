# CLI Command Reference

Every command the `wso2-integration-platform` binary currently supports, with its
arguments, flags and aliases. Run `wso2-integration-platform <command> --help` for
the same information from the binary itself.

- [Conventions](#conventions)
- [Global flags](#global-flags)
- [Common resource flags](#common-resource-flags)
- [Legacy `component` spellings](#legacy-component-spellings)
- [Authentication and session](#authentication-and-session)
- [create](#create)
- [list](#list)
- [describe](#describe)
- [delete](#delete)
- [suspend / resume](#suspend--resume)
- [logs](#logs)
- [connect](#connect)
- [Hidden commands](#hidden-commands)
- [Environment variables](#environment-variables)

---

## Conventions

- `<value>` is a required value, `[value]` an optional one.
- Most resource commands accept the resource name as a positional argument *or*
  as a flag; omitting it opens an interactive prompt unless `--non-interactive`
  is set.
- Commands that reach the platform require a session — run `login` first.

## Global flags

| Flag | Description |
|------|-------------|
| `--non-interactive` | Disable interactive prompts and fail when a required parameter is missing. Defaults to `true` when `CI=true`. |
| `-h`, `--help` | Help for the command. |
| `-v`, `--version` | Version of the CLI (root command only). |

## Common resource flags

These flags are shared by most resource commands and are not repeated in full below.

| Flag | Description |
|------|-------------|
| `--org` | Organization name, ID, or handle. |
| `-p`, `--project` | Project ID, name, or handle. |
| `-i`, `--integration` | Integration name, ID, or handle. |
| `-e`, `--env` | Environment name (`Development`, `Production`, …). |
| `-d`, `--deployment-track` | Deployment track, e.g. `main`. |
| `-o`, `--output` | Output format: `table` or `json`. Defaults to `$OUTPUT_FORMAT`, else `table`. |

---

## Legacy `component` spellings

Integrations were previously called components. The old spellings still work so
existing scripts keep running, but they are deprecated, hidden from `--help`,
and print a one-line notice on stderr:

| Deprecated | Use instead |
|------------|-------------|
| `create component` / `create components` | `create integration` |
| `list components` / `list component` | `list integrations` |
| `describe component` / `describe components` | `describe integration` |
| `delete component` / `delete components` | `delete integration` |
| `--component`, `-c` (selector flag) | `--integration`, `-i` |
| `-c` on `create integration` (i.e. `--component-name`) | `-n`, `--name` |

Two places still say `component`, because changing them would break something
outside the CLI: the `components` key in `describe project --output=json` (the
machine-readable contract), and the bridge component that `connect` creates on
the platform.

---

## Authentication and session

### `login` (alias: `signin`)

Authenticate with your Integration Platform account.

```bash
wso2-integration-platform login
wso2-integration-platform login --with-token   # token read from stdin
```

| Flag | Description |
|------|-------------|
| `--org` | Organization handle. |
| `--user` | User email for the login. |
| `--with-token` | Authenticate with a token read from standard input. |

A Personal Access Token in `WSO2IP_PAT` skips interactive login entirely.

### `logout` (alias: `signout`)

Ends the current session. No flags beyond the global ones.

### `change-org`

Switch the active organization for your local context.

```bash
wso2-integration-platform change-org
wso2-integration-platform change-org --org=<org-name>
```

| Flag | Description |
|------|-------------|
| `-o`, `--org` | Name or handle of the organization to select. |

### `region [region-name]`

View or set the region. With no argument it prints the current region. Valid
values are `US` and `EU`. Switching regions may require logging in again.

```bash
wso2-integration-platform region        # view
wso2-integration-platform region EU     # switch
```

### `set-context`

Associate the current repository directory with a project, so subsequent
commands in that directory pick up the organization and project automatically.

| Flag | Description |
|------|-------------|
| `--org` | Organization name, ID, or handle. |
| `-p`, `--project` | Project ID, name, or handle. |

### `env`

List every environment variable the CLI reads, with its current value, built-in
default and description. Sensitive values such as `WSO2IP_PAT` are masked — only
whether they are set is reported.

| Flag | Description |
|------|-------------|
| `-o`, `--output` | `table` or `json`. |

---

## create

Create a resource: `wso2-integration-platform create <resource>`.

### `create project <project-name>`

| Flag | Description |
|------|-------------|
| `--description` | Project description. |
| `--org` | Organization name, ID, or handle. |

```bash
wso2-integration-platform create project <project-name> --description <description>
```

### `create integration [integration-name]` (alias: `integrations`)

Create a new integration. Supported types: `service` (API, AI Agent,
MCP Server), `scheduleTask` (Automation), `eventHandler` (Event Integration,
File Integration). Buildpacks: `ballerina`, `microintegrator`.

| Flag | Description |
|------|-------------|
| `-n`, `--name` | Name of the new integration. |
| `-t`, `--type` | `service`, `scheduleTask`, or `eventHandler`. |
| `--build-pack` | `ballerina` or `microintegrator`. |
| `-r`, `--repo` | Git repository URL (multi-repo projects only). |
| `-b`, `--repo-branch` | Git branch (multi-repo projects only). |
| `--dir` | Subpath of the directory containing the integration. |
| `--git-credential` | Git credential name. |
| `--build-configs` | Build configurations as `key=value` pairs, comma-separated or repeated. |
| `--auto-build` | Trigger a build after creation (default `true`). |
| `--auto-deploy` | Deploy after a successful build (default `true`). |
| `--org`, `-p`/`--project` | Common resource flags. |

```bash
wso2-integration-platform create integration <name> --project='Default Project' \
  --type=service --build-pack='ballerina' --repo=<repo-url> --repo-branch=main
```

### `create build [integration-name]`

Trigger a build for an integration on a deployment track.

| Flag | Description |
|------|-------------|
| `--org`, `-p`/`--project`, `-d`/`--deployment-track` | Common resource flags. |

```bash
wso2-integration-platform create build <integration-name> --project='Default Project' --deployment-track=main
```

### `create deployment [integration-name]`

Trigger a deployment for an integration in an environment.

| Flag | Description |
|------|-------------|
| `--build-id` | Build ID to deploy (default `0`). |
| `--env-vars` | Environment variables, e.g. `--env-vars="key1=val1,key2=val2"`. |
| `--cron-expression` | Schedule expression for the cron job (scheduled tasks only). |
| `--cron-timezone` | Schedule timezone for the cron job (scheduled tasks only). |
| `--byoi-image` | Deployment image with tag (image-based integrations only). |
| `--byoi-endpoints-file` | Service endpoints file path (image-based integrations only). |
| `--byoi-api-schema-file` | API schema file path (image-based integrations only, repeatable). |
| `--proxy-target-url` | Target endpoint for the proxy (git-based proxies only). |
| `--proxy-sandbox-url` | Sandbox endpoint for the proxy (git-based proxies only). |
| `--org`, `-p`/`--project`, `-e`/`--env`, `-d`/`--deployment-track` | Common resource flags. |

```bash
wso2-integration-platform create deployment <integration-name> --project=<project-name> \
  --deployment-track=main --env=Development
```

### `create config`

Create a config-map or secret for an integration.

| Flag | Description |
|------|-------------|
| `-n`, `--name` | Name of the new config. |
| `-t`, `--type` | `config-map` or `secret`. |
| `--mount-type` | `env-variables` or `file-mount`. |
| `--env-vars` | Environment variables, e.g. `--env-vars="key1=val1,key2=val2"`. |
| `--mount-path` | Mount path of the config file. |
| `--mount-content` | Content of the file to be mounted. |
| `--org`, `-p`/`--project`, `-i`/`--integration`, `-e`/`--env`, `-d`/`--deployment-track` | Common resource flags. |

```bash
wso2-integration-platform create config --project=<project> --integration=<integration> \
  --type=config-map --mount-type=env-variables --name=<config-name> --env-vars=key1=val1,key2=val2
```

### `create connection`

Create a connection to a service.

| Flag | Description |
|------|-------------|
| `-n`, `--name` | Name of the new connection. |
| `-s`, `--service` | Service the connection is created for. |
| `--org`, `-p`/`--project`, `-i`/`--integration` | Common resource flags. |

### `create execution`

Trigger an execution of a scheduled or manual task.

| Flag | Description |
|------|-------------|
| `--org`, `-p`/`--project`, `-i`/`--integration`, `-e`/`--env`, `-d`/`--deployment-track` | Common resource flags. |

### `create test-key`

Get the test key needed to invoke a publicly deployed service.

| Flag | Description |
|------|-------------|
| `--endpoint` | Name of the endpoint. |
| `--org`, `-p`/`--project`, `-i`/`--integration`, `-e`/`--env`, `-d`/`--deployment-track` | Common resource flags. |

---

## list

Alias: `ls`. Every `list` subcommand supports `-o`/`--output` (`table`, `json`)
and accepts the singular form as an alias (`list project`, `list build`, …).

| Command | Lists | Extra flags |
|---------|-------|-------------|
| `list organizations` (`org`, `orgs`, `organization`) | Organizations you belong to | — |
| `list projects` | Projects in an organization | `--org` |
| `list integrations` (`integration`) | Integrations in a project | `--org`, `-p`/`--project` |
| `list builds` | Builds of an integration | `--org`, `-p`/`--project`, `-i`/`--integration`, `-d`/`--deployment-track` |
| `list configs` | Config-maps and secrets of an integration | `--org`, `-p`/`--project`, `-i`/`--integration`, `-e`/`--env`, `-d`/`--deployment-track` |
| `list connections` | Connections in a project | `--org`, `-p`/`--project`, `-i`/`--integration` |
| `list execution` (`executions`) | Executions of a scheduled/manual task | `--limit` (default `50`), `--org`, `-p`/`--project`, `-i`/`--integration`, `-e`/`--env`, `-d`/`--deployment-track` |

`list integrations` shows integrations only — automations, APIs, AI agents, MCP
servers, event and file integrations. Other component types in the same project
are filtered out.

```bash
wso2-integration-platform list projects --output=json > projects.json
```

---

## describe

Alias: `desc`. Every `describe` subcommand supports `-o`/`--output`
(`table`, `json`).

| Command | Shows | Extra flags |
|---------|-------|-------------|
| `describe project [project-name]` | Project details | `--org` |
| `describe integration [integration-name]` (`integrations`) | Build, deployment and endpoint details | `--org`, `-p`/`--project` |
| `describe build [build-id]` | Build status and details | `--org`, `-p`/`--project`, `-i`/`--integration`, `-d`/`--deployment-track` |
| `describe deployment` | Deployment status and deployed endpoints/URLs | `--org`, `-p`/`--project`, `-i`/`--integration`, `-e`/`--env`, `-d`/`--deployment-track` |
| `describe config` | A specific config-map or secret | `-n`/`--name`, `--org`, `-p`/`--project`, `-i`/`--integration`, `-e`/`--env`, `-d`/`--deployment-track` |
| `describe connection` | A specific connection | `-n`/`--name`, `--org`, `-p`/`--project`, `-i`/`--integration` |
| `describe execution` (`executions`) | A specific execution | `--id`, `--org`, `-p`/`--project`, `-i`/`--integration`, `-e`/`--env` |

```bash
wso2-integration-platform describe integration <integration-name> --project=<project-name> --output=json
```

---

## delete

| Command | Deletes | Flags |
|---------|---------|-------|
| `delete project [project-name]` | A project | `-f`/`--force`, `--org` |
| `delete integration [integration-name]` (`integrations`) | An integration | `--skip-confirm`, `--org`, `-p`/`--project` |
| `delete config [config-name]` | A config or secret | `--skip-confirm`, `--org`, `-p`/`--project`, `-i`/`--integration`, `-e`/`--env`, `-d`/`--deployment-track` |

```bash
wso2-integration-platform delete integration <integration-name> --project=<project-name>
```

---

## suspend / resume

| Command | Effect |
|---------|--------|
| `suspend deployment [integration-name]` | Suspend a running deployment |
| `resume deployment [integration-name]` | Resume a suspended deployment |

Both take `--org`, `-p`/`--project`, `-e`/`--env` and `-d`/`--deployment-track`.

```bash
wso2-integration-platform suspend deployment <integration-name> -e=Development
```

---

## logs

Alias: `log`.

| Command | Shows | Extra flags |
|---------|-------|-------------|
| `logs application` | Application logs of a deployed integration | `-f`/`--follow`, `-l`/`--limit` (default `100`), `-q`/`--query`, `-e`/`--env`, `-d`/`--deployment-track` |
| `logs gateway` | Gateway logs of a deployed integration | `-f`/`--follow`, `-l`/`--limit` (default `100`), `-q`/`--query`, `-e`/`--env`, `-d`/`--deployment-track` |
| `logs build` | Logs for a specific build run | `--build-id`, `--step`, `-d`/`--deployment-track` |
| `logs executions` (`execution`) | Logs of a scheduled/manual task execution | `--id`, `--attempt`, `-e`/`--env`, `-d`/`--deployment-track` |

All four also take `--org`, `-p`/`--project` and `-i`/`--integration`.

```bash
wso2-integration-platform logs application --integration=<integration> --project=<project> --follow
```

---

## connect

Connect your development environment to a remote project environment, injecting
the connection configurations your integration depends on.

| Flag | Description |
|------|-------------|
| `--skip-connection` | Skip injecting configurations for the named connections (repeatable). |
| `--recreate` | Recreate the bridge component. |
| `--delete-bridge` | Delete the bridge between the local and remote environments. |
| `--org`, `-p`/`--project`, `-i`/`--integration`, `-e`/`--env`, `-d`/`--deployment-track` | Common resource flags. |

```bash
# A shell connected to your project environment
wso2-integration-platform connect --project <project-name>

# Run a local integration against its remote dependencies
wso2-integration-platform connect --project <project-name> -- <command-to-start-local-component>
```

---

## Hidden commands

These are not listed in `--help`; they exist for tooling integrations.

| Command | Purpose | Flags |
|---------|---------|-------|
| `start-mcp-server` | Start the MCP server exposing CLI operations as tools for AI assistants | `--http` (HTTP mode), `--port` (default `8080`) |
| `start-rpc-server` | JSON-RPC bridge used by editor extensions | `--method`, `--params` (both hidden) |

See the [MCP Server](../README.md#mcp-server) section of the README for client
configuration.

---

## Environment variables

Reported by `wso2-integration-platform env`, with their current values:

| Variable | Purpose |
|----------|---------|
| `OUTPUT_FORMAT` | Default for `--output` (`table`, `json`). Default `table`. |
| `WSO2IP_PAT` | Personal Access Token for non-interactive authentication. |
| `WSO2IP_REGION` | Force the region (`US` or `EU`). |

Also read by the CLI, but not listed by `env`:

| Variable | Purpose |
|----------|---------|
| `CI` | When `true`, `--non-interactive` defaults on. |
| `TRACE_ENABLED` | When `true`, log all HTTP requests and responses. |
| `CLOUD_STS_TOKEN` | Bootstrap a session from a VS Code/cloud extension STS token. |
| `CLOUD_INITIAL_ORG_ID` | Pre-select an organization on first launch (cloud extension context). |
| `CLOUD_INITIAL_PROJECT_ID` | Pre-select a project on first launch (cloud extension context). |
| `WSO2IP_ENV` | Target `dev` or `stage` instead of production. Requires `WSO2IP_ENV_CONFIG`; for CLI development only. |
