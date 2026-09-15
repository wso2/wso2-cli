package execution

import (
	"strings"
	"testing"

	"github.com/wso2/integration-platform-tools/pkg/api/component"
)

// Executions exist only for the scheduleTask family, which the MCP server
// exposes as the single 'Automation' subtype. An agent that reads 'Automation'
// as an umbrella for anything unattended will call these tools on an Event
// Integration and fail, so both descriptions must state the restriction up
// front rather than leaving it to a runtime error.
func TestExecutionToolsDocumentAutomationOnly(t *testing.T) {
	tools := map[string]string{
		"get_executions": GetExecutionsTool.Description,
		"execute_task":   ExecuteTaskTool.Description,
	}

	for name, desc := range tools {
		t.Run(name, func(t *testing.T) {
			if !strings.Contains(desc, "'Automation'") {
				t.Error("description must name the 'Automation' subtype")
			}
			if !strings.Contains(desc, "APPLIES ONLY TO") && !strings.Contains(desc, "ONLY") {
				t.Error("description must state that the tool is restricted to one subtype")
			}
			// The subtypes that will otherwise be mistakenly passed here.
			for _, excluded := range []string{"API", "Event Integration", "File Integration"} {
				if !strings.Contains(desc, excluded) {
					t.Errorf("description must name %q as not applicable", excluded)
				}
			}
			if !strings.Contains(strings.ToLower(desc), "not an umbrella") {
				t.Error("description must deny the 'Automation means anything unattended' reading")
			}
		})
	}
}

// isExecutionSupported is the gate the descriptions describe; keep them honest.
func TestIsExecutionSupportedMatchesDocumentedScope(t *testing.T) {
	// Automation resolves to scheduleTask, whose display type must be accepted.
	if !isExecutionSupported(component.DisplayTypeScheduledTask) {
		t.Error("scheduled task must support executions — Automation depends on it")
	}

	// The service- and eventHandler-backed subtypes must not.
	for _, dt := range []string{
		component.DisplayTypeService,
		component.DisplayTypeBallerinaEventHandler,
		component.DisplayTypeMiEventHandler,
	} {
		if isExecutionSupported(dt) {
			t.Errorf("display type %q must NOT support executions", dt)
		}
	}
}
