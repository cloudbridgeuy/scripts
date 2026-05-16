# pkg/tmux

Tmux session and window management primitives. Wraps the `tmux` CLI directly via `os/exec` (with `bitfield/script` reserved for piped fzf flows).

## Session API

- `Switch(name string) error` — switches to `name`, creating the session if it doesn't exist. Inside tmux uses `switch-client`; outside it uses `attach`.
- `SwitchExisting(name string) error` — switches without creating; used by prev/next rotation through history.
- `NewSession(name string) error` — creates a detached session with `-c <name>` (working directory set to the session name).
- `KillSession(name string) error` — terminates the session.
- `HasSession(name string) error` — returns nil if the session exists, error otherwise.
- `SessionExists(name string) (bool, error)` — same check, surfaced as a boolean.
- `ListSessions() ([]string, error)` — lists session names, swallowing the "no server running" error as an empty list.
- `Attach(name string) error` / `SwitchClient(name string) error` — direct primitives behind `Switch`.
- `DisplaySessions() (string, error)` — fzf picker over sessions with pane-capture preview.
- `GetCurrentSession() (string, error)` — current session name (uses `$TMUX_PANE` / `display-message`).

## Window API

- `ListWindows() ([]string, error)`
- `NewWindow(name, command, directory string) error`
- `KillWindow(windowID string) error`
- `SelectWindow(name string) error`

## Session Name Canonicalisation

Sessions are named after directory paths. Dots in directory names conflict with tmux's target-pattern syntax (`session:window.pane`), so `canonicalSessionName()` replaces `.` with `_` before any tmux call. Callers pass real paths; the canonicalisation happens inside the package.

## Switch-then-Persist Pattern

History (`~/.scripts.yaml`) is updated **only** after a successful switch. If tmux returns an error, the file is left alone — preventing a broken session from polluting recent history.

## Error Handling

- `runTmux` / `runTmuxOutput` capture combined output and join non-empty stderr into the returned error with `fmt.Errorf("%w: %s", err, message)`.
- `isExitCode` / `isNoServerRunning` classify expected error shapes (no server running ⇒ empty session list).
