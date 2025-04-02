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
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bitfield/script"
	"github.com/cloudbridgeuy/scripts/pkg/errors"
	"github.com/cloudbridgeuy/scripts/pkg/term"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:                "watch [COMMAND...]",
	Short:              "Watch a command",
	DisableFlagParsing: true,
	Long: `The list of accounts will be constructed based on all environment
variables found that begin with 'GITHUB_PAT'. Once selected, its
value will be used to authenticate the 'gh' cli.`,
	Run: func(cmd *cobra.Command, args []string) {
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

		options := make([]string, len(args))
		command := []string{}
		for i, arg := range args {
			if arg == "--" {
				command = args[i+1:]
				break
			}
			options[i] = arg
		}

		if len(command) == 0 {
			err := fmt.Errorf("no command provided")
			errors.HandleErrorWithReason(err, "Error")
			os.Exit(1)
		}

		if err := cmd.ParseFlags(options); err != nil {
			errors.HandleErrorWithReason(err, "Can't parse watch command arguments")
			os.Exit(1)
		}

		interval, err := cmd.Flags().GetInt("interval")
		if err != nil {
			errors.HandleErrorWithReason(err, "Can't get the --interval flag")
			os.Exit(1)
		}

		term.Clear()

		go func() {
			var err error
			var curr, next string
			curr, err = script.Exec(strings.Join(command, " ")).String()
			if err != nil {
				errors.HandleErrorWithReason(err, "Can't execute command")
				os.Exit(1)
			}

			fmt.Printf(curr)

			yellow := color.New(color.FgYellow).SprintFunc()
			green := color.New(color.FgGreen).SprintFunc()

			for {
				// Wait for `interval` amount of seconds.
				time.Sleep(time.Duration(interval) * time.Second)

				next, err = script.Exec(strings.Join(command, " ")).String()
				if err != nil {
					errors.HandleErrorWithReason(err, "Can't execute command")
					os.Exit(1)
				}

				term.CenterCursor()

				// Iterate over each character of `next` and compare it to prev.
				// If they are different, print the character in yellow.
				// If they are the same, print the character in white.
				nextRunes := []rune(next)
				currRunes := []rune(curr)
				for i := 0; i < len(nextRunes); i++ {
					if i >= len(currRunes) {
						fmt.Print(green(string(nextRunes[i])))
					} else if currRunes[i] != nextRunes[i] {
						fmt.Print(yellow(string(nextRunes[i])))
					} else {
						fmt.Print(string(nextRunes[i]))
					}
				}
				fmt.Println()
				curr = next
			}
		}()

		sig := <-signalChan

		fmt.Printf("Received %s, exiting...\n", sig)
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().IntP("interval", "i", 1, "Interval in seconds to run the command")
}
