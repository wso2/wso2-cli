package integration

import (
	"strings"
	"testing"

	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
)

func TestResolveSubtype(t *testing.T) {
	cases := []struct {
		name                 string
		subtype              string
		wantComponentType    string
		wantComponentSubType string
	}{
		{"automation", SubtypeAutomation, component.ComponentTypeScheduledTask, ""},
		{"api", SubtypeAPI, component.ComponentTypeService, ""},
		{"ai agent", SubtypeAIAgent, component.ComponentTypeService, "aiAgent"},
		{"mcp server", SubtypeMCPServer, component.ComponentTypeService, "MCP"},
		{"event integration", SubtypeEventIntegration, component.ComponentTypeEventHandler, ""},
		{"file integration", SubtypeFileIntegration, component.ComponentTypeEventHandler, "fileIntegration"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotComponentType, gotComponentSubType, err := ResolveSubtype(tc.subtype)
			if err != nil {
				t.Fatalf("ResolveSubtype(%q) returned unexpected error: %v", tc.subtype, err)
			}
			if gotComponentType != tc.wantComponentType {
				t.Errorf("ResolveSubtype(%q) componentType = %q, want %q", tc.subtype, gotComponentType, tc.wantComponentType)
			}
			if gotComponentSubType != tc.wantComponentSubType {
				t.Errorf("ResolveSubtype(%q) componentSubType = %q, want %q", tc.subtype, gotComponentSubType, tc.wantComponentSubType)
			}
		})
	}
}

// Excluded Choreo component types (webApp, manualTask, webhook, proxy) and any
// unrecognized or mis-cased value must be rejected: the MCP server can only
// ever reach the six integration subtypes by omission, not by runtime guard,
// so ResolveSubtype is the only gate standing between a caller and the
// backend's full component type vocabulary.
func TestResolveSubtypeRejectsUnsupportedValues(t *testing.T) {
	unsupported := []string{
		component.ComponentTypeByocWebApp,
		component.ComponentTypeManualTrigger,
		component.ComponentTypeWebhook,
		component.ComponentTypeProxyGH,
		"automation", // wrong case
		"",
	}
	for _, subtype := range unsupported {
		t.Run(subtype, func(t *testing.T) {
			if _, _, err := ResolveSubtype(subtype); err == nil {
				t.Errorf("ResolveSubtype(%q) expected an error, got nil", subtype)
			}
		})
	}
}

func TestSubtypesListMatchesMappings(t *testing.T) {
	if len(Subtypes) != len(subtypeMappings) {
		t.Fatalf("Subtypes has %d entries, subtypeMappings has %d", len(Subtypes), len(subtypeMappings))
	}
	for _, s := range Subtypes {
		if _, ok := subtypeMappings[s]; !ok {
			t.Errorf("Subtypes contains %q which has no entry in subtypeMappings", s)
		}
	}
}

// The cloud editor has no integrated terminal, so an agent cannot ask the user
// to run `git branch --show-current`. create_integration therefore treats
// `branch` as optional and probes main then master itself.
func TestBranchDefaults(t *testing.T) {
	if DefaultBranch != "main" {
		t.Errorf("DefaultBranch = %q, want main", DefaultBranch)
	}
	if FallbackBranch != "master" {
		t.Errorf("FallbackBranch = %q, want master", FallbackBranch)
	}
	if DefaultBranch == FallbackBranch {
		t.Error("default and fallback branch must differ, or the retry is pointless")
	}
}

// The branch argument must not be marked required, and must tell the agent not
// to go looking for the branch via a terminal command.
func TestBranchArgumentIsOptionalAndDocumented(t *testing.T) {
	schema := CreateIntegrationTool.InputSchema

	for _, req := range schema.Required {
		if req == "branch" {
			t.Error("branch must NOT be required — the cloud editor cannot determine it")
		}
	}

	prop, ok := schema.Properties["branch"].(map[string]any)
	if !ok {
		t.Fatal("branch property missing from schema")
	}
	desc, _ := prop["description"].(string)

	for _, want := range []string{"OPTIONAL", DefaultBranch, FallbackBranch, "no terminal"} {
		if !strings.Contains(desc, want) {
			t.Errorf("branch description must mention %q", want)
		}
	}
	if !strings.Contains(desc, "git branch --show-current") {
		t.Error("branch description should name the command the agent must not ask for")
	}
}

