package markdown

import "testing"

func TestResolveOutputPath(t *testing.T) {
	tests := []struct {
		name       string
		inputPath  string
		outputFlag string
		want       string
	}{
		{"md input, no flag", "doc.md", "", "doc.html"},
		{"md input with dir, no flag", "/x/y/doc.md", "", "/x/y/doc.html"},
		{"extensionless input, no flag", "README", "", "README.html"},
		{"dotted directory, no flag", "/a.b/doc.md", "", "/a.b/doc.html"},
		{"explicit override wins", "doc.md", "/out/page.html", "/out/page.html"},
		{"explicit override, odd extension", "doc.markdown", "out.htm", "out.htm"},
		{"multiple dots, only last ext swapped", "archive.tar.gz", "", "archive.tar.html"},
		{"empty input", "", "", ".html"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveOutputPath(tt.inputPath, tt.outputFlag)
			if got != tt.want {
				t.Errorf("ResolveOutputPath(%q, %q) = %q, want %q",
					tt.inputPath, tt.outputFlag, got, tt.want)
			}
		})
	}
}

func TestResolveOutputTarget(t *testing.T) {
	tests := []struct {
		name       string
		inputPath  string
		outputFlag string
		open       bool
		want       OutputTarget
	}{
		{"no flags: sibling path", "notes/doc.md", "", false, OutputTarget{Path: "notes/doc.html"}},
		{"output flag wins, not temp", "doc.md", "/out/page.html", false, OutputTarget{Path: "/out/page.html"}},
		{"output flag wins even with open", "doc.md", "/out/page.html", true, OutputTarget{Path: "/out/page.html"}},
		{"open alone: temp pattern", "notes/doc.md", "", true, OutputTarget{Path: "doc-*.html", Temp: true}},
		{"open, extensionless input", "README", "", true, OutputTarget{Path: "README-*.html", Temp: true}},
		{"open, multiple dots: last ext dropped", "archive.tar.gz", "", true, OutputTarget{Path: "archive.tar-*.html", Temp: true}},
		{"open, dotfile input falls back", ".hidden", "", true, OutputTarget{Path: "markdown-*.html", Temp: true}},
		{"open, empty input falls back", "", "", true, OutputTarget{Path: "markdown-*.html", Temp: true}},
		{"no flags, empty input: sibling rule", "", "", false, OutputTarget{Path: ".html"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveOutputTarget(tt.inputPath, tt.outputFlag, tt.open)
			if got != tt.want {
				t.Errorf("ResolveOutputTarget(%q, %q, %v) = %#v, want %#v",
					tt.inputPath, tt.outputFlag, tt.open, got, tt.want)
			}
		})
	}
}
