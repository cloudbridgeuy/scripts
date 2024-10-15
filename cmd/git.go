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
	"github.com/spf13/cobra"

	"github.com/cloudbridgeuy/scripts/pkg/errors"
	"github.com/cloudbridgeuy/scripts/pkg/git"
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

		commit, err := git.CreateSemanticCommit()
		if err != nil {
			errors.HandleErrorWithReason(err, "can't create a semantic commit")
			return
		}

		if noCommit {
			return
		}

		if err := git.Commit(commit); err != nil {
			errors.HandleErrorWithReason(err, "can't create the git commit")
			return
		}

		if err := git.CommitAmend(); err != nil {
			errors.HandleErrorWithReason(err, "can't amend the commit")
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(gitCmd)

	gitCmd.AddCommand(semanticCmd)

	semanticCmd.Flags().Bool("no-commit", false, "Don't run the git commit command automatically")
}
