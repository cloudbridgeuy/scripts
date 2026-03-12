package report

import "strings"

// ParseActions parses a text block into a slice of Actions.
//
// Rules:
//  1. Lines starting with # → strip "# " prefix, trim whitespace, set as description for the next action
//  2. Lines ending with \ → join with next line (strip trailing \, strip leading whitespace from next, join with space)
//  3. Blank lines → skip
//  4. Any other line (or completed continuation) → becomes Command of a new Action
//  5. No preceding # comment → Description is empty string
//  6. Multiple consecutive # lines → last one wins
//  7. Trailing # comment with no following command → ignored
func ParseActions(text string) []Action {
	if text == "" {
		return []Action{}
	}

	lines := strings.Split(text, "\n")
	actions := []Action{}
	var description string

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Rule 3: skip blank lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Rule 1: comment line (trim leading whitespace before checking)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			description = strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			continue
		}

		// Rule 2 & 4: command line, possibly with continuation
		cmd := line

		for strings.HasSuffix(strings.TrimRight(cmd, " \t"), `\`) {
			cmd = strings.TrimRight(cmd, " \t")
			cmd = cmd[:len(cmd)-1]          // remove trailing '\'
			cmd = strings.TrimRight(cmd, " \t")

			i++
			if i < len(lines) {
				cmd += " " + strings.TrimLeft(lines[i], " \t")
			}
		}

		actions = append(actions, Action{
			Description: description,
			Command:     cmd,
		})
		description = ""
	}

	return actions
}
