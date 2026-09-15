package testkey

import (
	"strings"
	"testing"
)

// Test keys are only meaningful for integrations that serve HTTP, and the
// handler rejects anything else at runtime. An agent reading 'Automation' as a
// generic "runs unattended" category will try to generate a key for one, so the
// description must rule that out before the call.
func TestGenerateTestKeyDocumentsEndpointOnlyScope(t *testing.T) {
	desc := GenerateTestKeyTool.Description

	if !strings.Contains(desc, "APPLIES ONLY TO") {
		t.Error("description must state that the tool is restricted")
	}
	for _, applicable := range []string{"'API'", "'AI Agent'", "'MCP Server'"} {
		if !strings.Contains(desc, applicable) {
			t.Errorf("description must name %s as applicable", applicable)
		}
	}
	for _, excluded := range []string{"Automation", "Event Integration", "File Integration"} {
		if !strings.Contains(desc, excluded) {
			t.Errorf("description must name %q as not applicable", excluded)
		}
	}
	// Point the agent at the right tool instead of leaving a dead end.
	if !strings.Contains(desc, "execute_task") {
		t.Error("description should redirect Automation callers to execute_task")
	}
	// The other runtime gate, which is easy to hit and easy to document.
	if !strings.Contains(strings.ToLower(desc), "critical") {
		t.Error("description must mention that critical/production environments are not permitted")
	}
}
