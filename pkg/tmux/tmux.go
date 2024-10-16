package tmux

import (
	"fmt"
	"os"
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

	_, err := script.Exec(fmt.Sprintf("tmux -L %s capture-pane -ep -t %s", s.server, s.id)).Stdout()
	return err
}

// ListSessions returns a list of all the running Tmux sessions
func ListSessions() ([]string, error) {
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
	if err := HasSession(name); err != nil {
		return NewSession(name)
	}

	if err := Attach(name); err != nil {
		return SwitchClient(name)
	}

	return nil
}

// SwitchClient switches the client to the given session.
func SwitchClient(name string) error {
	logger.Debugf("Switching to session %s", name)
	return script.Exec(fmt.Sprintf("tmux switch-client -t %s", name)).Wait()
}

// Attach attaches the current tmux instance to the given session.
func Attach(name string) error {
	logger.Debugf("Attaching to session %s", name)
	return script.Exec(fmt.Sprintf("tmux attach -t '=%s'", name)).Wait()
}

// NewSession creates a new tmux session.
func NewSession(name string) error {
	logger.Debugf("Creating new session %s", name)
	return script.Exec(fmt.Sprintf("tmux new-session -s %s -c %s -d", strings.Replace(name, ".", "·", -1), name)).Wait()
}

// HasSession checks if the given session exists.
func HasSession(name string) error {
	logger.Debugf("Checking if session %s exists", name)
	return script.Exec(fmt.Sprintf("tmux has-session -t '=%s'", name)).Wait()
}

// DisplaySessions dynamically renders all the current active sessions and allows you to traverse to them.
func DisplaySessions() (string, error) {
	buf, err := script.
		Exec("tmux ls -F'#{session_name}'").
		Exec("sort -h").
		Exec(`fzf \
      --header 'Press CTRL-X to delete a session.' \
      --bind "ctrl-x:execute-silent(tmux kill-session -t {+})+reload(tmux ls -F'#{session_name}')" \
      --preview "tmux capture-pane -ep -t \"\$(tmux ls -F '#{session_id}' -f '#{==:#{session_name},{}}')\"" --preview-window="right:70%" --height="100%"`).
		WithStderr(os.Stdout).
		String()

	return strings.TrimSpace(buf), err
}
