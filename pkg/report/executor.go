package report

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// ResolveInput determines the input text from the available sources.
// Priority: filePath > piped stdin > args > error.
func ResolveInput(stdin io.Reader, filePath string, args []string, isInputTTY bool) (string, error) {
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("reading file %s: %w", filePath, err)
		}
		return string(data), nil
	}

	if !isInputTTY {
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return string(data), nil
	}

	if len(args) > 0 {
		return strings.Join(args, "\n"), nil
	}

	return "", errors.New("no input provided")
}

// exitCode returns the exit code from a command execution error.
// Returns 0 for nil errors, the actual exit code for ExitError, or 1 for other errors.
func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}

// ExecuteActions runs each action using the given shell and collects results.
// If onError is Stop and a command fails, execution halts and partial results are returned.
func ExecuteActions(actions []Action, onError OnErrorBehavior, shell string) []Result {
	results := make([]Result, 0, len(actions))

	for _, action := range actions {
		cmd := exec.Command(shell, "-ic", action.Command)
		output, err := cmd.CombinedOutput()
		code := exitCode(err)

		results = append(results, Result{
			Action:   action,
			ExitCode: code,
			Output:   string(output),
		})

		if onError == Stop && code != 0 {
			break
		}
	}

	return results
}
