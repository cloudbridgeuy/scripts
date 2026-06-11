package markdown

import (
	"strings"
	"testing"
)

func TestExtractLinksInlineAndImage(t *testing.T) {
	src := []byte(`# Doc

See [Goldmark docs](https://github.com/yuin/goldmark) and
![diagram](https://example.com/diagram.png).
`)
	links := ExtractLinks(src)

	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d: %#v", len(links), links)
	}
	if links[0].Text != "Goldmark docs" || links[0].URL != "https://github.com/yuin/goldmark" || links[0].IsImage {
		t.Errorf("unexpected first link: %#v", links[0])
	}
	if links[1].Text != "diagram" || links[1].URL != "https://example.com/diagram.png" || !links[1].IsImage {
		t.Errorf("unexpected second link: %#v", links[1])
	}
}

func TestExtractLinksReferenceStyle(t *testing.T) {
	src := []byte(`See [the spec][1].

[1]: https://spec.example.com/v2
`)
	links := ExtractLinks(src)

	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d: %#v", len(links), links)
	}
	if links[0].Text != "the spec" || links[0].URL != "https://spec.example.com/v2" {
		t.Errorf("unexpected link: %#v", links[0])
	}
}

func TestExtractLinksAutolink(t *testing.T) {
	// GFM linkify turns bare URLs into autolinks.
	src := []byte("Visit https://example.com/page for details.\n")
	links := ExtractLinks(src)

	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d: %#v", len(links), links)
	}
	if links[0].URL != "https://example.com/page" {
		t.Errorf("unexpected link: %#v", links[0])
	}
}

func TestExtractLinksSkipsNonHTTP(t *testing.T) {
	src := []byte(`[anchor](#section) [rel](./other.md) [mail](mailto:a@b.c) [ftp](ftp://x.y)
[ok](https://keep.me)
`)
	links := ExtractLinks(src)

	if len(links) != 1 || links[0].URL != "https://keep.me" {
		t.Fatalf("expected only https://keep.me, got %#v", links)
	}
}

func TestExtractLinksDedupesByURLPreservingOrder(t *testing.T) {
	src := []byte(`[first](https://a.example) [second](https://b.example) [again](https://a.example)
`)
	links := ExtractLinks(src)

	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d: %#v", len(links), links)
	}
	if links[0].URL != "https://a.example" || links[0].Text != "first" {
		t.Errorf("first occurrence should win: %#v", links[0])
	}
	if links[1].URL != "https://b.example" {
		t.Errorf("order not preserved: %#v", links)
	}
}

func TestExtractLinksIgnoresCodeFences(t *testing.T) {
	src := []byte("```\n[fake](https://fenced.example)\nhttps://bare.fenced.example\n```\n\n`[inline](https://code.example)`\n")
	links := ExtractLinks(src)

	if len(links) != 0 {
		t.Fatalf("expected no links from code, got %#v", links)
	}
}

func TestExtractLinksEmptyInput(t *testing.T) {
	if links := ExtractLinks(nil); len(links) != 0 {
		t.Fatalf("expected no links, got %#v", links)
	}
}

func TestLinksFooterEmpty(t *testing.T) {
	if got := LinksFooter(nil); got != "" {
		t.Fatalf("expected empty footer for nil, got %q", got)
	}
	if got := LinksFooter([]Link{}); got != "" {
		t.Fatalf("expected empty footer for empty slice, got %q", got)
	}
}

func TestExtractLinksBadgeImageInLink(t *testing.T) {
	src := []byte(`[![build status](https://img.example/badge.svg)](https://ci.example/run)
`)
	links := ExtractLinks(src)

	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d: %#v", len(links), links)
	}
	if links[0].URL != "https://ci.example/run" || links[0].IsImage {
		t.Errorf("expected outer link first: %#v", links[0])
	}
	if links[1].URL != "https://img.example/badge.svg" || !links[1].IsImage {
		t.Errorf("expected inner image second: %#v", links[1])
	}
}

func TestLinksFooterRendersEntries(t *testing.T) {
	footer := LinksFooter([]Link{
		{Text: "Goldmark docs", URL: "https://github.com/yuin/goldmark"},
		{Text: "diagram", URL: "https://example.com/diagram.png", IsImage: true},
	})

	for _, want := range []string{
		`<footer class="links">`,
		"<h2>Links</h2>",
		"<ol>",
		`<li><a href="https://github.com/yuin/goldmark">Goldmark docs</a> — <span class="url">https://github.com/yuin/goldmark</span></li>`,
		`<li><a href="https://example.com/diagram.png">diagram</a> — <span class="url">https://example.com/diagram.png</span> <em>(image)</em></li>`,
		"</ol>",
		"</footer>",
	} {
		if !strings.Contains(footer, want) {
			t.Errorf("footer missing %q:\n%s", want, footer)
		}
	}
}

func TestLinksFooterFallsBackToURLLabel(t *testing.T) {
	footer := LinksFooter([]Link{{URL: "https://example.com"}})

	if !strings.Contains(footer, `<a href="https://example.com">https://example.com</a>`) {
		t.Errorf("expected URL used as label:\n%s", footer)
	}
}

func TestLinksFooterEscapesHTML(t *testing.T) {
	footer := LinksFooter([]Link{{Text: `<b>"bold"</b>`, URL: `https://example.com/?a=1&b=2`}})

	if strings.Contains(footer, "<b>") {
		t.Errorf("text not escaped:\n%s", footer)
	}
	if !strings.Contains(footer, "&lt;b&gt;") {
		t.Errorf("expected escaped text:\n%s", footer)
	}
	if !strings.Contains(footer, "https://example.com/?a=1&amp;b=2") {
		t.Errorf("expected escaped URL:\n%s", footer)
	}
}

func TestExtractLinksLabelSoftLineBreak(t *testing.T) {
	src := []byte("[hello\nworld](https://example.com)\n")
	links := ExtractLinks(src)

	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d: %#v", len(links), links)
	}
	// Soft line breaks are dropped, not replaced with spaces.
	if links[0].Text != "helloworld" {
		t.Errorf("unexpected label: %q", links[0].Text)
	}
}

func TestExtractLinksDedupesAcrossLinkAndImage(t *testing.T) {
	src := []byte(`[link](https://same.example) and ![img](https://same.example)
`)
	links := ExtractLinks(src)

	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d: %#v", len(links), links)
	}
	if links[0].IsImage {
		t.Errorf("first occurrence (link) should win: %#v", links[0])
	}
}
