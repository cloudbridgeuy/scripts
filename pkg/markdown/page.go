/*
Copyright © 2024 Guzmán Monné guzman.monne@cloudbridge.com.uy

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package markdown

import (
	_ "embed"
	"html"
	"strings"
)

//go:embed template.html
var pageTemplate string

//go:embed styles.css
var pageCSS string

// BuildPage assembles a complete HTML document from a rendered body fragment,
// a page title, a generated chroma stylesheet, and an optional links footer
// (pass "" for documents without external links).
func BuildPage(body, title, chromaCSS, linksHTML string) string {
	return strings.NewReplacer(
		"{{TITLE}}", html.EscapeString(title),
		"{{PAGE_CSS}}", pageCSS,
		"{{CHROMA_CSS}}", chromaCSS,
		"{{BODY}}", body,
		"{{LINKS}}", linksHTML,
	).Replace(pageTemplate)
}
