/*
Copyright © 2024 Guzmán Monné guzman.monne@cloudbridge.com.uy

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitfield/script"
	"github.com/cloudbridgeuy/scripts/pkg/errors"
)

// gitCmd represents the case command
var gitCmd = &cobra.Command{
	Use:   "git COMMAND [OPTIONS]",
	Short: "Wrapper around the `git` cli",
	Long:  "This commands aim to simplify common actions or alias more complex commands.",
}

var semanticCmd = &cobra.Command{
	Use:   "semantic [OPTIONS]",
	Short: "Create a semantic git commit from the git diff output",
	Run: func(cmd *cobra.Command, args []string) {
		noCommit, err := cmd.Flags().GetBool("no-commit")
		if err != nil {
			errors.HandleErrorWithReason(err, "can't get the --no-commit flag")
			return
		}

		branch, err := script.Exec("git branch --show-current").String()
		if err != nil {
			errors.HandleErrorWithReason(err, "can't get the current branch")
			return
		}
		branch = strings.TrimSpace(branch)

		// var buf bytes.Buffer
		//
		// script.
		// 	Exec("git diff --staged -- . ':(exclude)package-lock.json' ':(exclude)lazy-lock.json' ':(exclude)*.lock'").
		// 	Exec(fmt.Sprintf("llm-stream --template git-semantic-commit --vars '{ \"branch\": \"%s\" }' --preset sonnet", branch)).
		// 	Tee(&buf).
		// 	Stdout()

		if noCommit {
			return
		}

		buf := `<context>
The changes made in this diff involve moving the regular expression declaration inside the 'Run' function of the 'authCmd' command. The regular expression 'var re = regexp.MustCompile("=.*")' was previously declared at the package level and has now been moved inside the function scope.
</context>

<thinking>
Given the changes made, we can observe that:

1. The modification is relatively small and doesn't introduce new functionality or fix a bug.
2. The change is related to code organization and scope, moving a variable declaration from package level to function level.
3. This type of change is best categorized as a refactor, as it improves code structure without changing its external behavior.
4. The affected code is part of a command-line interface (CLI) tool, likely related to GitHub authentication, as evidenced by the 'gh' and 'auth' commands.

Based on these observations, the most appropriate semantic commit type would be 'refactor'. The main service affected appears to be a GitHub-related CLI tool, which we can refer to as 'gh-cli' for the purpose of this commit message.
</thinking>

<output>
refactor(gh-cli): move regex declaration to function scope
</output>`

		re := regexp.MustCompile(`(?s)<output>(.*?)</output>`)
		// match := re.FindStringSubmatch(buf.String())
		match := re.FindStringSubmatch(buf)
		fmt.Println(match)
		if len(match) < 2 || match[1] == "" {
			errors.HandleErrorWithReason(fmt.Errorf("can't find the output"), "there's an issue when parsing the output")
			return
		}
		output := strings.TrimSpace(match[1])

		_, err = script.Echo(output).Exec("git commit -F -").Stdout()
		if err != nil {
			errors.HandleErrorWithReason(err, "can't commit output")
			return
		}

		command := exec.Command("git", "commit", "--amend")
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.Stdin = os.Stdin

		if err := command.Run(); err != nil {
			errors.HandleErrorWithReason(err, "can't amend commit")
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(gitCmd)

	gitCmd.AddCommand(semanticCmd)

	semanticCmd.Flags().Bool("no-commit", false, "Don't run the git commit command automatically")
}
