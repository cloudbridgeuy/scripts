package report

// Format represents the output format for the report.
type Format string

const (
	XML      Format = "xml"
	Markdown Format = "md"
)

// OnErrorBehavior controls what happens when a command fails.
type OnErrorBehavior string

const (
	Continue OnErrorBehavior = "continue"
	Stop     OnErrorBehavior = "stop"
)

// Action represents a parsed command with its description.
type Action struct {
	Description string
	Command     string
}

// Result represents the outcome of executing an Action.
type Result struct {
	Action   Action
	ExitCode int
	Output   string
}
