package tmux

import (
	"os/exec"
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

	cases := map[string]string{
		"/tmp/opencode.test": "/tmp/opencode_test",
		"web1:/a/b":          "web1_/a/b",
		"web1:/tmp/a.b":      "web1_/tmp/a_b",
		"/a/b:c":             "/a/b_c",
	}

	for input, expected := range cases {
		if got := CanonicalSessionName(input); got != expected {
			t.Fatalf("CanonicalSessionName(%q) = %q, want %q", input, got, expected)
		}
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

	stale := []HarnessPane{
		{Harness: "claude", State: "done", Title: "zsh in mcptools", Command: "zsh"},
		{Harness: "claude", State: "done", Title: "OC | Task", Command: "zsh"},
		{Harness: "opencode", Title: "claude ?", Command: "zsh"},
	}
	for _, p := range stale {
		if !IsStaleHarness(p) {
			t.Fatalf("expected pane to be stale: %#v", p)
		}
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

func TestParseHarnessFieldsLiveCommandOverridesStoredOpts(t *testing.T) {
	t.Parallel()

	p, ok := ParseHarnessFields([]string{
		"$5", "/Users/test/mcptools", "@5", "3", "GUZ-161", "%5", "1",
		"claude", "waiting-input", "OC | Auto-DAG sliced feature orchestration", "opencode", "", "/Users/test/mcptools",
	})
	if !ok {
		t.Fatal("expected fields to parse")
	}
	if p.Harness != "opencode" {
		t.Fatalf("unexpected harness: %q", p.Harness)
	}
	if p.State != "" {
		t.Fatalf("expected reset state, got %q", p.State)
	}
	if IsStaleHarness(p) {
		t.Fatal("expected overridden pane to be kept")
	}
}

func TestParseTarget(t *testing.T) {
	t.Parallel()

	cases := []struct {
		arg     string
		host    string
		session string
	}{
		{"/a/b", "", "/a/b"},
		{"web1:/a/b", "web1", "/a/b"},
		{"web1:foo", "web1", "foo"},
		{"/a/b:c", "", "/a/b:c"},
		{"web1:", "web1", ""},
		{":x", "", ":x"},
		{"my:session", "my", "session"},
	}

	for _, c := range cases {
		host, session := ParseTarget(c.arg)
		if host != c.host || session != c.session {
			t.Fatalf("ParseTarget(%q) = (%q, %q), want (%q, %q)", c.arg, host, session, c.host, c.session)
		}
	}
}

func TestHistoryName(t *testing.T) {
	t.Parallel()

	if got := HistoryName("", "/a/b"); got != "/a/b" {
		t.Fatalf("unexpected local history name: %q", got)
	}
	if got := HistoryName("web1", "/a/b"); got != "web1:/a/b" {
		t.Fatalf("unexpected remote history name: %q", got)
	}
}

func TestShellQuoteAndJoin(t *testing.T) {
	t.Parallel()

	if got := shellQuote("a b"); got != "'a b'" {
		t.Fatalf("unexpected quote: %q", got)
	}
	if got := shellQuote("it's"); got != `'it'"'"'s'` {
		t.Fatalf("unexpected quote with apostrophe: %q", got)
	}
	if got := shellJoin([]string{"tmux", "ls", "-F", "a b"}); got != `'tmux' 'ls' '-F' 'a b'` {
		t.Fatalf("unexpected join: %q", got)
	}
}

func TestRemotePaneCommand(t *testing.T) {
	t.Parallel()

	got := remotePaneCommand("h", "/x", "/tmp/sock")
	want := `sh -c 'sock=/tmp/scripts-tmux-$$.sock; exec ssh -t -R "$sock:"'"'"'/tmp/sock'"'"' h "export SCRIPTS_TMUX_SOCKET=$sock; " '"'"'cd '"'"'"'"'"'"'"'"'/x'"'"'"'"'"'"'"'"' && "$SHELL" -l; rm -f "$SCRIPTS_TMUX_SOCKET"'"'"''`
	if got != want {
		t.Fatalf("unexpected pane command: %q", got)
	}

	if err := exec.Command("sh", "-n", "-c", got).Run(); err != nil {
		t.Fatalf("pane command failed sh -n: %v", err)
	}
}

func TestParseRemoteEnv(t *testing.T) {
	t.Parallel()

	host, dir, ok := parseRemoteEnv("SCRIPTS_REMOTE_HOST=web1\nSCRIPTS_REMOTE_DIR=/a/b\n")
	if !ok || host != "web1" || dir != "/a/b" {
		t.Fatalf("unexpected remote env: (%q, %q, %v)", host, dir, ok)
	}

	if _, _, ok := parseRemoteEnv("SCRIPTS_REMOTE_HOST=web1\n"); ok {
		t.Fatal("expected missing dir to be rejected")
	}

	if _, _, ok := parseRemoteEnv("SCRIPTS_REMOTE_DIR=/a/b\n"); ok {
		t.Fatal("expected missing host to be rejected")
	}

	if _, _, ok := parseRemoteEnv(""); ok {
		t.Fatal("expected empty output to be rejected")
	}

	if _, _, ok := parseRemoteEnv("SCRIPTS_REMOTE_HOST=\nSCRIPTS_REMOTE_DIR=/a/b\n"); ok {
		t.Fatal("expected empty host to be rejected")
	}
}

func TestTmuxSocketArgs(t *testing.T) {
	t.Setenv("SCRIPTS_TMUX_SOCKET", "")
	if got := tmuxSocketArgs("ls", "-F"); !reflect.DeepEqual(got, []string{"ls", "-F"}) {
		t.Fatalf("unexpected args with unset socket: %#v", got)
	}

	t.Setenv("SCRIPTS_TMUX_SOCKET", "/tmp/my sock")
	if got := tmuxSocketArgs("ls", "-F"); !reflect.DeepEqual(got, []string{"-S", "/tmp/my sock", "ls", "-F"}) {
		t.Fatalf("unexpected args with set socket: %#v", got)
	}
}
