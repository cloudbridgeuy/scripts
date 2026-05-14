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
