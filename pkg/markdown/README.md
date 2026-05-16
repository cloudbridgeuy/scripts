# pkg/markdown

Functional core for the `markdown` (alias `md`) command. All exports are pure functions; the only I/O lives in `cmd/markdown.go` (imperative shell).

## Pipeline

```
src bytes
  └─ StripFrontmatter ─▶ ExtractTitle ─▶ RenderMarkdown ─▶ ChromaCSS ─▶ BuildPage ─▶ html string
```

## Files

| File | Exports | Role |
|---|---|---|
| `paths.go` | `ResolveOutputPath` | Compute the output path. Default swaps the source extension to `.html`; extensionless inputs gain `.html`; the `-o/--output` flag overrides verbatim. |
| `types.go` | `RenderConfig`, `NewRenderConfig` | Validated configuration record built from CLI args. |
| `frontmatter.go` | `StripFrontmatter`, `ExtractTitle` | Strip a leading `---`-delimited YAML block. Title comes from the first non-empty ATX H1; falls back to the supplied default when none is found. |
| `convert.go` | `RenderMarkdown` | goldmark + GFM with a custom code-block renderer registered at priority 100 (beats the default 1000). `mermaid` fences pass through as `<pre class="mermaid">` (with `util.EscapeHTML` on the source); all other fences run through chroma. |
| `chroma.go` | `ChromaCSS` | Emit the class-based chroma stylesheet for the `tokyonight-night` style. |
| `page.go` | `BuildPage` | Replace `{{TITLE}}`, `{{PAGE_CSS}}`, `{{CHROMA_CSS}}`, `{{BODY}}` in `template.html` in a single `strings.NewReplacer` pass. |
| `template.html` | (embedded via `//go:embed`) | HTML scaffold with the `mermaid.js` `<script type="module">` block. |
| `styles.css` | (embedded via `//go:embed`) | Tokyonight-night palette, monospace body, heading colour ramp, yellow inline code, mermaid block frame. |

## Notes

- The `if !entering { return ast.WalkContinue, nil }` guard in `renderFencedCodeBlock` **must remain**. goldmark's `ast.Walk` still fires the exit pass for code blocks regardless of `WalkSkipChildren`, so the guard prevents emitting the block twice. (Reviewers occasionally flag it as dead code — it isn't.)
- `RenderMarkdown` enables `goldmarkhtml.WithUnsafe()` so the `<pre class="mermaid">` output reaches the page unescaped.
- Mermaid is loaded from `https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js` at view time; diagram rendering needs network access.
- The chroma style name is the single constant `chromaStyleName = "tokyonight-night"` in `chroma.go`; change it there to retheme highlighted code.

## Tests

Each `*.go` file has a sibling `*_test.go` covering the pure function it exports. Run them with:

```
go test ./pkg/markdown
```

See also: [`cmd/README.md`](../../cmd/README.md) — the command-level perspective (flags, output rules, end-to-end behaviour).
