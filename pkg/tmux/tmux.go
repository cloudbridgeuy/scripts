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
	// Check if the current session has the same name
	currentSession, err := script.Exec("tmux display-message -p '#S'").String()
	if err != nil {
		return fmt.Errorf("failed to get current session name: %w", err)
	}

	if name == currentSession {
		logger.Infof("Already in session %s", name)
		return nil
	}

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
