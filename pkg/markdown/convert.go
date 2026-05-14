package markdown

import (
	"bytes"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// codeBlockRenderer renders fenced code blocks. A fence tagged "mermaid"
// passes through for client-side rendering; every other fence is
// syntax-highlighted with chroma.
type codeBlockRenderer struct{}

func (r *codeBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
}

func (r *codeBlockRenderer) renderFencedCodeBlock(
	w util.BufWriter, source []byte, node ast.Node, entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.FencedCodeBlock)

	var code bytes.Buffer
	for i := 0; i < n.Lines().Len(); i++ {
		seg := n.Lines().At(i)
		code.Write(seg.Value(source))
	}

	language := string(n.Language(source))

	if language == "mermaid" {
		_, _ = w.WriteString(`<pre class="mermaid">`)
		_, _ = w.Write(util.EscapeHTML(code.Bytes()))
		_, _ = w.WriteString("</pre>\n")
		return ast.WalkSkipChildren, nil
	}

	if err := highlightCode(w, code.String(), language); err != nil {
		return ast.WalkStop, err
	}
	return ast.WalkSkipChildren, nil
}

// highlightCode writes class-based, chroma-highlighted HTML for a code block.
func highlightCode(w util.BufWriter, code, language string) error {
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return err
	}
	formatter := chromahtml.New(chromahtml.WithClasses(true))
	style := styles.Get(chromaStyleName)
	return formatter.Format(w, style, iterator)
}

// RenderMarkdown converts Markdown source into an HTML body fragment.
func RenderMarkdown(src []byte) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(
			goldmarkhtml.WithUnsafe(),
			renderer.WithNodeRenderers(
				util.Prioritized(&codeBlockRenderer{}, 100),
			),
		),
	)
	var b bytes.Buffer
	if err := md.Convert(src, &b); err != nil {
		return "", err
	}
	return b.String(), nil
}
