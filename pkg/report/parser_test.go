package report

import (
	"testing"
)

func TestParseActions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Action
	}{
		{
			name:  "single command with description",
			input: "# list files\nls -la",
			expected: []Action{
				{Description: "list files", Command: "ls -la"},
			},
		},
		{
			name:  "multiple commands",
			input: "# first\nls\n# second\npwd",
			expected: []Action{
				{Description: "first", Command: "ls"},
				{Description: "second", Command: "pwd"},
			},
		},
		{
			name:  "line continuation",
			input: "gcloud foo \\\n  --bar baz \\\n  --qux",
			expected: []Action{
				{Description: "", Command: "gcloud foo --bar baz --qux"},
			},
		},
		{
			name:  "no description",
			input: "ls -la",
			expected: []Action{
				{Description: "", Command: "ls -la"},
			},
		},
		{
			name:  "blank lines between commands",
			input: "# first\nls\n\n# second\npwd",
			expected: []Action{
				{Description: "first", Command: "ls"},
				{Description: "second", Command: "pwd"},
			},
		},
		{
			name:     "comment without command",
			input:    "# orphan comment",
			expected: []Action{},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []Action{},
		},
		{
			name:  "mixed: some with descriptions, some without",
			input: "# described\nls\npwd",
			expected: []Action{
				{Description: "described", Command: "ls"},
				{Description: "", Command: "pwd"},
			},
		},
		{
			name:  "multiple comments before command",
			input: "# first\n# second\nls",
			expected: []Action{
				{Description: "second", Command: "ls"},
			},
		},
		{
			name:  "continuation with description",
			input: "# deploy\ngcloud \\\n  --project foo",
			expected: []Action{
				{Description: "deploy", Command: "gcloud --project foo"},
			},
		},
		{
			name:  "indented comment lines",
			input: "  # Check role\n  gcloud iam roles describe foo\n  # Check binding\n  gcloud projects get-iam-policy bar",
			expected: []Action{
				{Description: "Check role", Command: "  gcloud iam roles describe foo"},
				{Description: "Check binding", Command: "  gcloud projects get-iam-policy bar"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseActions(tt.input)

			if len(got) != len(tt.expected) {
				t.Fatalf("ParseActions() returned %d actions, want %d\ngot:  %+v\nwant: %+v", len(got), len(tt.expected), got, tt.expected)
			}

			for i := range tt.expected {
				if got[i].Description != tt.expected[i].Description {
					t.Errorf("action[%d].Description = %q, want %q", i, got[i].Description, tt.expected[i].Description)
				}
				if got[i].Command != tt.expected[i].Command {
					t.Errorf("action[%d].Command = %q, want %q", i, got[i].Command, tt.expected[i].Command)
				}
			}
		})
	}
}
