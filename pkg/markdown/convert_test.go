package markdown

import (
	"strings"
	"testing"
)

func TestRenderMarkdown(t *testing.T) {
	t.Run("plain prose renders to html", func(t *testing.T) {
		out, err := RenderMarkdown([]byte("# Heading\n\nA paragraph.\n"))
		if err != nil {
			t.Fatalf("RenderMarkdown() error = %v", err)
		}
		if !strings.Contains(out, "<h1") {
			t.Errorf("expected <h1> in output:\n%s", out)
		}
		if !strings.Contains(out, "<p>A paragraph.</p>") {
			t.Errorf("expected paragraph in output:\n%s", out)
		}
	})

	t.Run("highlighted code fence uses chroma classes", func(t *testing.T) {
		src := "```go\nfunc main() {}\n```\n"
		out, err := RenderMarkdown([]byte(src))
		if err != nil {
			t.Fatalf("RenderMarkdown() error = %v", err)
		}
		if !strings.Contains(out, "class=\"chroma\"") {
			t.Errorf("expected chroma class wrapper in output:\n%s", out)
		}
		if strings.Contains(out, "class=\"mermaid\"") {
			t.Errorf("go fence must not be treated as mermaid:\n%s", out)
		}
	})

	t.Run("mermaid fence passes through without highlighting", func(t *testing.T) {
		src := "```mermaid\ngraph TD; A-->B;\n```\n"
		out, err := RenderMarkdown([]byte(src))
		if err != nil {
			t.Fatalf("RenderMarkdown() error = %v", err)
		}
		if !strings.Contains(out, `<pre class="mermaid">`) {
			t.Errorf("expected <pre class=\"mermaid\"> in output:\n%s", out)
		}
		if !strings.Contains(out, "graph TD; A--&gt;B;") {
			t.Errorf("expected escaped mermaid source in output:\n%s", out)
		}
		if strings.Contains(out, "class=\"chroma\"") {
			t.Errorf("mermaid fence must not be highlighted by chroma:\n%s", out)
		}
	})

	t.Run("unlabeled code fence still renders", func(t *testing.T) {
		src := "```\nsome plain text\n```\n"
		out, err := RenderMarkdown([]byte(src))
		if err != nil {
			t.Fatalf("RenderMarkdown() error = %v", err)
		}
		if !strings.Contains(out, `class="chroma"`) {
			t.Errorf("expected chroma class wrapper in output:\n%s", out)
		}
	})

	t.Run("unknown language tag still renders", func(t *testing.T) {
		src := "```wat-no-such-lang\nsome code\n```\n"
		out, err := RenderMarkdown([]byte(src))
		if err != nil {
			t.Fatalf("RenderMarkdown() error = %v", err)
		}
		if !strings.Contains(out, `class="chroma"`) {
			t.Errorf("expected chroma class wrapper in output:\n%s", out)
		}
	})

	t.Run("gfm table renders", func(t *testing.T) {
		src := "| a | b |\n|---|---|\n| 1 | 2 |\n"
		out, err := RenderMarkdown([]byte(src))
		if err != nil {
			t.Fatalf("RenderMarkdown() error = %v", err)
		}
		if !strings.Contains(out, "<table>") {
			t.Errorf("expected <table> in output:\n%s", out)
		}
	})
}
