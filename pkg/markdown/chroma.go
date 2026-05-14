package markdown

import (
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
)

// chromaStyleName is the chroma style driving code-fence highlighting.
const chromaStyleName = "tokyonight-night"

// ChromaCSS generates the class-based stylesheet for highlighted code fences.
func ChromaCSS() (string, error) {
	style := styles.Get(chromaStyleName)
	formatter := chromahtml.New(chromahtml.WithClasses(true))
	var b strings.Builder
	if err := formatter.WriteCSS(&b, style); err != nil {
		return "", err
	}
	return b.String(), nil
}
