package markdown

// RenderConfig holds the fully resolved configuration for one render run.
// Open drives the browser-open step in the shell; it is distinct from
// Output.Temp, which only selects the destination.
type RenderConfig struct {
	InputPath string
	Output    OutputTarget
	Open      bool
}

// NewRenderConfig resolves raw CLI inputs into a valid RenderConfig.
// The output destination is decided once, at the boundary, so the rest of
// the program holds only valid state.
func NewRenderConfig(inputPath, outputFlag string, open bool) RenderConfig {
	return RenderConfig{
		InputPath: inputPath,
		Output:    ResolveOutputTarget(inputPath, outputFlag, open),
		Open:      open,
	}
}
