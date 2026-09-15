package testkey

import "github.com/mark3labs/mcp-go/mcp"

var GenerateTestKeyTool = mcp.NewTool("generate_test_key",
	mcp.WithTitleAnnotation("Generate Test Key"),
	mcp.WithDescription("Generates a test key for invoking a secured integration's HTTP endpoint in a non-production environment. Pass the key in the Test-Key header. If the integration has multiple endpoints, specify endpoint_uuid; otherwise the first public endpoint is used.\n\nAPPLIES ONLY TO INTEGRATIONS THAT EXPOSE AN HTTP ENDPOINT: the 'API', 'AI Agent' and 'MCP Server' subtypes. Not valid for 'Automation' (a scheduled task — run it with execute_task instead) or 'Event Integration' / 'File Integration' (event handlers with no endpoint to call). Calling this on any of those returns an error. Also not permitted in critical/production environments — direct the user to the web console for those."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(false),
	mcp.WithOpenWorldHintAnnotation(true),
	mcp.WithString("project_uuid",
		mcp.Required(),
		mcp.Description("The UUID of the project containing the component."),
	),
	mcp.WithString("integration_uuid",
		mcp.Required(),
		mcp.Description("The UUID of the component to generate the test key for."),
	),
	mcp.WithString("environment_uuid",
		mcp.Required(),
		mcp.Description("The UUID of the environment (e.g., Development, Staging) where the test key is needed."),
	),
	mcp.WithString("endpoint_uuid",
		mcp.Description("The UUID of a specific endpoint to generate the test key for. If not provided, the first public endpoint will be used."),
	),
)
