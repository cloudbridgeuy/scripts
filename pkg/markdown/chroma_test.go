package markdown

import (
	"strings"
	"testing"
)

func TestChromaCSS(t *testing.T) {
	css, err := ChromaCSS()
	if err != nil {
		t.Fatalf("ChromaCSS() error = %v", err)
	}
	if strings.TrimSpace(css) == "" {
		t.Fatal("ChromaCSS() returned empty string")
	}
	if !strings.Contains(css, ".chroma") {
		t.Errorf("ChromaCSS() output missing .chroma selectors:\n%s", css)
	}
}
