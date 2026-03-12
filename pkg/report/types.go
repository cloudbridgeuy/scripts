package report

import "fmt"

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

// ParseFormat validates and returns a Format from a string.
func ParseFormat(s string) (Format, error) {
	switch Format(s) {
	case XML, Markdown:
		return Format(s), nil
	default:
		return "", fmt.Errorf("unsupported format: %q (expected %q or %q)", s, XML, Markdown)
	}
}

// ParseOnErrorBehavior validates and returns an OnErrorBehavior from a string.
func ParseOnErrorBehavior(s string) (OnErrorBehavior, error) {
	switch OnErrorBehavior(s) {
	case Continue, Stop:
		return OnErrorBehavior(s), nil
	default:
		return "", fmt.Errorf("unsupported on-error behavior: %q (expected %q or %q)", s, Continue, Stop)
	}
}

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
