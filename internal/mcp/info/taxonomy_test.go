package info

import (
	"encoding/json"
	"strings"
	"testing"
)

// getStartedResponse is a hand-maintained JSON literal. If it stops parsing,
// every agent loses its entire operational manual, so assert it directly.
func TestGetStartedResponseIsValidJSON(t *testing.T) {
	var v map[string]any
	if err := json.Unmarshal([]byte(getStartedResponse), &v); err != nil {
		t.Fatalf("getStartedResponse is not valid JSON: %v", err)
	}
}

// subtypes returns the documented integration subtype map.
func subtypes(t *testing.T) map[string]any {
	t.Helper()
	var kb map[string]any
	if err := json.Unmarshal([]byte(getStartedResponse), &kb); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	guide, ok := kb["integration_guide"].(map[string]any)
	if !ok {
		t.Fatal("integration_guide missing")
	}
	st, ok := guide["integration_subtypes"].(map[string]any)
	if !ok {
		t.Fatal("integration_subtypes missing")
	}
	return st
}

// The knowledge base must document exactly the six subtypes the server can
// create — no more, no fewer. Drift here is what let an agent conclude that
// Event Integration was a kind of Automation.
//
// Kept as a literal list rather than importing internal/mcp/integration, since
// that package imports pkg/api and this assertion is about the documentation
// contract, not the mapping implementation (which subtypes_test.go covers).
func TestDocumentedSubtypesMatchSupportedSet(t *testing.T) {
	want := map[string]string{
		"Automation":        "scheduleTask",
		"API":               "service",
		"AI Agent":          "service",
		"MCP Server":        "service",
		"Event Integration": "eventHandler",
		"File Integration":  "eventHandler",
	}

	types, ok := subtypes(t)["types"].(map[string]any)
	if !ok {
		t.Fatal("integration_subtypes.types missing")
	}

	if len(types) != len(want) {
		t.Errorf("documented %d subtypes, want %d: %v", len(types), len(want), keys(types))
	}

	for name, wantComponentType := range want {
		entry, ok := types[name].(map[string]any)
		if !ok {
			t.Errorf("subtype %q is not documented", name)
			continue
		}
		got, _ := entry["backend_component_type"].(string)
		if got != wantComponentType {
			t.Errorf("subtype %q: backend_component_type = %q, want %q", name, got, wantComponentType)
		}
		// Each entry must say what it does and does not support, or the
		// capability confusion returns.
		for _, field := range []string{"trigger", "supports", "does_not_support"} {
			if s, _ := entry[field].(string); strings.TrimSpace(s) == "" {
				t.Errorf("subtype %q: %q is empty", name, field)
			}
		}
	}

	for name := range types {
		if _, ok := want[name]; !ok {
			t.Errorf("subtype %q is documented but not supported by the server", name)
		}
	}
}

// Executions are the sharpest capability split: only the scheduleTask-backed
// Automation supports them (see isExecutionSupported in internal/mcp/execution).
// An agent that treats Event Integration as an Automation will call
// get_executions on it and fail, so the docs must state this explicitly.
func TestOnlyAutomationDocumentsExecutionSupport(t *testing.T) {
	types := subtypes(t)["types"].(map[string]any)

	for name, raw := range types {
		entry := raw.(map[string]any)
		supports, _ := entry["supports"].(string)
		notSupports, _ := entry["does_not_support"].(string)

		mentionsInSupports := strings.Contains(supports, "get_executions") || strings.Contains(supports, "execute_task")
		mentionsInDenial := strings.Contains(notSupports, "executions")

		if name == "Automation" {
			if !mentionsInSupports {
				t.Error("Automation must document that it supports executions")
			}
			continue
		}
		if mentionsInSupports {
			t.Errorf("subtype %q claims execution support, but only Automation has it", name)
		}
		if !mentionsInDenial {
			t.Errorf("subtype %q must state that executions do not apply", name)
		}
	}
}

// The rules must explicitly deny the hierarchy an agent previously invented.
func TestTaxonomyRulesDenyHierarchy(t *testing.T) {
	st := subtypes(t)
	rulesRaw, ok := st["taxonomy_rules"].([]any)
	if !ok || len(rulesRaw) == 0 {
		t.Fatal("taxonomy_rules missing or empty")
	}
	var all string
	for _, r := range rulesRaw {
		s, _ := r.(string)
		all += " " + s
	}
	// Compare case-insensitively: the assertion is about the meaning being
	// present, not the exact emphasis used.
	allLower := strings.ToLower(all)

	for _, phrase := range []string{"mutually exclusive", "no hierarchy", "not a kind of automation"} {
		if !strings.Contains(allLower, phrase) {
			t.Errorf("taxonomy_rules must convey %q", phrase)
		}
	}

	if _, ok := st["capability_matrix"].(map[string]any); !ok {
		t.Error("capability_matrix missing — agents need the per-subtype tool mapping")
	}
}

