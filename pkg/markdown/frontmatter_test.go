package markdown

import "testing"

func TestStripFrontmatter(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "no frontmatter returned unchanged",
			src:  "# Title\n\nbody\n",
			want: "# Title\n\nbody\n",
		},
		{
			name: "valid frontmatter block removed",
			src:  "---\ntitle: Hello\ndate: 2026-05-14\n---\n# Title\n\nbody\n",
			want: "# Title\n\nbody\n",
		},
		{
			name: "leading --- that is a horizontal rule, not frontmatter",
			src:  "---\n\njust a rule above\n",
			want: "---\n\njust a rule above\n",
		},
		{
			name: "empty input",
			src:  "",
			want: "",
		},
		{
			name: "CRLF line endings",
			src:  "---\r\ntitle: Hello\r\n---\r\n# Title\r\n",
			want: "# Title\r\n",
		},
		{
			name: "frontmatter with empty body",
			src:  "---\ntitle: x\n---\n",
			want: "",
		},
		{
			name: "unclosed frontmatter block",
			src:  "---\ntitle: x\nno closing delimiter\n",
			want: "---\ntitle: x\nno closing delimiter\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(StripFrontmatter([]byte(tt.src)))
			if got != tt.want {
				t.Errorf("StripFrontmatter() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		fallback string
		want     string
	}{
		{"h1 present", "# My Doc\n\nbody", "fallback", "My Doc"},
		{"no h1 uses fallback", "no heading here\n", "fallback", "fallback"},
		{"h1 not on first line", "intro line\n\n#  Spaced Title \n", "fallback", "Spaced Title"},
		{"h2 is not an h1", "## Subhead\n", "fallback", "fallback"},
		{"hash with no space is not a heading", "#Title\n", "fallback", "fallback"},
		{"empty input uses fallback", "", "fallback", "fallback"},
		{"bare # heading falls back", "# \n\nbody\n", "fallback", "fallback"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractTitle([]byte(tt.src), tt.fallback)
			if got != tt.want {
				t.Errorf("ExtractTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}