// The MCP server cannot reload the editor window: the protocol has no
// host-command primitive, and the server runs as a stdio child of the editor, so
// a reload would kill the process before it could return the tool result.
// Reloading is therefore the user's decision — and it ends the agent
// conversation, so the notice must state that cost rather than presenting a
// reload as a routine step an agent can suggest freely.
func TestReloadCostNoticeStatesTheCost(t *testing.T) {
	notice := reloadCostNotice

	// The consequence must be unmissable.
	for _, want := range []string{"WILL END", "chat history", "lost"} {
		if !strings.Contains(notice, want) {
			t.Errorf("reloadCostNotice must mention %q", want)
		}
	}

	// The agent must not act on it unilaterally, nor imply it is required.
	for _, want := range []string{
		"Do not reload on the user's behalf",
		"do not present it as a",
		"let them decide",
	} {
		if !strings.Contains(notice, want) {
			t.Errorf("reloadCostNotice must contain the restraint %q", want)
		}
	}

	// It must not describe reloading as necessary for the integration to work.
	for _, banned := range []string{"you must reload", "reload is required", "must be reloaded"} {
		if strings.Contains(strings.ToLower(notice), banned) {
			t.Errorf("reloadCostNotice must not frame reloading as required (found %q)", banned)
		}
	}
}

func comp(name, handler, org, repo, branch, subpath, displayType string) models.Component {
	c := models.Component{
		Name: name, Handler: handler, Id: "uuid-" + handler, DisplayType: displayType,
	}
	c.Repository.OrganizationApp = org
	c.Repository.NameApp = repo
	c.Repository.BranchApp = branch
	c.Repository.AppSubPath = subpath
	return c
}

// Asking to "deploy" twice must not silently produce a second integration
// building from the same source in the same project.
func TestFindExistingIntegration(t *testing.T) {
	existing := []models.Component{
		comp("orders-api", "orders-api", "acme", "orders", "main", "", "ballerinaService"),
		// A monorepo: same repo and branch, different subpath. Legitimately a
		// separate integration.
		comp("billing", "billing", "acme", "mono", "main", "services/billing", "ballerinaService"),
	}

	cases := []struct {
		name                                string
		newName, org, repo, branch, subpath string
		wantMatch                           bool
		wantReason                          string
	}{
		{"same name", "orders-api", "other", "different", "main", "", true, "name"},
		{"same name, different case", "ORDERS-API", "other", "different", "main", "", true, "name"},
		{"same source, different name", "orders-api-v2", "acme", "orders", "main", "", true, "source"},
		{"same source, case/slash variance", "x", "ACME", "Orders", "main", "/", true, "source"},

		{"different repo", "new-thing", "acme", "payments", "main", "", false, ""},
		{"same repo, different branch", "new-thing", "acme", "orders", "develop", "", false, ""},

		// The case that must not regress: another subpath in the same monorepo.
		{"monorepo, different subpath", "shipping", "acme", "mono", "main", "services/shipping", false, ""},
		{"monorepo, same subpath is a dup", "shipping", "acme", "mono", "main", "services/billing", true, "source"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := findExistingIntegration(existing, tc.newName, tc.org, tc.repo, tc.branch, tc.subpath)
			if tc.wantMatch {
				if got == nil {
					t.Fatal("expected a duplicate to be detected")
				}
				if got.reason != tc.wantReason {
					t.Errorf("reason = %q, want %q", got.reason, tc.wantReason)
				}
			} else if got != nil {
				t.Errorf("unexpected duplicate: matched %q by %s", got.component.Name, got.reason)
			}
		})
	}
}

// Non-integration components (e.g. legacy web apps) share the project but are
// not integrations, so they must never block an integration create.
func TestFindExistingIntegrationIgnoresNonIntegrations(t *testing.T) {
	existing := []models.Component{
		comp("legacy-ui", "legacy-ui", "acme", "orders", "main", "", "byocWebApp"),
	}
	if got := findExistingIntegration(existing, "legacy-ui", "acme", "orders", "main", ""); got != nil {
		t.Errorf("a non-integration component must not block: matched by %s", got.reason)
	}
}

// The guidance must point at the supported way to deploy elsewhere, and must not
// suggest working around the block by renaming.
func TestDuplicateGuidanceDirectsToAnotherProject(t *testing.T) {
	c := comp("orders-api", "orders-api", "acme", "orders", "main", "", "ballerinaService")
	d := duplicateMatch{component: &c, reason: "source"}
	msg := d.detail("Payments", "https://console.example/overview")

	for _, want := range []string{
		"project_uuid", // how to target another project
		"get_projects", // how to find it
		"create_build", // what to do if they just want new code shipped
		"Do NOT create a duplicate",
		"uuid-orders-api", // the existing integration is identified
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("guidance must mention %q", want)
		}
	}
	if !strings.Contains(msg, "Payments") {
		t.Error("guidance should name the project the conflict is in")
	}
}
