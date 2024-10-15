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

	"github.com/spf13/cobra"

	"github.com/cloudbridgeuy/scripts/pkg/errors"
	"github.com/cloudbridgeuy/scripts/pkg/utils"
	"github.com/iancoleman/strcase"
)

// caseCmd represents the case command
var caseCmd = &cobra.Command{
	Use:   "case COMMAND [STRING]",
	Short: "Change the casing of a string to conform with typical norms.",
	Long:  "With no [STRING], or when [STRING] is -, it reads from `stdin`.",
}

var toSnakeCmd = &cobra.Command{
	Use:   "to-snake [STRING]",
	Short: "Changes a string to snake-case.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s, err := utils.FirstOrStdin(args)
		if err != nil {
			errors.HandleErrorWithReason(err, "can't read input from stdin")
		}

		fmt.Print(strcase.ToSnake(s))
	},
}

var toCamelCmd = &cobra.Command{
	Use:   "to-camel [STRING]",
	Short: "Changes a string to camelCase.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s, err := utils.FirstOrStdin(args)
		if err != nil {
			errors.HandleErrorWithReason(err, "can't read input from stdin")
		}

		fmt.Print(strcase.ToLowerCamel(s))
	},
}

var toScreamingSnakeCmd = &cobra.Command{
	Use:   "to-screaming-snake [STRING]",
	Short: "Changes a string to SCREAMING_SNAKE.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s, err := utils.FirstOrStdin(args)
		if err != nil {
			errors.HandleErrorWithReason(err, "can't read input from stdin")
		}

		fmt.Print(strcase.ToScreamingSnake(s))
	},
}

var toKebabCmd = &cobra.Command{
	Use:   "to-kebab [STRING]",
	Short: "Changes a string to kebab-case.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s, err := utils.FirstOrStdin(args)
		if err != nil {
			errors.HandleErrorWithReason(err, "can't read input from stdin")
		}

		fmt.Print(strcase.ToKebab(s))
	},
}

var toScreamingCamel = &cobra.Command{
	Use:   "to-screaming-camel [STRING]",
	Short: "Changes a string to ScreamingCamel.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s, err := utils.FirstOrStdin(args)
		if err != nil {
			errors.HandleErrorWithReason(err, "can't read input from stdin")
		}

		fmt.Print(strcase.ToCamel(s))
	},
}

func init() {
	rootCmd.AddCommand(caseCmd)

	caseCmd.AddCommand(toCamelCmd)
	caseCmd.AddCommand(toKebabCmd)
	caseCmd.AddCommand(toScreamingCamel)
	caseCmd.AddCommand(toScreamingSnakeCmd)
	caseCmd.AddCommand(toSnakeCmd)
}
