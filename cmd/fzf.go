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
	"strings"

	"github.com/bitfield/script"
	"github.com/cloudbridgeuy/scripts/pkg/errors"
	"github.com/spf13/cobra"
)

// fzfCmd represents the tmux command
var fzfCmd = &cobra.Command{
	Use:   "fzf COMMAND [OPTIONS]",
	Short: "Wrapper around the `fzf` cli.",
	Long:  "This commands aim to simplify common actions or alias more complex commands when using fzf.",
}

var rgCmd = &cobra.Command{
	Use:   "rg [OPTIONS] [INITIAL_QUERY]",
	Short: "Use fzf as a selector interface for RipGrep",
	Long:  `Every time you type, the process will restart with the updated query.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initialQuery := ""
		if len(args) == 1 {
			initialQuery = args[0]
		}

		rgOptions, err := cmd.Flags().GetString("ripgrep-options")
		if err != nil {
			errors.HandleErrorWithReason(err, "can't get the selected session")
		}

		line, err := script.
			Exec(fmt.Sprintf(`fzf \
        --bind "change:reload:rg %s {q} | column -s: -t || true" \
        --ansi --disabled --query "%s" \
        --height=50%% --layout=reverse`, rgOptions, initialQuery)).
			WithEnv(append([]string{
				"RG_PREFIX=\"rg --column --line-number --no-heading --color=always --smart-case\"",
				fmt.Sprintf("FZF_DEFAULT_COMMAND=\"rg %s '%s' | column -s: -t\"", rgOptions, initialQuery),
			}, os.Environ()...)).
			String()
		if err != nil {
			errors.HandleErrorWithReason(err, "Can't run RipGrep with FZF")
			return
		}

		line = strings.TrimSpace(line)

		editor := os.Getenv("EDITOR")
		if editor == "" {
			fmt.Println(line)
			return
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			fmt.Println("Invalid input format")
			return
		}

		file := parts[0]
		lineNumber := parts[1]
		columnNumber := parts[2]

		command := exec.Command(editor, fmt.Sprintf("+normal %sG%s|zv", lineNumber, columnNumber), file)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.Run()
	},
}

func init() {
	rootCmd.AddCommand(fzfCmd)

	fzfCmd.AddCommand(rgCmd)

	rgCmd.Flags().String("ripgrep-options", "--column --line-number --no-heading --color=always --smart-case", "RipGrep command options")
}
