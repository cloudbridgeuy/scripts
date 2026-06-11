package markdown

import (
	"fmt"
	"html"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// Link is one external reference found in a Markdown document.
type Link struct {
	Text    string
	URL     string
	IsImage bool
}

// ExtractLinks parses src with the same GFM parser used for rendering and
// returns every external (http/https) link, autolink, and image, in document
// order, deduplicated by URL with the first occurrence winning (a URL appearing
// as both link and image keeps only its first form). Code fences produce no
// link nodes, so their contents are ignored by construction.
func ExtractLinks(src []byte) []Link {
	parser := goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser()
	root := parser.Parse(text.NewReader(src))

	var links []Link
	seen := map[string]bool{}

	add := func(textValue, url string, isImage bool) {
		if !isExternalURL(url) || seen[url] {
			return
		}
		seen[url] = true
		links = append(links, Link{Text: textValue, URL: url, IsImage: isImage})
	}

	_ = ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := node.(type) {
		case *ast.Link:
			add(nodeText(n, src), string(n.Destination), false)
		case *ast.AutoLink:
			add(string(n.Label(src)), string(n.URL(src)), false)
		case *ast.Image:
			add(nodeText(n, src), string(n.Destination), true)
		}
		return ast.WalkContinue, nil
	})

	return links
}

// isExternalURL reports whether url points at an http or https destination.
func isExternalURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// nodeText concatenates the plain text content beneath a node — the link
// label or the image alt text. Soft line breaks inside a label are dropped,
// not replaced with spaces.
func nodeText(node ast.Node, src []byte) string {
	var b strings.Builder
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch t := n.(type) {
		case *ast.Text:
			b.Write(t.Segment.Value(src))
		case *ast.String:
			b.Write(t.Value)
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

// LinksFooter renders the Links footer section for a page. An empty slice
// yields an empty string so the template placeholder collapses cleanly.
func LinksFooter(links []Link) string {
	if len(links) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("<footer class=\"links\">\n<h2>Links</h2>\n<ol>\n")
	for _, l := range links {
		label := l.Text
		if label == "" {
			label = l.URL
		}
		url := html.EscapeString(l.URL)
		fmt.Fprintf(&b, `<li><a href="%s">%s</a> — <span class="url">%s</span>`, url, html.EscapeString(label), url)
		if l.IsImage {
			b.WriteString(" <em>(image)</em>")
		}
		b.WriteString("</li>\n")
	}
	b.WriteString("</ol>\n</footer>\n")
	return b.String()
}
