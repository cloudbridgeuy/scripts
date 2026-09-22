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

## Harness API

- `ListHarnessPanes() ([]HarnessPane, error)` — all panes running a harness (claude, codex, opencode), detected via `@harness` / `@harness_state` pane options (set by the shell hooks) with fallback inference from title, command, start command, and window name. A bare `major.minor.patch` foreground command means `claude` (it rewrites its process name to its version). A live foreground command naming a harness always overrides conflicting stored options (pane reused by another harness), resetting state to title-derived. Drops stale entries where the foreground is back to a shell with no harness title marker, or where the title marker names a different harness than the stored one.
- `IsStaleHarness(pane) bool` — true when the harness exited: foreground is an interactive shell or empty with no harness title marker, or the title marker names a different harness than the recorded one.
- `DisplayHarnessPanes(panes) (HarnessPane, error)` — fzf picker with pane-capture preview; returns the selected pane.
- `JumpToHarnessPane(pane) error` — jumps to the pane across sessions via stable IDs (`switch-client` + `select-window` + `select-pane`), avoiding session-name escaping issues.
- `FormatHarnessPane(pane) string` — tab-separated display line (`pane_id`, status, session, `window:pane`, title) for fzf with hidden ID column.
- `HarnessSymbol(state) string` — maps hook state (`working`, `waiting-input`, `waiting-permission`, `done`) to `● ? ! ✓`.

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
