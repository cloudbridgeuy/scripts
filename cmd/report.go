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

	"github.com/cloudbridgeuy/scripts/pkg/errors"
	"github.com/cloudbridgeuy/scripts/pkg/report"
	"github.com/cloudbridgeuy/scripts/pkg/term"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report [flags] [INPUT]",
	Short: "Run commands and produce a structured report",
	Long: `Executes bash commands from input and produces a report with each
command's description, exit code, and output. Input can come from stdin,
a file, or inline arguments.

Comments (# lines) become command descriptions. Line continuations (\)
are supported. Output format is XML or Markdown.`,
	Run: func(cmd *cobra.Command, args []string) {
		formatStr, err := cmd.Flags().GetString("format")
		if err != nil {
			errors.HandleErrorWithReason(err, "Can't get the --format flag")
		}

		fileFlag, err := cmd.Flags().GetString("file")
		if err != nil {
			errors.HandleErrorWithReason(err, "Can't get the --file flag")
		}

		onErrorStr, err := cmd.Flags().GetString("on-error")
		if err != nil {
			errors.HandleErrorWithReason(err, "Can't get the --on-error flag")
		}

		format, err := report.ParseFormat(formatStr)
		if err != nil {
			errors.HandleError(err)
		}

		onError, err := report.ParseOnErrorBehavior(onErrorStr)
		if err != nil {
			errors.HandleError(err)
		}

		input, err := report.ResolveInput(os.Stdin, fileFlag, args, term.IsInputTTY())
		if err != nil {
			errors.HandleError(err)
		}

		actions := report.ParseActions(input)

		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}

		results := report.ExecuteActions(actions, onError, shell)

		output, err := report.FormatReport(results, format)
		if err != nil {
			errors.HandleError(err)
		}

		fmt.Println(output)
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.Flags().StringP("format", "f", "md", "Output format: xml or md")
	reportCmd.Flags().String("file", "", "Read commands from file")
	reportCmd.Flags().String("on-error", "continue", "Error behavior: continue or stop")
}
