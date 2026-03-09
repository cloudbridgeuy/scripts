package tmux

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bitfield/script"
	"github.com/cloudbridgeuy/scripts/pkg/logger"
)

func runTmux(args ...string) error {
	cmd := exec.Command("tmux", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			return fmt.Errorf("%w: %s", err, message)
		}
		return err
	}

	return nil
}

func runTmuxOutput(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			return "", fmt.Errorf("%w: %s", err, message)
		}
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func parseNonEmptyLines(result string) []string {
	var lines []string
	for _, line := range strings.Split(result, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lines = append(lines, trimmed)
	}

	return lines
}

func isExitCode(err error, exitCode int) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}

	return exitErr.ExitCode() == exitCode
}

func isNoServerRunning(err error) bool {
	return strings.Contains(err.Error(), "no server running")
}

func canonicalSessionName(name string) string {
	return strings.ReplaceAll(name, ".", "_")
}

// ListSessions returns a list of all the running Tmux sessions
func ListSessions() ([]string, error) {
	logger.Infof("Listing all tmux sessions")
	logger.Debugf("tmux ls -F #{session_name}")
	text, err := runTmuxOutput("ls", "-F", "#{session_name}")
	if err != nil {
		if isNoServerRunning(err) || isExitCode(err, 1) {
			return []string{}, nil
		}
		return nil, err
	}

	return parseNonEmptyLines(text), nil
}

// Switch ensures that you create/switch/attach to a new session by name.
//
// The value of `name` is supposed to be a directory path.
func Switch(name string) error {
	canonical := canonicalSessionName(name)

	currentSession, err := GetCurrentSession()
	if err == nil && canonical == currentSession {
		logger.Infof("Already in session %s", canonical)
		return nil
	}

	if err := HasSession(canonical); err != nil {
		if createErr := NewSession(name); createErr != nil {
			return createErr
		}
	}

	return switchToCanonicalSession(canonical)
}

// SwitchExisting switches to an existing session without creating it.
func SwitchExisting(name string) error {
	canonical := canonicalSessionName(name)

	currentSession, err := GetCurrentSession()
	if err == nil && canonical == currentSession {
		logger.Infof("Already in session %s", canonical)
		return nil
	}

	if err := HasSession(canonical); err != nil {
		return err
	}

	return switchToCanonicalSession(canonical)
}

func switchToCanonicalSession(canonical string) error {

	if err := SwitchClient(canonical); err == nil {
		return nil
	}

	return Attach(canonical)
}

// SwitchClient switches the client to the given session.
func SwitchClient(name string) error {
	logger.Infof("Switching to session %s", name)
	logger.Debugf("tmux switch-client -t %s", name)
	return runTmux("switch-client", "-t", name)
}

// Attach attaches the current tmux instance to the given session.
//
// NOTE:
// We can't use the `scripts` package because (for some unknown reason to me)
// tmux` requires that we bind `stdout`, `stderr`, and `stdin` to the spawned
// process for it to work.
func Attach(name string) error {
	logger.Infof("Attaching to session %s", name)
	logger.Debugf("tmux attach -t %s", name)
	cmd := exec.Command("tmux", "attach", "-d", "-t", name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// NewSession creates a new tmux session.
func NewSession(name string) error {
	canonical := canonicalSessionName(name)

	logger.Infof("Creating new session %s", name)
	logger.Debugf("tmux new-session -s %s -c %s -d", canonical, name)
	return runTmux("new-session", "-s", canonical, "-c", name, "-d")
}

// KillSessions kills a session.
func KillSession(name string) error {
	canonical := canonicalSessionName(name)

	logger.Infof("Killing session %s", name)
	if err := HasSession(canonical); err == nil {
		logger.Debugf("tmux kill-session -t %s", canonical)
		return runTmux("kill-session", "-t", canonical)
	}
	return nil
}

// HasSession checks if the given session exists.
func HasSession(name string) error {
	canonical := canonicalSessionName(name)

	logger.Infof("Checking if session %s exists", name)
	logger.Debugf("tmux has-session -t %s", canonical)
	return runTmux("has-session", "-t", canonical)
}

// SessionExists returns whether a session exists.
func SessionExists(name string) (bool, error) {
	err := HasSession(name)
	if err == nil {
		return true, nil
	}

	if isExitCode(err, 1) || isNoServerRunning(err) {
		return false, nil
	}

	return false, err
}

// DisplaySessions dynamically renders all the current active sessions and allows you to traverse to them.
func DisplaySessions() (string, error) {
	fzfCmd := fmt.Sprintf(`fzf \
      --header 'Press CTRL-X to delete a session.' \
      --bind "ctrl-x:execute-silent(tmux kill-session -t {})+reload(tmux ls -F'#{session_name}')" \
      --preview "tmux capture-pane -ep -t \"\$(tmux ls -F '#{session_id}' -f '#{==:#{session_name},{}}')\"" --preview-window="right:70%%" --height="100%%"`)

	buf, err := script.
		Exec("tmux ls -F'#{session_name}'").
		Exec("sort -h").
		Exec(fzfCmd).
		WithStderr(os.Stdout).
		String()

	return strings.TrimSpace(buf), err
}

// Ls returns a list of `tmux` running sessions.
func Ls() ([]string, error) {
	result, err := runTmuxOutput("ls", "-F", "#{session_name}")
	if err != nil {
		if isNoServerRunning(err) || isExitCode(err, 1) {
			return []string{}, nil
		}
		return nil, err
	}

	return parseNonEmptyLines(result), nil
}

// GetCurrentSession returns the name of the current tmux session.
func GetCurrentSession() (string, error) {
	session, err := runTmuxOutput("display-message", "-p", "#S")
	if err != nil {
		return "", err
	}
	return session, nil
}

// ListWindows returns a list of window IDs in the current session.
func ListWindows() ([]string, error) {
	result, err := runTmuxOutput("list-windows", "-F", "#{window_id}")
	if err != nil {
		return nil, err
	}

	return parseNonEmptyLines(result), nil
}

// NewWindow creates a new window with the given name, command, and directory.
func NewWindow(name, command, directory string) error {
	logger.Infof("Creating new window %s", name)
	logger.Debugf("tmux new-window -n %s -c %s %s", name, directory, command)
	return runTmux("new-window", "-n", name, "-c", directory, command)
}

// KillWindow kills a window by its ID.
func KillWindow(windowID string) error {
	logger.Infof("Killing window %s", windowID)
	logger.Debugf("tmux kill-window -t %s", windowID)
	return runTmux("kill-window", "-t", windowID)
}

// SelectWindow selects (focuses) a window by name.
func SelectWindow(name string) error {
	logger.Infof("Selecting window %s", name)
	logger.Debugf("tmux select-window -t %s", name)
	return runTmux("select-window", "-t", name)
}
