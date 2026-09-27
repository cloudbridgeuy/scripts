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

const tmuxSocketEnvVar = "SCRIPTS_TMUX_SOCKET"

func tmuxSocketArgs(args ...string) []string {
	if socket := os.Getenv(tmuxSocketEnvVar); socket != "" {
		return append([]string{"-S", socket}, args...)
	}
	return args
}

func tmuxShellCmd() string {
	if socket := os.Getenv(tmuxSocketEnvVar); socket != "" {
		return "tmux -S " + shellQuote(socket)
	}
	return "tmux"
}

func ParseTarget(arg string) (host, session string) {
	i := strings.Index(arg, ":")
	if i <= 0 {
		return "", arg
	}
	host = arg[:i]
	if strings.Contains(host, "/") {
		return "", arg
	}
	return host, arg[i+1:]
}

func HistoryName(host, session string) string {
	if host == "" {
		return session
	}
	return host + ":" + session
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

func shellJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = shellQuote(arg)
	}
	return strings.Join(quoted, " ")
}

func remotePaneCommand(host, dir, localSocket string) string {
	remoteCmd := shellQuote("cd " + shellQuote(dir) + ` && "$SHELL" -l; rm -f "$SCRIPTS_TMUX_SOCKET"`)
	forward := `"$sock:"` + shellQuote(localSocket)
	script := `sock=/tmp/scripts-tmux-$$.sock; exec ssh -t -R ` + forward + " " + host +
		` "export SCRIPTS_TMUX_SOCKET=$sock; " ` + remoteCmd
	return "sh -c " + shellQuote(script)
}

