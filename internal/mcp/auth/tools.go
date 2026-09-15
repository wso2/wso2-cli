package auth

import "github.com/mark3labs/mcp-go/mcp"

var LoginTool = mcp.NewTool("login",
	mcp.WithDescription("Authenticate with your Integration Platform account. Returns a browser URL to open to complete authentication. After opening the URL and completing sign-in, call check_login_status to confirm."),
	// Writes local credential state, but only adds a session — it destroys
	// nothing. Not idempotent: each call starts a fresh browser auth flow.
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(false),
	mcp.WithOpenWorldHintAnnotation(true),
)

var CheckLoginStatusTool = mcp.NewTool("check_login_status",
	mcp.WithDescription("Check if you are currently authenticated with the Integration Platform."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(true),
	mcp.WithOpenWorldHintAnnotation(true),
)
