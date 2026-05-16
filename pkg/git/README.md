# pkg/git

Git operations focused on the semantic-commit workflow. Backed by `git` and the external `llm-stream` tool.

## API

- `GetCurrentBranch() (string, error)` — `git branch --show-current`, trimmed.
- `HasStagedFiles() (bool, error)` — true when `git diff --cached --name-only` is non-empty.
- `CreateSemanticCommit() (string, error)` — pipes the staged diff into `llm-stream` and extracts the message from the `<output>…</output>` tag. Fails fast when nothing is staged, with a hint pointing at `--add` / `--all`.
- `Commit(message string) error` — `git commit -F -` with the message on stdin.
- `AddAll() error` — `git add -A`.
- `Add() error` — interactive fzf picker over `git ls-files --modified --others` with a live `git diff --color` preview.
- `CommitAmend() error` — opens `git commit --amend` bound to the user's stdio so the editor pops up.

## Diff Filtering

The diff fed to `llm-stream` excludes lock files so they don't pollute the prompt:

```
':(exclude)package-lock.json' ':(exclude)lazy-lock.json' ':(exclude)*.lock'
```

## llm-stream Contract

The `git-semantic-commit` template returns the commit message wrapped in `<output></output>`. `CreateSemanticCommit` matches with `(?s)<output>(.*?)</output>`; missing or empty contents are a hard error that surfaces the raw LLM output so the user can see what went wrong.

## Semantic Commit Workflow

1. Stage files (or pass `--all` / `--add` to the command).
2. Build the filtered staged diff.
3. Call `llm-stream --template git-semantic-commit --vars '{"branch": "<branch>"}' --preset gpt`.
4. Extract the message from `<output></output>`.
5. `Commit` writes it.
6. `CommitAmend` opens the editor so the user can review and edit before the commit lands.
