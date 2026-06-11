# pkg/markdown

Functional core for the `markdown` (alias `md`) command. All exports are pure functions; the only I/O lives in `cmd/markdown.go` (imperative shell).

## Pipeline

```
src bytes
  └─ StripFrontmatter ─▶ ExtractTitle ─▶ RenderMarkdown ─▶ ChromaCSS ─▶ BuildPage ─▶ html string
                      └─ ExtractLinks ─▶ LinksFooter ──────────────────────┘
```

## Files

| File | Exports | Role |
|---|---|---|
| `paths.go` | `ResolveOutputPath`, `OutputTarget`, `ResolveOutputTarget` | Compute the output destination. `OutputTarget{Path, Temp}` names either a concrete path or an `os.CreateTemp` pattern. `ResolveOutputTarget` applies precedence: `--output` wins and is never temporary; `--open` alone yields a temp pattern `<base>-*.html` (nameless/dotfile inputs fall back to `"markdown"`); otherwise delegates to the unchanged sibling rule in `ResolveOutputPath`. The directory portion of the input path is stripped from the temp pattern. |
| `types.go` | `RenderConfig`, `NewRenderConfig` | Validated configuration record: `InputPath string`, `Output OutputTarget`, `Open bool`. `Open` drives the browser-open step; `Output.Temp` only selects the destination. Built from CLI args by `NewRenderConfig`. |
| `frontmatter.go` | `StripFrontmatter`, `ExtractTitle` | Strip a leading `---`-delimited YAML block. Title comes from the first non-empty ATX H1; falls back to the supplied default when none is found. |
| `convert.go` | `RenderMarkdown` | goldmark + GFM with a custom code-block renderer registered at priority 100 (beats the default 1000). `mermaid` fences pass through as `<pre class="mermaid">` (with `util.EscapeHTML` on the source); all other fences run through chroma. |
| `chroma.go` | `ChromaCSS` | Emit the class-based chroma stylesheet for the `tokyonight-night` style. |
| `links.go` | `Link`, `ExtractLinks`, `LinksFooter` | Walk the goldmark+GFM AST to collect external (`http`/`https`) inline links, reference links, autolinks, and images; deduplicated by URL, first occurrence wins, document order. Code fences produce no link nodes. `LinksFooter` renders a `<footer class="links">` with a numbered `<ol>`; label falls back to URL; images are marked `<em>(image)</em>`; returns `""` when there are no links so the placeholder collapses. |
| `page.go` | `BuildPage` | Replace `{{TITLE}}`, `{{PAGE_CSS}}`, `{{CHROMA_CSS}}`, `{{BODY}}`, `{{LINKS}}` in `template.html` in a single `strings.NewReplacer` pass. |
| `template.html` | (embedded via `//go:embed`) | HTML scaffold with the `mermaid.js` `<script type="module">` block. |
| `styles.css` | (embedded via `//go:embed`) | Tokyonight-night palette, monospace body, heading colour ramp, yellow inline code, mermaid block frame, links footer (top border, dim heading, smaller font, word-break on URLs), wide media (tables, standalone images, and mermaid blocks may grow past the 96ch text column up to `--wide: min(140ch, 100vw - 3rem)`, centered on the column; inline images stay inline). |

## Notes

- The `if !entering { return ast.WalkContinue, nil }` guard in `renderFencedCodeBlock` **must remain**. goldmark's `ast.Walk` still fires the exit pass for code blocks regardless of `WalkSkipChildren`, so the guard prevents emitting the block twice. (Reviewers occasionally flag it as dead code — it isn't.)
- `RenderMarkdown` enables `goldmarkhtml.WithUnsafe()` so the `<pre class="mermaid">` output reaches the page unescaped.
- Mermaid is loaded from `https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js` at view time; diagram rendering needs network access.
- `mermaid.initialize` sets `useMaxWidth: false` per diagram type so each SVG gets its natural pixel width. The `pre.mermaid` frame (`width: fit-content`, capped at `--wide`) then tracks the diagram instead of mermaid scaling it down to the text column; diagrams wider than the cap scroll inside the frame.
- The chroma style name is the single constant `chromaStyleName = "tokyonight-night"` in `chroma.go`; change it there to retheme highlighted code.

## Tests

Each `*.go` file has a sibling `*_test.go` covering the pure function it exports. Run them with:

```
go test ./pkg/markdown
```

See also: [`cmd/README.md`](../../cmd/README.md) — the command-level perspective (flags, output rules, end-to-end behaviour).
