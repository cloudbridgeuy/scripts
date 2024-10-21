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

	"github.com/spf13/cobra"

	"github.com/bitfield/script"
	"github.com/cloudbridgeuy/scripts/pkg/errors"
)

// sshCmd represents the git command
var sshCmd = &cobra.Command{
	Use:   "ssh COMMAND [OPTIONS]",
	Short: "Tools related to ssh commands.",
	Long:  "Useful commands to use in conjunction with ssh.",
}

var nvimCmd = &cobra.Command{
	Use:   "nvim [OPTIONS] IP_OR_URL",
	Short: "Create a semantic git commit from the `git diff` output of the staged files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key, err := cmd.Flags().GetString("key")
		if err != nil {
			errors.HandleErrorWithReason(err, "can't get the --key flag")
			return
		}

		user, err := cmd.Flags().GetString("user")
		if err != nil {
			errors.HandleErrorWithReason(err, "can't get the --user flag")
			return
		}

		config, err := cmd.Flags().GetString("config")
		if err != nil {
			errors.HandleErrorWithReason(err, "can't get the --config flag")
			return
		}

		home, err := os.UserHomeDir()
		if err != nil {
			errors.HandleErrorWithReason(err, "can't get the user's home directory")
			return
		}

		script.Exec(fmt.Sprintf("ssh -r %s %s/%s/*.vim %s@%s:/home/%s/.config/nvim", key, home, config, user, args[0], user))
		script.Exec(fmt.Sprintf("ssh -r %s %s/%s/*.lua %s@%s:/home/%s/.config/nvim", key, home, config, user, args[0], user))
		script.Exec(fmt.Sprintf("ssh -r %s %s/%s/lua/ %s@%s:/home/%s/.config/nvim/lua", key, home, config, user, args[0], user))
	},
}

func init() {
	rootCmd.AddCommand(sshCmd)

	sshCmd.AddCommand(nvimCmd)

	nvimCmd.Flags().StringP("key", "k", "", "The SSH key to use to connect to the server.")
	nvimCmd.Flags().StringP("user", "u", "ec2-user", "The user to use to connect to the server.")
	nvimCmd.Flags().StringP("config", "c", ".config/nvim", "Nvim's configuration directory.")
}
