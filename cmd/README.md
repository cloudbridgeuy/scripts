# cmd

Command definitions following Cobra's command pattern. Each file defines a single command or subcommand tree.

## Watch Command (`watch.go`)

Repeatedly executes a command at a configurable interval and highlights character-level differences between consecutive runs.

### Usage

```
scripts watch [flags] -- <COMMAND...>
```

### Flags

- `--interval, -i` — Interval in seconds (default: 1)
- `--columns, -c` — Number of columns (default: terminal width)

### Architecture

**Argument parsing** is handled by `parseWatchArgs(args []string) (flags []string, command []string)`, a pure function that splits at the first `--`. Uses `append` to avoid trailing empty strings from pre-allocated slices.

**Command execution** uses `os/exec.Command` with `CombinedOutput()` (not `bitfield/script`). Environment variables include `COLUMNS` to pass terminal width to child processes.

**Error handling:**
- Initial execution failure is fatal (exits before starting the loop)
- Loop execution errors display the command's combined output (or the error message as fallback) on screen, then continue to the next tick

**Terminal control** uses ANSI escape sequences via `pkg/term/commands.go`:
- `Clear()` — `\033[2J\033[H` (clear screen + home cursor)
- `CenterCursor()` — `\033[H` (home cursor)
- `ClearFromCursor()` — `\033[J` (erase from cursor to end of screen, prevents stale content)

**Diff highlighting** compares rune-by-rune between current and previous output:
- Yellow: changed characters
- Green: new characters (output grew longer)
- Default: unchanged characters
