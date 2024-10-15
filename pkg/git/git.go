package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bitfield/script"
)

// GetCurrentBranch returns the current branch name
func GetCurrentBranch() (string, error) {
	branch, err := script.Exec("git branch --show-current").String()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(branch), nil
}

// CreateSemanticCommit creates a semantic commit message based on the git diff output
func CreateSemanticCommit() (string, error) {
	var buf bytes.Buffer

	branch, err := GetCurrentBranch()
	if err != nil {
		return "", err
	}

	if _, err = script.
		Exec("git diff --staged -- . ':(exclude)package-lock.json' ':(exclude)lazy-lock.json' ':(exclude)*.lock'").
		Exec(fmt.Sprintf("llm-stream --template git-semantic-commit --vars '{ \"branch\": \"%s\" }' --preset sonnet", branch)).
		Tee(&buf).
		Stdout(); err != nil {
		return "", err
	}

	output := buf.String()

	re := regexp.MustCompile(`(?s)<output>(.*?)</output>`)
	match := re.FindStringSubmatch(output)
	if len(match) < 2 || match[1] == "" {
		return "", fmt.Errorf("can't find the <output></output> tag on the llm result: %s", output)
	}

	return strings.TrimSpace(match[1]), nil
}

// Commit creates a commit with the given message
func Commit(commit string) error {
	_, err := script.Echo(commit).Exec("git commit -F -").Stdout()
	return err
}

// AddAll stages all changes.
func AddAll() error {
	_, err := script.Exec("git add -A").Stdout()
	return err
}

// Add opens up an fzf window to select the files that need to be staged before the semantic commit is generated.
func Add() error {
	files, err := script.
		Exec("git ls-files --modified").
		Exec("fzf -m --preview 'git diff -- {}' --preview-window=right:80% --height=100% --border").
		WithStderr(os.Stdout).
		String()
	if err != nil {
		return err
	}
	files = strings.TrimSpace(files)
	if files == "" {
		return fmt.Errorf("No files selected")
	}

	_, err = script.Echo(files).Exec("xargs git add").Stdout()
	return err
}

// CommitAmend amends the last commit
func CommitAmend() error {
	command := exec.Command("git", "commit", "--amend")
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	return command.Run()
}
