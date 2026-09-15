package component

import "testing"

// The integration set is the Devant console's DevantComponentDisplayTypeSet
// (workspaces/apps/choreo-console/src/types/projects.ts), ORed with the MCP
// subtype, exactly as src/hooks/project.tsx stamps isDevantComponent.
//
// Several of these cases reverse what this package asserted previously, when
// membership was derived from GetTypeForDisplayType. That derivation both
// admitted types the console excludes (prism mocks, byoc/buildpack services)
// and dropped types it includes (manual triggers, MI jobs, webhooks), so the
// reversals are the point of the test, not incidental.
func TestIsIntegrationComponent(t *testing.T) {
	integrations := []string{
		DisplayTypeRestApi,
		DisplayTypeManualTrigger, // was excluded: maps to "manual-task"
		DisplayTypeScheduledTask,
		DisplayTypeWebhook, // was excluded: maps to "web-hook"
		DisplayTypeMiRestApi,
		DisplayTypeMiEventHandler,
		DisplayTypeService, // "ballerinaService"
		DisplayTypeMiApiService,
		DisplayTypeMiCronjob,
		DisplayTypeMiJob,     // was excluded
		DisplayTypeMiWebhook, // had no constant at all
		DisplayTypeBallerinaEventHandler,
		DisplayTypeBallerinaWebhook,         // had no constant at all
		DisplayTypeBallerinaFileIntegration, // had no constant at all
		DisplayTypeMiFileIntegration,        // had no constant at all
		DisplayTypeByoiService,
	}
	for _, dt := range integrations {
		if !IsIntegrationComponent(dt, "") {
			t.Errorf("IsIntegrationComponent(%q, \"\") = false, want true", dt)
		}
	}

	nonIntegrations := []string{
		DisplayTypeByocWebApp,
		DisplayTypeByoiWebApp,
		DisplayTypeBuildpackWebApp,
		DisplayTypeByocWebAppDockerLess,
		DisplayTypePrismMockService, // was INCLUDED: mapped to "service"
		DisplayTypeByocService,      // was INCLUDED
		DisplayTypeBuildpackService, // was INCLUDED
		DisplayTypeByocRestApi,      // was INCLUDED
		DisplayTypeBuildpackRestApi, // was INCLUDED
		DisplayTypeByocCronjob,      // was INCLUDED
		DisplayTypeBuildpackCronJob, // was INCLUDED
		DisplayTypeByoiCronjob,      // was INCLUDED
		DisplayTypeGraphQL,          // was INCLUDED
		DisplayTypeWebsocket,        // was INCLUDED
		DisplayTypeProxy,
		DisplayTypeGitProxy,
		DisplayTypeBuildpackTestRunner,
		DisplayTypeBuildpackWebhook,
		"unknown-display-type",
		"",
	}
	for _, dt := range nonIntegrations {
		if IsIntegrationComponent(dt, "") {
			t.Errorf("IsIntegrationComponent(%q, \"\") = true, want false", dt)
		}
	}
}

// An MCP Server is admitted by subtype alone — its displayType is not in the
// set, so a displayType-only check drops it entirely.
func TestMcpServerAdmittedBySubtypeAlone(t *testing.T) {
	if IsIntegrationComponent("mcpService", "") {
		t.Error("mcpService displayType alone should not be in the set")
	}
	if !IsIntegrationComponent("mcpService", ComponentSubTypeMCP) {
		t.Error("an MCP subtype must admit the component regardless of displayType")
	}
}

// File integrations arrive with the value in either field, so both must match.
func TestFileIntegrationMatchesInEitherField(t *testing.T) {
	for _, v := range []string{DisplayTypeBallerinaFileIntegration, DisplayTypeMiFileIntegration} {
		if !IsIntegrationComponent(v, "") {
			t.Errorf("%q as displayType must be an integration", v)
		}
		if !IsIntegrationComponent(DisplayTypeService, v) {
			t.Errorf("%q as componentSubType must be an integration", v)
		}
	}
}

// The subtype refines the reported kind: a service carrying an aiAgent subtype
// is an AI Agent, not a plain API. This is the distinction the CLI had no
// concept of, since it never read componentSubType.
func TestIntegrationKindUsesSubtype(t *testing.T) {
	cases := []struct {
		displayType, subType, want string
	}{
		{DisplayTypeService, ComponentSubTypeAiAgent, "AI Agent"},
		{DisplayTypeService, ComponentSubTypeMCP, "MCP Server"},
		{DisplayTypeService, DisplayTypeBallerinaFileIntegration, "File Integration"},
		{DisplayTypeMiFileIntegration, "", "File Integration"},
		{DisplayTypeScheduledTask, "", "Automation"},
		{DisplayTypeMiJob, "", "Automation"},
		{DisplayTypeBallerinaEventHandler, "", "Event Integration"},
		{DisplayTypeMiWebhook, "", "Event Integration"},
		{DisplayTypeService, "", "API"},
		{DisplayTypeRestApi, "", "API"},
	}
	for _, c := range cases {
		if got := IntegrationKind(c.displayType, c.subType); got != c.want {
			t.Errorf("IntegrationKind(%q, %q) = %q, want %q", c.displayType, c.subType, got, c.want)
		}
	}
}
