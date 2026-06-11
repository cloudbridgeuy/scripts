package markdown

import "testing"

func TestNewRenderConfig(t *testing.T) {
	t.Run("resolves sibling path when no output flag", func(t *testing.T) {
		cfg := NewRenderConfig("notes/doc.md", "", false)
		if cfg.InputPath != "notes/doc.md" {
			t.Errorf("InputPath = %q, want %q", cfg.InputPath, "notes/doc.md")
		}
		if cfg.Output != (OutputTarget{Path: "notes/doc.html"}) {
			t.Errorf("Output = %#v, want sibling notes/doc.html", cfg.Output)
		}
		if cfg.Open {
			t.Errorf("Open = true, want false")
		}
	})

	t.Run("honors output flag and open flag", func(t *testing.T) {
		cfg := NewRenderConfig("doc.md", "/tmp/page.html", true)
		if cfg.InputPath != "doc.md" {
			t.Errorf("InputPath = %q, want %q", cfg.InputPath, "doc.md")
		}
		if cfg.Output != (OutputTarget{Path: "/tmp/page.html"}) {
			t.Errorf("Output = %#v, want explicit /tmp/page.html", cfg.Output)
		}
		if !cfg.Open {
			t.Errorf("Open = false, want true")
		}
	})

	t.Run("open without output flag resolves a temp pattern", func(t *testing.T) {
		cfg := NewRenderConfig("notes/doc.md", "", true)
		if cfg.Output != (OutputTarget{Path: "doc-*.html", Temp: true}) {
			t.Errorf("Output = %#v, want temp pattern doc-*.html", cfg.Output)
		}
		if !cfg.Open {
			t.Errorf("Open = false, want true")
		}
	})
}