func runTmux(args ...string) error {
	cmd := exec.Command("tmux", tmuxSocketArgs(args...)...)
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
	cmd := exec.Command("tmux", tmuxSocketArgs(args...)...)
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

func CanonicalSessionName(name string) string {
	name = strings.ReplaceAll(name, ".", "_")
	return strings.ReplaceAll(name, ":", "_")
}

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

func Switch(name string) error {
	canonical := CanonicalSessionName(name)

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

func SwitchExisting(name string) error {
	canonical := CanonicalSessionName(name)

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

func SwitchClient(name string) error {
	logger.Infof("Switching to session %s", name)
	logger.Debugf("tmux switch-client -t %s", name)
	return runTmux("switch-client", "-t", name)
}

func Attach(name string) error {
	logger.Infof("Attaching to session %s", name)
	logger.Debugf("tmux attach -t %s", name)
	cmd := exec.Command("tmux", tmuxSocketArgs("attach", "-d", "-t", name)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func NewSession(name string) error {
	canonical := CanonicalSessionName(name)

	if host, dir := ParseTarget(name); host != "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		logger.Infof("Creating remote session %s", name)
		if err := runTmux("new-session", "-d", "-s", canonical, "-c", home); err != nil {
			return err
		}
		localSocket, err := runTmuxOutput("display-message", "-p", "-t", canonical, "#{socket_path}")
		if err != nil {
			return err
		}
		paneCmd := remotePaneCommand(host, dir, localSocket)
		if err := runTmux("set-option", "-w", "-t", canonical, "default-command", paneCmd); err != nil {
			return err
		}
		if err := runTmux("respawn-pane", "-k", "-t", canonical, paneCmd); err != nil {
			return err
		}
		if err := runTmux("set-environment", "-t", canonical, "SCRIPTS_REMOTE_HOST", host); err != nil {
			return err
		}
		return runTmux("set-environment", "-t", canonical, "SCRIPTS_REMOTE_DIR", dir)
	}

	logger.Infof("Creating new session %s", name)
	logger.Debugf("tmux new-session -s %s -c %s -d", canonical, name)
	return runTmux("new-session", "-s", canonical, "-c", name, "-d")
}

func KillSession(name string) error {
	canonical := CanonicalSessionName(name)

	logger.Infof("Killing session %s", name)
	if err := HasSession(canonical); err == nil {
		logger.Debugf("tmux kill-session -t %s", canonical)
		return runTmux("kill-session", "-t", canonical)
	}
	return nil
}

func HasSession(name string) error {
	canonical := CanonicalSessionName(name)

	logger.Infof("Checking if session %s exists", name)
	logger.Debugf("tmux has-session -t %s", canonical)
	return runTmux("has-session", "-t", canonical)
}

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

func parseRemoteEnv(output string) (string, string, bool) {
	var host, dir string
	var hasHost, hasDir bool

	for _, line := range strings.Split(output, "\n") {
		if value, found := strings.CutPrefix(line, "SCRIPTS_REMOTE_HOST="); found {
			host, hasHost = value, true
		}
		if value, found := strings.CutPrefix(line, "SCRIPTS_REMOTE_DIR="); found {
			dir, hasDir = value, true
		}
	}

	if !hasHost || !hasDir || host == "" {
		return "", "", false
	}

	return host, dir, true
}

func RemoteInfo(session string) (string, string, bool) {
	output, err := runTmuxOutput("show-environment", "-t", CanonicalSessionName(session))
	if err != nil {
		return "", "", false
	}

	return parseRemoteEnv(output)
}

func DisplaySessions() (string, error) {
	tmuxCmd := tmuxShellCmd()
	fzfCmd := fmt.Sprintf(`fzf \
      --header 'Press CTRL-X to delete a session.' \
      --bind "ctrl-x:execute-silent(%[1]s kill-session -t {})+reload(%[1]s ls -F'#{session_name}')" \
      --preview "%[1]s capture-pane -ep -t \"\$(%[1]s ls -F '#{session_id}' -f '#{==:#{session_name},{}}')\"" --preview-window="right:70%%" --height="100%%"`, tmuxCmd)

	buf, err := script.
		Exec(tmuxCmd + " ls -F'#{session_name}'").
		Exec("sort -h").
		Exec(fzfCmd).
		WithStderr(os.Stdout).
		String()

	return strings.TrimSpace(buf), err
}

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

func GetCurrentSession() (string, error) {
	session, err := runTmuxOutput("display-message", "-p", "#S")
	if err != nil {
		return "", err
	}
	return session, nil
}

func ListWindows() ([]string, error) {
	result, err := runTmuxOutput("list-windows", "-F", "#{window_id}")
	if err != nil {
		return nil, err
	}

	return parseNonEmptyLines(result), nil
}

func NewWindow(name, command, directory string) error {
	logger.Infof("Creating new window %s", name)
	logger.Debugf("tmux new-window -n %s -c %s %s", name, directory, command)
	return runTmux("new-window", "-n", name, "-c", directory, command)
}

func KillWindow(windowID string) error {
	logger.Infof("Killing window %s", windowID)
	logger.Debugf("tmux kill-window -t %s", windowID)
	return runTmux("kill-window", "-t", windowID)
}

func SelectWindow(name string) error {
	logger.Infof("Selecting window %s", name)
	logger.Debugf("tmux select-window -t %s", name)
	return runTmux("select-window", "-t", name)
}

type HarnessPane struct {
	SessionID    string
	SessionName  string
	WindowID     string
	WindowIndex  string
	WindowName   string
	PaneID       string
	PaneIndex    string
	Harness      string
	State        string
	Title        string
	Command      string
	StartCommand string
	Path         string
}

func HarnessSymbol(state string) string {
	switch state {
	case "working":
		return "●"
	case "waiting-input":
		return "?"
	case "waiting-permission":
		return "!"
	case "done":
		return "✓"
	default:
		return "·"
	}
}

func ShortenHome(path string) string {
	home := os.Getenv("HOME")
	if home != "" && strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

func inferHarnessState(title string) string {
	trimmed := strings.TrimSpace(title)
	switch {
	case strings.HasSuffix(trimmed, " ●"):
		return "working"
	case strings.HasSuffix(trimmed, " ?"):
		return "waiting-input"
	case strings.HasSuffix(trimmed, " !"):
		return "waiting-permission"
	case strings.HasSuffix(trimmed, " ✓"):
		return "done"
	default:
		return ""
	}
}

func titleHasHarnessMarker(title string) string {
	trimmed := strings.TrimSpace(title)
	if strings.HasPrefix(trimmed, "OC |") {
		return "opencode"
	}
	lower := strings.ToLower(trimmed)
	for _, name := range []string{"claude", "codex", "opencode"} {
		if lower == name || strings.HasPrefix(lower, name+" ") {
			return name
		}
	}
	return ""
}

func IsStaleHarness(p HarnessPane) bool {
	if marker := titleHasHarnessMarker(p.Title); marker != "" && marker != p.Harness {
		return true
	}
	switch p.Command {
	case "zsh", "bash", "fish", "sh", "dash", "ksh", "tcsh", "tmux", "screen", "":
		return titleHasHarnessMarker(p.Title) == ""
	default:
		return false
	}
}

func looksLikeVersion(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}

func inferHarness(title, command, startCommand, windowName string) (string, string) {
	for _, cmd := range []string{"claude", "codex", "opencode"} {
		if command == cmd || strings.Contains(startCommand, cmd) {
			return cmd, inferHarnessState(title)
		}
	}
	if looksLikeVersion(command) {
		return "claude", inferHarnessState(title)
	}
	lower := strings.ToLower(strings.TrimSpace(title))
	for _, name := range []string{"claude", "codex", "opencode"} {
		if lower == name || strings.HasPrefix(lower, name+" ") {
			return name, inferHarnessState(title)
		}
	}
	if strings.HasPrefix(title, "OC |") {
		return "opencode", inferHarnessState(title)
	}
	if windowName == "claude" {
		return "claude", inferHarnessState(title)
	}
	return "", ""
}

func ParseHarnessFields(fields []string) (HarnessPane, bool) {
	if len(fields) != 13 {
		return HarnessPane{}, false
	}
	p := HarnessPane{
		SessionID:    fields[0],
		SessionName:  fields[1],
		WindowID:     fields[2],
		WindowIndex:  fields[3],
		WindowName:   fields[4],
		PaneID:       fields[5],
		PaneIndex:    fields[6],
		Harness:      fields[7],
		State:        fields[8],
		Title:        fields[9],
		Command:      fields[10],
		StartCommand: fields[11],
		Path:         fields[12],
	}
	if p.PaneID == "" {
		return HarnessPane{}, false
	}
	if p.Harness == "" {
		harness, state := inferHarness(p.Title, p.Command, p.StartCommand, p.WindowName)
		if harness == "" {
			return HarnessPane{}, false
		}
		p.Harness = harness
		p.State = state
	} else {
		live := ""
		for _, cmd := range []string{"claude", "codex", "opencode"} {
			if p.Command == cmd {
				live = cmd
				break
			}
		}
		if live == "" && looksLikeVersion(p.Command) {
			live = "claude"
		}
		if live != "" && live != p.Harness {
			p.Harness = live
			p.State = inferHarnessState(p.Title)
		}
	}
	return p, true
}

func ListHarnessPanes() ([]HarnessPane, error) {
	format := strings.Join([]string{
		"#{session_id}", "#{session_name}", "#{window_id}", "#{window_index}",
		"#{window_name}", "#{pane_id}", "#{pane_index}", "#{@harness}",
		"#{@harness_state}", "#{pane_title}", "#{pane_current_command}",
		"#{pane_start_command}", "#{pane_current_path}",
		"#{?pane_id,HARNESS_RECORD_END,}",
	}, "\n")
	text, err := runTmuxOutput("list-panes", "-a", "-F", format)
	if err != nil {
		if isNoServerRunning(err) || isExitCode(err, 1) {
			return []HarnessPane{}, nil
		}
		return nil, err
	}
	var panes []HarnessPane
	lines := strings.Split(text, "\n")
	for i := 0; i+13 < len(lines); i += 14 {
		if lines[i+13] != "HARNESS_RECORD_END" {
			continue
		}
		if p, ok := ParseHarnessFields(lines[i : i+13]); ok {
			if !IsStaleHarness(p) {
				panes = append(panes, p)
			}
		}
	}
	return panes, nil
}

func FormatHarnessPane(p HarnessPane) string {
	clean := func(s string) string {
		s = strings.ReplaceAll(s, "\t", " ")
		s = strings.ReplaceAll(s, "\n", " ")
		return strings.TrimSpace(s)
	}
	title := clean(p.Title)
	if title == "" {
		title = clean(p.Path)
	}
	session := clean(ShortenHome(p.SessionName))
	return fmt.Sprintf("%s\t%s %s\t%s\t%s:%s\t%s",
		p.PaneID, HarnessSymbol(p.State), p.Harness, session,
		p.WindowIndex, p.PaneIndex, title)
}

func DisplayHarnessPanes(panes []HarnessPane) (HarnessPane, error) {
	lines := make([]string, 0, len(panes))
	index := make(map[string]HarnessPane, len(panes))
	for _, p := range panes {
		lines = append(lines, FormatHarnessPane(p))
		index[p.PaneID] = p
	}
	fzfCmd := `fzf --delimiter='\t' --with-nth=2,3,4,5 ` +
		`--header 'Select harness pane to jump to.' ` +
		`--preview "` + tmuxShellCmd() + ` capture-pane -ep -t {1}" --preview-window="right:60%" --height="100%"`
	buf, err := script.
		Echo(strings.Join(lines, "\n")).
		Exec(fzfCmd).
		WithStderr(os.Stdout).
		String()
	if err != nil {
		return HarnessPane{}, err
	}
	selected := strings.TrimSpace(buf)
	if selected == "" {
		return HarnessPane{}, fmt.Errorf("no pane selected")
	}
	paneID := selected
	if i := strings.Index(selected, "\t"); i >= 0 {
		paneID = selected[:i]
	}
	p, ok := index[paneID]
	if !ok {
		return HarnessPane{}, fmt.Errorf("unknown pane selected: %s", paneID)
	}
	return p, nil
}

func JumpToHarnessPane(p HarnessPane) error {
	inside := os.Getenv("TMUX") != ""
	if inside {
		if err := runTmux("switch-client", "-t", p.SessionID); err != nil {
			return err
		}
	}
	if err := runTmux("select-window", "-t", p.WindowID); err != nil {
		return err
	}
	if err := runTmux("select-pane", "-t", p.PaneID); err != nil {
		return err
	}
	if !inside {
		return Attach(p.SessionID)
	}
	return nil
}
