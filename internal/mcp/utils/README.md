# MCP Utils - Client Access Pattern

This document describes the generalized client access pattern for MCP tool implementations.

## New Pattern for Client Access inside MCP tool Implementations

### 1. Authentication Check
```go
if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
    return authErr, nil
}
```

### 2. Client Access
```go
projectClient := utils.GetProjectClient(ctx)
componentClient := utils.GetComponentClient(ctx)
// ... other clients
```

### 3. Context-Aware Utils
```go
selectedProject, err := utils.GetTargetProjectCtx(ctx, *targetOrg, request)
```

## Complete Example

```go
func getProjects(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // 1. Check authentication (stdio mode only)
    if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
        return authErr, nil
    }

    // 2. Get organization
    targetOrg, err := utils.GetTargetOrgByUuid(request)
    if err != nil {
        return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
    }

    // 3. Get client (context-aware)
    projectClient := utils.GetProjectClient(ctx)
    
    // 4. Use client
    projects, err := projectClient.GetProjectsByOrgID(targetOrg.ID)
    if err != nil {
        return mcp.NewToolResultErrorFromErr("failed to get projects", err), nil
    }

    return utils.CreateJSONResponse(projects)
}
```

## Available Client Functions

All client functions automatically handle context-aware selection:

- `utils.GetProjectClient(ctx)`
- `utils.GetComponentClient(ctx)`
- `utils.GetOrgClient(ctx)`
- `utils.GetGraphQLClient(ctx)`
- `utils.GetGitClient(ctx)`
- `utils.GetLogsClient(ctx)`
- `utils.GetApimPublisherClient(ctx)`
- `utils.GetDevopsClient(ctx)`
- `utils.GetDeploymentBuildClient(ctx)`
- `utils.GetConnectionsClient(ctx)`
- `utils.GetMarketplaceClient(ctx)`
- `utils.GetGQLBuildClient(ctx)`
- `utils.GetCredentialClient(ctx)`
- `utils.GetWorkflowMgtClient(ctx)`
- `utils.GetSubscriptionsClient(ctx)`
- `utils.GetConfigMappingApiClient(ctx)`
- `utils.GetPlatformServicesClient(ctx)`
- `utils.GetUserStoreMgtClient(ctx)`

## Context-Aware Utils Functions

For functions that need to access clients:

- `utils.GetTargetProjectCtx(ctx, targetOrg, request)` - replaces `utils.GetTargetProject()`
- `utils.GetTargetComponentCtx(ctx, targetOrg, targetProject, request)` - replaces `utils.GetTargetComponent()`
- `utils.GetTargetEnvCtx(ctx, targetOrg, targetProject, request)` - replaces `utils.GetTargetEnv()`

Legacy functions still work for backward compatibility but use background context.

## How It Works

### HTTP Mode
- Context contains authentication token extracted from HTTP request headers
- Each request creates temporary `ContextTokenStore` with that token
- Client functions detect token in context and create request-scoped clients
- No global state mutation - thread-safe for concurrent requests

### Stdio Mode
- Context has no authentication token
- Falls back to global clients (`auth.ProjectClient`, etc.)
- Global authentication check required via `auth.IsLoggedIn()`
- Single-user experience

## Migration Guide

### Before
```go
func myTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    if !auth.IsLoggedIn() {
        return mcp.NewToolResultError("Error: You are not logged in..."), nil
    }
    
    // Get things
    targetOrg, err := utils.GetTargetOrgByUuid(request)
    targetProject, err := utils.GetTargetProject(*targetOrg, request)
    
    // Use global client
    result, err := auth.ProjectClient.SomeMethod(...)
    return utils.CreateJSONResponse(result)
}
```

### After
```go
func myTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
        return authErr, nil
    }
    
    // Get things (context-aware)
    targetOrg, err := utils.GetTargetOrgByUuid(request)
    targetProject, err := utils.GetTargetProjectCtx(ctx, *targetOrg, request)
    
    // Use context-aware client
    projectClient := utils.GetProjectClient(ctx)
    result, err := projectClient.SomeMethod(...)
    return utils.CreateJSONResponse(result)
}
```

## Benefits

✅ **DRY**: Eliminates repetitive client selection logic  
✅ **Consistent**: All tools use the same pattern  
✅ **Thread-Safe**: No race conditions in HTTP mode  
✅ **Maintainable**: Changes in one place  
✅ **Backward Compatible**: Legacy functions still work  
✅ **Type Safe**: Proper client types returned