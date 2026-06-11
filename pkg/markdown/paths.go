package markdown

import (
	"path/filepath"
	"strings"
)

// ResolveOutputPath determines where the HTML output is written.
// An empty outputFlag swaps the input file's extension for ".html";
// an extensionless input gains ".html". A set outputFlag is returned verbatim.
func ResolveOutputPath(inputPath, outputFlag string) string {
	if outputFlag != "" {
		return outputFlag
	}
	ext := filepath.Ext(inputPath)
	if ext == "" {
		return inputPath + ".html"
	}
	return strings.TrimSuffix(inputPath, ext) + ".html"
}

// OutputTarget says where the rendered HTML goes.
// Temp=false: Path is the concrete output path.
// Temp=true:  Path is an os.CreateTemp pattern like "doc-*.html"; the
// imperative shell turns it into a real file in the OS temp directory.
type OutputTarget struct {
	Path string
	Temp bool
}

// ResolveOutputTarget decides the output destination. An explicit outputFlag
// always wins and is never temporary. Without it, open renders to a
// temporary file so the source directory stays clean; otherwise the sibling
// rule from ResolveOutputPath applies.
//
// When constructing the temp pattern the directory portion of inputPath is
// stripped — only the basename without its extension feeds the pattern.
// Nameless inputs (empty string, dotfiles) fall back to "markdown".
func ResolveOutputTarget(inputPath, outputFlag string, open bool) OutputTarget {
	if outputFlag != "" {
		return OutputTarget{Path: outputFlag}
	}
	if open {
		base := filepath.Base(inputPath)
		base = strings.TrimSuffix(base, filepath.Ext(base))
		if base == "" || base == "." {
			base = "markdown"
		}
		return OutputTarget{Path: base + "-*.html", Temp: true}
	}
	return OutputTarget{Path: ResolveOutputPath(inputPath, "")}
}
