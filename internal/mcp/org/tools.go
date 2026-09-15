package org

import "github.com/mark3labs/mcp-go/mcp"

var GetActiveOrgTool = mcp.NewTool("get_active_org",
	mcp.WithTitleAnnotation("Get Active Organization"),
	mcp.WithDescription("Retrieves the details of the currently active organization."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(true),
	mcp.WithOpenWorldHintAnnotation(false),
)

var ChangeOrgTool = mcp.NewTool("change_org",
	mcp.WithTitleAnnotation("Change Organization"),
	mcp.WithDescription("Changes the active organization for all subsequent operations within the current session. This allows you to switch your working context to a different organization you have access to. All future tool calls will apply to this newly set organization."),
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithDestructiveHintAnnotation(true),
	mcp.WithIdempotentHintAnnotation(true),
	mcp.WithOpenWorldHintAnnotation(true),
	mcp.WithString("org_uuid",
		mcp.Required(),
		mcp.Description("The UUID of the organization to switch to."),
	),
)

var GetOrganizationsTool = mcp.NewTool("get_organizations",
	mcp.WithTitleAnnotation("Get Organizations"),
	mcp.WithDescription("Retrieves a list of all organizations that you have access to."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(true),
	mcp.WithOpenWorldHintAnnotation(true),
)
