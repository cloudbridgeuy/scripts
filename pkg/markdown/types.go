package markdown

// RenderConfig holds the fully resolved configuration for one render run.
type RenderConfig struct {
	InputPath  string
	OutputPath string
	Open       bool
}

// NewRenderConfig resolves raw CLI inputs into a valid RenderConfig.
// The output path is computed once, at the boundary, so the rest of the
// program holds only valid state.
func NewRenderConfig(inputPath, outputFlag string, open bool) RenderConfig {
	return RenderConfig{
		InputPath:  inputPath,
		OutputPath: ResolveOutputPath(inputPath, outputFlag),
		Open:       open,
	}
}
