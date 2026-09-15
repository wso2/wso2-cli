package buildpack

import "github.com/mark3labs/mcp-go/mcp"

var GetBuildPacksTool = mcp.NewTool("get_buildpacks",
	mcp.WithTitleAnnotation("Get Build Packs"),
	mcp.WithDescription("Fetches the buildpacks the platform supports for a given integration component type. Only WSO2 Integrator profiles are returned: Ballerina (the default profile) and WSO2 Micro Integrator/MI. Each entry's 'supportedVersions' field lists the runtime versions the platform can build.\n\nNote that create_integration does NOT take a buildpack id or a language version — it takes only `buildpack` ('ballerina' or 'microintegrator'). The Ballerina distribution version is maintained by the source repository in its Ballerina.toml, and the platform builds whatever that file specifies. Use this tool to confirm the platform supports the version the repository targets, not to supply one to create_integration."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(true),
	mcp.WithOpenWorldHintAnnotation(true),
	mcp.WithString("type",
		mcp.Required(),
		mcp.Description("The underlying component type to list buildpacks for. Valid values: 'scheduleTask' (Automation), 'service' (API / AI Agent / MCP Server), 'eventHandler' (Event Integration / File Integration)."),
		mcp.Enum(
			"scheduleTask",
			"service",
			"eventHandler",
		),
	),
)
