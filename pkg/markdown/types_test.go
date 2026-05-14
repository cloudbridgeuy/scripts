package markdown

import "testing"

func TestNewRenderConfig(t *testing.T) {
	t.Run("resolves sibling path when no output flag", func(t *testing.T) {
		cfg := NewRenderConfig("notes/doc.md", "", false)
		if cfg.InputPath != "notes/doc.md" {
			t.Errorf("InputPath = %q, want %q", cfg.InputPath, "notes/doc.md")
		}
		if cfg.OutputPath != "notes/doc.html" {
			t.Errorf("OutputPath = %q, want %q", cfg.OutputPath, "notes/doc.html")
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
		if cfg.OutputPath != "/tmp/page.html" {
			t.Errorf("OutputPath = %q, want %q", cfg.OutputPath, "/tmp/page.html")
		}
		if !cfg.Open {
			t.Errorf("Open = false, want true")
		}
	})
}
