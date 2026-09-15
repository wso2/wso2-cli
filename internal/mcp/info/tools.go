package info

import "github.com/mark3labs/mcp-go/mcp"

var GetStartedTool = mcp.NewTool("get_started",
	mcp.WithTitleAnnotation("Get Started"),
	mcp.WithDescription("MUST be called at the start of any task involving integration, automation, scheduling, event handling, API development, or cloud deployment. Loads the operational directives, supported runtimes, and platform constraints for the WSO2 Integration Platform. Always call this before writing any code or making implementation decisions."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(true),
	mcp.WithOpenWorldHintAnnotation(false),
)
