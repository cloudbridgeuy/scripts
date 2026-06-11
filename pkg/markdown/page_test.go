package markdown

import (
	"strings"
	"testing"
)

func TestBuildPage(t *testing.T) {
	page := BuildPage("<p>hello</p>", "My <Title>", ".chroma { color: #fff; }", "")

	if !strings.Contains(page, "<p>hello</p>") {
		t.Errorf("body not injected:\n%s", page)
	}
	if !strings.Contains(page, "<title>My &lt;Title&gt;</title>") {
		t.Errorf("title not injected and escaped:\n%s", page)
	}
	if !strings.Contains(page, ".chroma { color: #fff; }") {
		t.Errorf("chroma CSS not injected:\n%s", page)
	}
	if !strings.Contains(page, "--bg: #1a1b26;") {
		t.Errorf("page CSS not injected:\n%s", page)
	}
	if !strings.Contains(page, "mermaid.min.js") {
		t.Errorf("mermaid script tag missing:\n%s", page)
	}
	if strings.Contains(page, "{{") {
		t.Errorf("unreplaced placeholder remains:\n%s", page)
	}
}

// TestBuildPageDoesNotReExpandPlaceholders documents that strings.NewReplacer
// performs a single left-to-right pass and never re-scans text it has already
// emitted. Consequently, a literal "{{TITLE}}" that appears inside the body
// argument is preserved verbatim in the output — it is NOT expanded a second
// time into the real title value.
func TestBuildPageDoesNotReExpandPlaceholders(t *testing.T) {
	body := "<p>literal {{TITLE}} text</p>"
	title := "RealTitle"
	chromaCSS := ".x{}"

	page := BuildPage(body, title, chromaCSS, "")

	// The real {{TITLE}} token in the template must be replaced with the title.
	if !strings.Contains(page, "<title>RealTitle</title>") {
		t.Errorf("expected <title>RealTitle</title> in output:\n%s", page)
	}

	// The literal {{TITLE}} carried in from the body must survive unchanged;
	// strings.NewReplacer does not re-scan already-emitted output.
	if !strings.Contains(page, "literal {{TITLE}} text") {
		t.Errorf("expected body-injected literal {{TITLE}} to be preserved verbatim:\n%s", page)
	}
}

func TestBuildPageInjectsLinksFooter(t *testing.T) {
	footer := `<footer class="links"><h2>Links</h2></footer>`
	page := BuildPage("<p>x</p>", "T", "", footer)

	if !strings.Contains(page, footer) {
		t.Errorf("links footer not injected:\n%s", page)
	}

	empty := BuildPage("<p>x</p>", "T", "", "")
	if strings.Contains(empty, "{{LINKS}}") {
		t.Errorf("unreplaced LINKS placeholder remains:\n%s", empty)
	}
	if strings.Contains(empty, "<footer") {
		t.Errorf("empty links footer should leave no footer element:\n%s", empty)
	}
}
