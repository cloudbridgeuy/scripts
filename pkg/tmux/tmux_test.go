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
