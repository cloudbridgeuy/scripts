package report

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// xmlReport is the top-level XML structure for marshalling.
type xmlReport struct {
	XMLName xml.Name    `xml:"report"`
	Actions []xmlAction `xml:"action"`
}

// xmlAction represents a single action in the XML output.
type xmlAction struct {
	Description string `xml:"description"`
	Command     string `xml:"command"`
	Status      int    `xml:"status"`
	Output      string `xml:"output"`
}

// FormatXML formats the results as an XML report using encoding/xml for proper escaping.
func FormatXML(results []Result) (string, error) {
	report := xmlReport{
		Actions: make([]xmlAction, len(results)),
	}

	for i, r := range results {
		report.Actions[i] = xmlAction{
			Description: r.Action.Description,
			Command:     r.Action.Command,
			Status:      r.ExitCode,
			Output:      "\n" + strings.TrimRight(r.Output, "\n") + "\n    ",
		}
	}

	out, err := xml.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("xml marshal: %w", err)
	}
	return string(out), nil
}

// FormatMarkdown formats the results as a Markdown report.
func FormatMarkdown(results []Result) string {
	var b strings.Builder

	b.WriteString("# Report")

	for i, r := range results {
		fmt.Fprintf(&b, "\n\n## Command %d\n", i+1)

		if r.Action.Description != "" {
			fmt.Fprintf(&b, "\n%s\n", r.Action.Description)
		}

		fmt.Fprintf(&b, "\n**Status Code**: %d\n", r.ExitCode)

		fmt.Fprintf(&b, "\n```\n%s\n```\n", r.Action.Command)

		output := strings.TrimRight(r.Output, "\n")
		fmt.Fprintf(&b, "\n**Output**:\n\n```\n%s\n```", output)
	}

	return b.String()
}

// FormatReport dispatches to the appropriate formatter based on the format.
func FormatReport(results []Result, format Format) (string, error) {
	switch format {
	case XML:
		return FormatXML(results)
	case Markdown:
		return FormatMarkdown(results), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}
