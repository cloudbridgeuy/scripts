package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bitfield/script"
	"github.com/cloudbridgeuy/scripts/pkg/logger"
)

type session struct {
	id     string
	server string
	name   string
}

type sessionBuilder struct {
	inner *session
}

// NewSessionBuilder creates a new session
func NewSessionBuilder(name string) *sessionBuilder {
	var builder sessionBuilder
	builder.inner.name = name
	builder.inner.server = "default"

	return &builder
}

// WithName sets the name for the session
func (b *sessionBuilder) WithName(name string) *sessionBuilder {
	b.inner.name = name
	return b
}

// WithId sets the id for the session
func (b *sessionBuilder) WithId(id string) *sessionBuilder {
	b.inner.id = id
	return b
}

// WithServer sets the server for the session
func (b *sessionBuilder) WithServer(server string) *sessionBuilder {
	b.inner.server = server
	return b
}

// Build returns the inner struct
func (b *sessionBuilder) Build() (*session, error) {
	id, err := script.Exec(fmt.Sprintf("tmux -L %s ls -F '#{session_id}' -f \"#{==:#{session_name},%s}\"", b.inner.server, b.inner.name)).String()
	if err != nil {
		return nil, err
	}
	b.inner.id = id

	return b.inner, nil
}

// Display shows the state of the current session
func (s *session) Display() error {
	if s.id == "" {
		return fmt.Errorf("can't find `id` for session %s", s.name)
	}

	logger.Infof("Displaying session %s", s.name)
	logger.Debugf("tmux -L %s capture-pane -ep -t %s", s.server, s.id)
	_, err := script.Exec(fmt.Sprintf("tmux -L %s capture-pane -ep -t %s", s.server, s.id)).Stdout()
	return err
}

// ListSessions returns a list of all the running Tmux sessions
func ListSessions() ([]string, error) {
	logger.Infof("Listing all tmux sessions")
	logger.Debugf("tmux ls -F'#{session_name}'")
	text, err := script.Exec("tmux ls -F'#{session_name}'").String()
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(text), "\n"), nil
}

// Switch ensures that you create/switch/attach to a new session by name.
//
// The value of `name` is supposed to be a directory path.
func Switch(name string) error {
	// Check if we're currently in a tmux session
	currentSession, err := script.Exec("tmux display-message -p '#S'").String()
	if err == nil {
		// We're in a tmux session, check if it's the same one
		if name == currentSession {
			logger.Infof("Already in session %s", name)
			return nil
		}
	}
	// If err != nil, we're not in a tmux session, which is fine - proceed to create/attach

	if err := HasSession(name); err != nil {
		NewSession(name)
	}

	if err := Attach(name); err != nil {
		return SwitchClient(name)
	}

	return nil
}

// SwitchClient switches the client to the given session.
func SwitchClient(name string) error {
	logger.Infof("Switching to session %s", name)
	logger.Debugf("tmux switch-client -t %s", name)
	return script.Exec(fmt.Sprintf("tmux switch-client -t %s", name)).Wait()
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
	cmd := exec.Command("tmux", "attach", "-d", "-t", fmt.Sprintf("=%s", name))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	return cmd.Wait()
}

// NewSession creates a new tmux session.
func NewSession(name string) error {
	logger.Infof("Creating new session %s", name)
	logger.Debugf("tmux new-session -s %s -c %s -d", strings.Replace(name, ".", "·", -1), name)
	return script.Exec(fmt.Sprintf("tmux new-session -s %s -c %s -d", strings.Replace(name, ".", "·", -1), name)).Wait()
}

// KillSessions kills a session.
func KillSession(name string) error {
	logger.Infof("Killing session %s", name)
	if err := HasSession(name); err == nil {
		logger.Debugf("tmux kill-session -t %s", name)
		return script.Exec(fmt.Sprintf("tmux kill-session -t %s", name)).Wait()
	}
	return nil
}

// HasSession checks if the given session exists.
func HasSession(name string) error {
	logger.Infof("Checking if session %s exists", name)
	logger.Debugf("tmux has-session -t %s", name)
	return script.Exec(fmt.Sprintf("tmux has-session -t %s", name)).Wait()
}

// DisplaySessions dynamically renders all the current active sessions and allows you to traverse to them.
func DisplaySessions() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		execPath = "scripts" // fallback
	}

	fzfCmd := fmt.Sprintf(`fzf \
      --header 'Press CTRL-X to delete a session.' \
      --bind "ctrl-x:execute-silent(%s tmux remove {})+reload(tmux ls -F'#{session_name}')" \
      --preview "tmux capture-pane -ep -t \"\$(tmux ls -F '#{session_id}' -f '#{==:#{session_name},{}}')\"" --preview-window="right:70%%" --height="100%%"`, execPath)

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
	result, err := script.
		Exec("tmux ls -F'#{session_name}'").
		String()
	if err != nil {
		return nil, err
	}

	var sessions []string
	for _, session := range strings.Split(result, "\n") {
		if session == "" {
			continue
		}
		sessions = append(sessions, strings.TrimSpace(session))
	}

	return sessions, nil
}

// GetCurrentSession returns the name of the current tmux session.
func GetCurrentSession() (string, error) {
	session, err := script.Exec("tmux display-message -p '#S'").String()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(session), nil
}

// ListWindows returns a list of window IDs in the current session.
func ListWindows() ([]string, error) {
	result, err := script.
		Exec("tmux list-windows -F'#{window_id}'").
		String()
	if err != nil {
		return nil, err
	}

	var windows []string
	for _, window := range strings.Split(result, "\n") {
		window = strings.TrimSpace(window)
		if window == "" {
			continue
		}
		windows = append(windows, window)
	}

	return windows, nil
}

// NewWindow creates a new window with the given name, command, and directory.
func NewWindow(name, command, directory string) error {
	logger.Infof("Creating new window %s", name)
	logger.Debugf("tmux new-window -n %s -c %s %s", name, directory, command)
	return script.Exec(fmt.Sprintf("tmux new-window -n %s -c %s %s", name, directory, command)).Wait()
}

// KillWindow kills a window by its ID.
func KillWindow(windowID string) error {
	logger.Infof("Killing window %s", windowID)
	logger.Debugf("tmux kill-window -t %s", windowID)
	return script.Exec(fmt.Sprintf("tmux kill-window -t %s", windowID)).Wait()
}

// SelectWindow selects (focuses) a window by name.
func SelectWindow(name string) error {
	logger.Infof("Selecting window %s", name)
	logger.Debugf("tmux select-window -t %s", name)
	return script.Exec(fmt.Sprintf("tmux select-window -t %s", name)).Wait()
}
