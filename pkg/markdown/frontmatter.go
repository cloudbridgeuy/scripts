package markdown

import "strings"

// StripFrontmatter removes a leading YAML frontmatter block delimited by "---"
// lines. Input without a complete frontmatter block is returned unchanged.
func StripFrontmatter(src []byte) []byte {
	s := string(src)
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		return src
	}
	lines := strings.SplitAfter(s, "\n")
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r\n") == "---" {
			return []byte(strings.Join(lines[i+1:], ""))
		}
	}
	return src
}

// ExtractTitle returns the text of the first ATX H1 heading ("# ...").
// Without an H1, the fallback is returned.
func ExtractTitle(src []byte, fallback string) string {
	for _, line := range strings.Split(string(src), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			if text := strings.TrimSpace(trimmed[2:]); text != "" {
				return text
			}
		}
	}
	return fallback
}
