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

## Report Command (`report.go`)

Executes bash commands from structured input and produces a report containing each command's description, exit code, and combined stdout/stderr output.

### Usage

```
scripts report [flags] [INPUT]
```

Input can come from three sources (checked in priority order):

1. **File** — `scripts report --file commands.txt`
2. **Piped stdin** — `cat commands.txt | scripts report`
3. **Inline args** — `scripts report 'echo hello' 'date'`

### Input Format

```
# List running containers
docker ps

# Show disk usage
df -h \
  --type=ext4

date +%Y-%m-%d
```

- Lines starting with `#` become the description for the next command
- Multiple consecutive `#` lines: last one wins
- Lines ending with `\` are joined with the next line (continuation)
- Blank lines are skipped
- Commands without a preceding `#` comment have an empty description

### Flags

- `--format, -f` — Output format: `xml` or `md` (default: `md`)
- `--file` — Read commands from a file instead of stdin or args
- `--on-error` — Error behavior: `continue` or `stop` (default: `continue`)

### Output Formats

**Markdown** (`--format md`) produces a structured document with `# Report` heading, numbered `## Command N` sections, description text, status code, command in a fenced block, and output in a fenced block.

**XML** (`--format xml`) produces `<report>` with `<action>` elements, each containing `<description>`, `<command>`, `<status>`, and `<output>` children. Uses `encoding/xml` for proper escaping.

### Architecture

**Functional core** — pure functions with no side effects:

- `ParseActions(text string) []Action` (`parser.go`) — splits input text into a slice of `Action{Description, Command}` structs following the comment/continuation/blank-line rules
- `FormatXML(results []Result) (string, error)` (`format.go`) — marshals results into indented XML via `encoding/xml`
- `FormatMarkdown(results []Result) string` (`format.go`) — builds a Markdown string with numbered command sections
- `FormatReport(results []Result, format Format) (string, error)` (`format.go`) — dispatches to the appropriate formatter
- `ParseFormat(s string) (Format, error)` and `ParseOnErrorBehavior(s string) (OnErrorBehavior, error)` (`types.go`) — validate flag strings into typed constants

**Imperative shell** — functions that perform I/O:

- `ResolveInput(stdin, filePath, args, isInputTTY) (string, error)` (`executor.go`) — reads input from file, stdin, or args in priority order
- `ExecuteActions(actions, onError, shell) []Result` (`executor.go`) — runs each command via `exec.Command(shell, "-c", cmd)` with `CombinedOutput()`; halts early when `onError` is `Stop` and a command exits non-zero
- Cobra command handler (`cmd/report.go`) — wires flags, resolves input, parses, executes, formats, and prints

**Shell selection** uses `$SHELL` environment variable with `/bin/sh` as fallback. Commands run with `-c` (command string) to avoid sourcing interactive shell profiles.
