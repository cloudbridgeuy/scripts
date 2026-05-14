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
