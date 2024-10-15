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
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitfield/script"
	"github.com/cloudbridgeuy/scripts/pkg/errors"
)

// ghCmd represents the case command
var ghCmd = &cobra.Command{
	Use:   "gh COMMAND [OPTIONS]",
	Short: "Wrapper around the `gh` cli",
	Long:  "This commands aim to simplify common actions or alias more complex commands.",
}

var re = regexp.MustCompile("=.*")

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate a different GitHub account.",
	Long: `The list of accounts will be constructed based on all environment
variables found that begin with 'GITHUB_PAT'. Once selected, its
value will be used to authenticate the 'gh' cli.`,
	Run: func(cmd *cobra.Command, args []string) {
		env_key, err := script.Exec("env").Match("GITHUB_PAT").ReplaceRegexp(re, "").Exec("fzf").WithStderr(os.Stdout).String()
		if err != nil {
			errors.HandleErrorWithReason(err, "can't get profile from environment variables")
			return
		}

		env_key = strings.TrimSpace(env_key)
		fmt.Println("profile =", env_key)

		env_value := os.Getenv(env_key)
		if env_value == "" {
			errors.HandleErrorWithReason(fmt.Errorf("can't find value for key"), fmt.Sprintf("there's an issue when reading the %s environment variable", env_key))
			return
		}

		_, err = script.Echo(env_value).Exec("gh auth login --with-token").Stdout()
		if err != nil {
			errors.HandleErrorWithReason(err, "can't echo the value")
		}
	},
}

func init() {
	rootCmd.AddCommand(ghCmd)

	ghCmd.AddCommand(authCmd)
}