func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// The Ballerina distribution version lives in the repository's Ballerina.toml,
// not in any platform setting. Models tend to hand-write that file with a
// version from training data, which silently pins the project to an old
// runtime — so the guidance must warn against that and must not itself pin a
// version that will age.
func TestProjectScaffoldingGuidance(t *testing.T) {
	var kb map[string]any
	if err := json.Unmarshal([]byte(getStartedResponse), &kb); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	guide := kb["integration_guide"].(map[string]any)

	ps, ok := guide["project_scaffolding"].(map[string]any)
	if !ok {
		t.Fatal("integration_guide.project_scaffolding missing")
	}

	src, _ := ps["version_source_of_truth"].(string)
	if !strings.Contains(src, "Ballerina.toml") {
		t.Error("must state that Ballerina.toml holds the version")
	}

	rulesRaw, ok := ps["rules"].([]any)
	if !ok || len(rulesRaw) == 0 {
		t.Fatal("project_scaffolding.rules missing or empty")
	}
	var all string
	for _, r := range rulesRaw {
		s, _ := r.(string)
		all += " " + s
	}
	lower := strings.ToLower(all)

	for _, want := range []string{"memory", "bal new", "bal version"} {
		if !strings.Contains(lower, strings.ToLower(want)) {
			t.Errorf("rules must mention %q", want)
		}
	}

	// A concrete Swan Lake version here would go stale and reintroduce the very
	// problem the guidance exists to prevent.
	for _, stale := range []string{"2201.", "swan-lake-alpha", "swan-lake-beta"} {
		if strings.Contains(lower, strings.ToLower(stale)) {
			t.Errorf("guidance must not pin a concrete version (found %q)", stale)
		}
	}
}

// The product is WSO2 Integrator; Ballerina is its default runtime profile.
// Agent-facing text should lead with the product rather than describing the work
// as "writing Ballerina". Technical identifiers (Ballerina.toml, bal new, the
// 'ballerina' buildpack value) are exempt — those are real names.
func TestProductLeadsRuntimeInAgentFacingText(t *testing.T) {
	var kb map[string]any
	if err := json.Unmarshal([]byte(getStartedResponse), &kb); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Phrasings that frame the runtime as the thing being built.
	banned := []string{
		"implemented in Ballerina",
		"non-Ballerina code",
		"writing Ballerina code for",
		"build a Ballerina integration",
		"deploy Ballerina to",
	}

	blob, err := json.Marshal(kb)
	if err != nil {
		t.Fatal(err)
	}
	text := string(blob)

	for _, phrase := range banned {
		if strings.Contains(text, phrase) {
			t.Errorf("agent-facing text should lead with WSO2 Integrator, found %q", phrase)
		}
	}

	// The disambiguation must still be present — dropping it entirely would
	// reintroduce the opposite bug, where an agent writes MI code by default.
	if !strings.Contains(text, "default") || !strings.Contains(text, "Ballerina") {
		t.Error("text must still identify Ballerina as the default profile")
	}
	if !strings.Contains(text, "Micro Integrator") {
		t.Error("text must still name the MI profile")
	}
}

// The real friction for an external agent is getting the files it just wrote into
// git at all — the platform only builds from a repository. Handing the user a
// manual upload works, but the smoother fix is for the agent to gain push ability,
// so gaining it must be offered FIRST, at the moment the content is ready, rather
// than mentioned afterwards as a "next time" footnote.
func TestGitHandoffOffersToolingBeforeManualUpload(t *testing.T) {
	local, cloud := gitWorkflowVariants()

	mcpIdx := strings.Index(local, "GitHub MCP server to their client")
	if mcpIdx == -1 {
		t.Fatal("local workflow must offer connecting a GitHub MCP server")
	}
	browserIdx := strings.Index(local, "https://github.com/new")
	if browserIdx == -1 {
		t.Fatal("local workflow must retain the browser fallback")
	}

	// Order is the point: tooling offered before the manual route.
	if mcpIdx > browserIdx {
		t.Error("the GitHub MCP offer must come before the manual browser route, not after it")
	}
	if !strings.Contains(local, "OPTION A") || !strings.Contains(local, "OPTION B") {
		t.Error("the two routes should be labelled so the agent presents them in order")
	}

	// Restarting a client to pick up a new MCP server can end the conversation —
	// the same cost documented for the editor reload. The user must be told.
	for _, want := range []string{"reconnected or restarted", "end the current conversation"} {
		if !strings.Contains(local, want) {
			t.Errorf("must warn that adding an MCP server may need a restart: expected %q", want)
		}
	}

	// Still soft, and still never a prerequisite.
	for _, want := range []string{"Offer Option A once", "accept a 'no' immediately", "never block the deployment"} {
		if !strings.Contains(local, want) {
			t.Errorf("the offer must stay soft: expected %q", want)
		}
	}

	// Files first, so the user is never left waiting.
	writeIdx := strings.Index(local, "BEFORE asking the")
	if writeIdx == -1 {
		t.Error("must instruct writing the project before asking the user for anything")
	}

	// Tokens are never required by either route.
	if !strings.Contains(local, "NEVER ask the user for a GitHub token") {
		t.Error("must forbid asking for a GitHub token")
	}

	// The cloud editor has its own Source Control UI flow and no terminal;
	// suggesting extra tooling there is noise.
	if strings.Contains(cloud, "GitHub MCP") {
		t.Error("cloud workflow must not suggest a GitHub MCP server")
	}
}
