package report

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestFormatXML(t *testing.T) {
	tests := []struct {
		name    string
		results []Result
		verify  func(t *testing.T, output string)
	}{
		{
			name: "single result with all fields",
			results: []Result{
				{
					Action:   Action{Description: "Say hello", Command: "echo hello"},
					ExitCode: 0,
					Output:   "hello",
				},
			},
			verify: func(t *testing.T, output string) {
				// Must be valid XML
				var r xmlReport
				if err := xml.Unmarshal([]byte(output), &r); err != nil {
					t.Fatalf("invalid XML: %v", err)
				}
				if len(r.Actions) != 1 {
					t.Fatalf("expected 1 action, got %d", len(r.Actions))
				}
				a := r.Actions[0]
				if a.Description != "Say hello" {
					t.Errorf("description = %q, want %q", a.Description, "Say hello")
				}
				if a.Command != "echo hello" {
					t.Errorf("command = %q, want %q", a.Command, "echo hello")
				}
				if a.Status != 0 {
					t.Errorf("status = %d, want 0", a.Status)
				}
				if strings.TrimSpace(a.Output) != "hello" {
					t.Errorf("output = %q, want %q", strings.TrimSpace(a.Output), "hello")
				}
			},
		},
		{
			name: "multiple results",
			results: []Result{
				{
					Action:   Action{Description: "First", Command: "echo first"},
					ExitCode: 0,
					Output:   "first",
				},
				{
					Action:   Action{Description: "Second", Command: "echo second"},
					ExitCode: 1,
					Output:   "second",
				},
			},
			verify: func(t *testing.T, output string) {
				var r xmlReport
				if err := xml.Unmarshal([]byte(output), &r); err != nil {
					t.Fatalf("invalid XML: %v", err)
				}
				if len(r.Actions) != 2 {
					t.Fatalf("expected 2 actions, got %d", len(r.Actions))
				}
				if strings.TrimSpace(r.Actions[0].Output) != "first" {
					t.Errorf("first output = %q", strings.TrimSpace(r.Actions[0].Output))
				}
				if strings.TrimSpace(r.Actions[1].Output) != "second" {
					t.Errorf("second output = %q", strings.TrimSpace(r.Actions[1].Output))
				}
				if r.Actions[1].Status != 1 {
					t.Errorf("second status = %d, want 1", r.Actions[1].Status)
				}
			},
		},
		{
			name: "XML escaping of special characters",
			results: []Result{
				{
					Action:   Action{Description: "Check if a < b & c > d", Command: "test 1 < 2 && echo \"yes\""},
					ExitCode: 0,
					Output:   "<output>&result</output>",
				},
			},
			verify: func(t *testing.T, output string) {
				// Must be valid XML despite special chars
				var r xmlReport
				if err := xml.Unmarshal([]byte(output), &r); err != nil {
					t.Fatalf("invalid XML: %v", err)
				}
				a := r.Actions[0]
				if a.Description != "Check if a < b & c > d" {
					t.Errorf("description not properly round-tripped: %q", a.Description)
				}
				if a.Command != "test 1 < 2 && echo \"yes\"" {
					t.Errorf("command not properly round-tripped: %q", a.Command)
				}
				// Check raw output contains escaped forms
				if !strings.Contains(output, "&lt;") || !strings.Contains(output, "&amp;") {
					t.Error("expected XML escaping in raw output")
				}
			},
		},
		{
			name: "empty description",
			results: []Result{
				{
					Action:   Action{Description: "", Command: "ls"},
					ExitCode: 0,
					Output:   "file.txt",
				},
			},
			verify: func(t *testing.T, output string) {
				if !strings.Contains(output, "<description></description>") {
					t.Error("expected <description></description> for empty description")
				}
				var r xmlReport
				if err := xml.Unmarshal([]byte(output), &r); err != nil {
					t.Fatalf("invalid XML: %v", err)
				}
			},
		},
		{
			name: "empty results",
			results: []Result{},
			verify: func(t *testing.T, output string) {
				var r xmlReport
				if err := xml.Unmarshal([]byte(output), &r); err != nil {
					t.Fatalf("invalid XML: %v", err)
				}
				if len(r.Actions) != 0 {
					t.Errorf("expected 0 actions, got %d", len(r.Actions))
				}
				if !strings.Contains(output, "<report>") {
					t.Error("expected <report> wrapper")
				}
			},
		},
		{
			name: "output with trailing newline is trimmed",
			results: []Result{
				{
					Action:   Action{Description: "test", Command: "echo hi"},
					ExitCode: 0,
					Output:   "hi\n\n",
				},
			},
			verify: func(t *testing.T, output string) {
				var r xmlReport
				if err := xml.Unmarshal([]byte(output), &r); err != nil {
					t.Fatalf("invalid XML: %v", err)
				}
				trimmed := strings.TrimSpace(r.Actions[0].Output)
				if trimmed != "hi" {
					t.Errorf("output = %q, want %q", trimmed, "hi")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := FormatXML(tt.results)
			if err != nil {
				t.Fatalf("FormatXML returned unexpected error: %v", err)
			}
			tt.verify(t, output)
		})
	}
}

func TestFormatMarkdown(t *testing.T) {
	tests := []struct {
		name    string
		results []Result
		verify  func(t *testing.T, output string)
	}{
		{
			name: "single result with all fields",
			results: []Result{
				{
					Action:   Action{Description: "Say hello", Command: "echo hello"},
					ExitCode: 0,
					Output:   "hello",
				},
			},
			verify: func(t *testing.T, output string) {
				if !strings.Contains(output, "# Report") {
					t.Error("missing # Report heading")
				}
				if !strings.Contains(output, "## Command 1") {
					t.Error("missing ## Command 1 heading")
				}
				if !strings.Contains(output, "Say hello") {
					t.Error("missing description")
				}
				if !strings.Contains(output, "**Status Code**: 0") {
					t.Error("missing status code")
				}
				if !strings.Contains(output, "```\necho hello\n```") {
					t.Error("missing command code block")
				}
				if !strings.Contains(output, "**Output**:") {
					t.Error("missing output label")
				}
				if !strings.Contains(output, "```\nhello\n```") {
					t.Error("missing output code block")
				}
			},
		},
		{
			name: "multiple results",
			results: []Result{
				{
					Action:   Action{Description: "First", Command: "echo 1"},
					ExitCode: 0,
					Output:   "1",
				},
				{
					Action:   Action{Description: "Second", Command: "echo 2"},
					ExitCode: 1,
					Output:   "2",
				},
			},
			verify: func(t *testing.T, output string) {
				if !strings.Contains(output, "## Command 1") {
					t.Error("missing ## Command 1")
				}
				if !strings.Contains(output, "## Command 2") {
					t.Error("missing ## Command 2")
				}
				if !strings.Contains(output, "**Status Code**: 1") {
					t.Error("missing status code 1")
				}
			},
		},
		{
			name: "empty description omitted",
			results: []Result{
				{
					Action:   Action{Description: "", Command: "ls"},
					ExitCode: 0,
					Output:   "file.txt",
				},
			},
			verify: func(t *testing.T, output string) {
				// Between "## Command 1\n" and "**Status Code**" there should be no description line
				lines := strings.Split(output, "\n")
				foundHeading := false
				for i, line := range lines {
					if strings.Contains(line, "## Command 1") {
						foundHeading = true
						// Next non-empty line should be **Status Code**
						for j := i + 1; j < len(lines); j++ {
							if strings.TrimSpace(lines[j]) != "" {
								if !strings.Contains(lines[j], "**Status Code**") {
									t.Errorf("expected Status Code after heading, got %q", lines[j])
								}
								break
							}
						}
						break
					}
				}
				if !foundHeading {
					t.Error("missing ## Command 1 heading")
				}
			},
		},
		{
			name: "empty results",
			results: []Result{},
			verify: func(t *testing.T, output string) {
				if !strings.Contains(output, "# Report") {
					t.Error("missing # Report heading")
				}
				// Should just be the heading
				if strings.Contains(output, "## Command") {
					t.Error("should have no command sections")
				}
			},
		},
		{
			name: "output with trailing newline is trimmed",
			results: []Result{
				{
					Action:   Action{Description: "test", Command: "echo hi"},
					ExitCode: 0,
					Output:   "hi\n\n",
				},
			},
			verify: func(t *testing.T, output string) {
				// The output block should contain "hi" without trailing newlines
				if !strings.Contains(output, "```\nhi\n```") {
					t.Error("output should be trimmed of trailing newlines")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatMarkdown(tt.results)
			tt.verify(t, output)
		})
	}
}

func TestFormatReport(t *testing.T) {
	results := []Result{
		{
			Action:   Action{Description: "test", Command: "echo hi"},
			ExitCode: 0,
			Output:   "hi",
		},
	}

	t.Run("dispatches to XML", func(t *testing.T) {
		got, err := FormatReport(results, XML)
		if err != nil {
			t.Fatalf("FormatReport(XML) returned unexpected error: %v", err)
		}
		want, err := FormatXML(results)
		if err != nil {
			t.Fatalf("FormatXML returned unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("FormatReport(XML) differs from FormatXML:\ngot:  %q\nwant: %q", got, want)
		}
	})

	t.Run("dispatches to Markdown", func(t *testing.T) {
		got, err := FormatReport(results, Markdown)
		if err != nil {
			t.Fatalf("FormatReport(Markdown) returned unexpected error: %v", err)
		}
		want := FormatMarkdown(results)
		if got != want {
			t.Errorf("FormatReport(Markdown) differs from FormatMarkdown:\ngot:  %q\nwant: %q", got, want)
		}
	})

	t.Run("returns error for unknown format", func(t *testing.T) {
		_, err := FormatReport(results, Format("unknown"))
		if err == nil {
			t.Fatal("expected error for unknown format, got nil")
		}
		if !strings.Contains(err.Error(), "unsupported format") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "unsupported format")
		}
	})
}
