package tmux

import (
	"reflect"
	"testing"
)

func TestParseNonEmptyLines(t *testing.T) {
	t.Parallel()

	result := parseNonEmptyLines("\n first \n\nsecond\n  \nthird\n")
	expected := []string{"first", "second", "third"}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("unexpected parsed lines: %#v", result)
	}
}

func TestCanonicalSessionName(t *testing.T) {
	t.Parallel()

	if got := canonicalSessionName("/tmp/opencode.test"); got != "/tmp/opencode_test" {
		t.Fatalf("unexpected canonical name: %q", got)
	}
}

func TestHarnessSymbol(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"working":            "●",
		"waiting-input":      "?",
		"waiting-permission": "!",
		"done":               "✓",
		"":                   "·",
		"unknown":            "·",
	}

	for state, expected := range cases {
		if got := HarnessSymbol(state); got != expected {
			t.Fatalf("state %q: got %q, want %q", state, got, expected)
		}
	}
}

func TestParseHarnessFieldsWithState(t *testing.T) {
	t.Parallel()

	p, ok := ParseHarnessFields([]string{
		"$1", "/Users/test/proj", "@1", "1", "claude", "%1", "1",
		"claude", "waiting-input", "claude ?", "zsh", "claude", "/Users/test/proj",
	})
	if !ok {
		t.Fatal("expected fields to parse")
	}
	if p.Harness != "claude" || p.State != "waiting-input" || p.PaneID != "%1" {
		t.Fatalf("unexpected pane: %#v", p)
	}
}

func TestParseHarnessFieldsInfersFromCommand(t *testing.T) {
	t.Parallel()

	p, ok := ParseHarnessFields([]string{
		"$2", "/Users/test/proj", "@2", "2", "zsh", "%2", "1",
		"", "", "OC | Some task", "opencode", "opencode", "/Users/test/proj",
	})
	if !ok {
		t.Fatal("expected fields to parse")
	}
	if p.Harness != "opencode" {
		t.Fatalf("unexpected harness: %q", p.Harness)
	}
}

func TestParseHarnessFieldsSkipsNonHarness(t *testing.T) {
	t.Parallel()

	if _, ok := ParseHarnessFields([]string{
		"$3", "/Users/test/proj", "@3", "1", "zsh", "%3", "2",
		"", "", "zsh in proj", "zsh", "zsh", "/Users/test/proj",
	}); ok {
		t.Fatal("expected non-harness fields to be skipped")
	}
}

func TestParseHarnessFieldsRejectsBadShape(t *testing.T) {
	t.Parallel()

	if _, ok := ParseHarnessFields([]string{"a", "b"}); ok {
		t.Fatal("expected malformed fields to be rejected")
	}
}

func TestParseHarnessFieldsDetectsClaudeByVersionCommand(t *testing.T) {
	t.Parallel()

	p, ok := ParseHarnessFields([]string{
		"$4", "/Users/test/mcptools", "@4", "3", "zsh", "%4", "1",
		"", "", "", "2.1.278", "", "/Users/test/mcptools",
	})
	if !ok {
		t.Fatal("expected version-command pane to parse")
	}
	if p.Harness != "claude" {
		t.Fatalf("unexpected harness: %q", p.Harness)
	}
	if IsStaleHarness(p) {
		t.Fatal("expected version-command pane to be kept")
	}
}

func TestIsStaleHarness(t *testing.T) {
	t.Parallel()

	stale := HarnessPane{Harness: "claude", State: "done", Title: "zsh in mcptools", Command: "zsh"}
	if !IsStaleHarness(stale) {
		t.Fatal("expected exited harness with shell foreground to be stale")
	}

	kept := []HarnessPane{
		{Harness: "claude", State: "waiting-input", Title: "claude ?", Command: "2.1.278"},
		{Harness: "opencode", Title: "OC | Task", Command: "opencode"},
		{Harness: "claude", State: "working", Title: "claude ●", Command: "zsh"},
		{Harness: "codex", Title: "Implement x | proj", Command: "codex"},
	}
	for _, p := range kept {
		if IsStaleHarness(p) {
			t.Fatalf("expected pane to be kept: %#v", p)
		}
	}
}
