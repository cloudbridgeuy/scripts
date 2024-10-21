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
	"strings"
	"sync"

	"github.com/bitfield/script"
	"github.com/cloudbridgeuy/scripts/pkg/errors"
	"github.com/cloudbridgeuy/scripts/pkg/logger"
	"github.com/cloudbridgeuy/scripts/pkg/tmux"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// tmuxCmd represents the tmux command
var tmuxCmd = &cobra.Command{
	Use:   "tmux COMMAND [OPTIONS]",
	Short: "Wrapper around the `tmux` cli.",
	Long:  "This commands aim to simplify common actions or alias more complex commands.",
}

type directory struct {
	path     string
	mindepth int
	maxdepth int
	grep     string
}

var newCmd = &cobra.Command{
	Use:   "new [OPTIONS]",
	Short: "Create a new session from a subset of all the system directories.",
	Long: `The goal of this command is to create a new Tmux session named after the
working directory assigned to it. This way, we can ensure not to duplicate
session names when creating them, and can use them to jump between
projects, keeping all the required configuration namespaced inside.`,
	Run: func(cmd *cobra.Command, args []string) {
		home := os.Getenv("HOME")
		if home == "" {
			home = "/"
		}
		directories := []directory{
			{
				path:     home,
				mindepth: 1,
				maxdepth: 1,
				grep:     "*",
			},
			{
				path:     home + "/Projects",
				mindepth: 3,
				maxdepth: 4,
				grep:     ".*/Projects/[^/]*/[^/]*/branches/",
			},
			{
				path:     home + "/Projects",
				mindepth: 2,
				maxdepth: 4,
				grep:     ".*/Projects/[^/]*/[^/]*$",
			},
			{
				path:     home + "/Projects",
				mindepth: 1,
				maxdepth: 1,
				grep:     "*",
			},
		}

		var (
			wg        sync.WaitGroup
			mu        sync.Mutex
			allOutput string
			errChan   = make(chan error, len(directories))
		)

		for _, dir := range directories {
			wg.Add(1)
			go func(dir directory) {
				defer wg.Done()

				logger.Debugf("find %s -mindepth %d -maxdepth %d -name '%s'", dir.path, dir.mindepth, dir.maxdepth, dir.grep)
				output, err := script.Exec(fmt.Sprintf("find %s -mindepth %d -maxdepth %d -name '%s'", dir.path, dir.mindepth, dir.maxdepth, dir.grep)).String()
				if err != nil {
					errChan <- err
					return
				}

				mu.Lock()
				allOutput = allOutput + "\n" + strings.TrimSpace(output)
				mu.Unlock()
			}(dir)
		}

		go func() {
			wg.Wait()
			close(errChan)
		}()

		for err := range errChan {
			if err != nil {
				errors.HandleErrorWithReason(err, "can't evaluate the list of possible directories")
				return
			}
		}

		script.Echo(strings.TrimSpace(allOutput)).Exec("sort -u").Exec("fzf").WithStderr(os.Stdout).Stdout()
	},
}

var displayCmd = &cobra.Command{
	Use:   "display [OPTIONS]",
	Short: "Display all the running tmux sessions",
	Long:  `You can use this command to traverse to a different session.`,
	Run: func(cmd *cobra.Command, args []string) {
		noSwitch, err := cmd.Flags().GetBool("no-switch")
		if err != nil {
			errors.HandleErrorWithReason(err, "can't get the --no-switch flag")
			return
		}

		session, err := tmux.DisplaySessions()
		if err != nil {
			errors.HandleErrorWithReason(err, "can't display tmux sessions")
			return
		}

		if noSwitch {
			return
		}

		err = tmux.Switch(session)
		if err != nil {
			errors.HandleErrorWithReason(err, fmt.Sprintf("can't switch to session %s", session))
			return
		}
	},
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "See the session history.",
	Long: `Every movement through tmux session done through this tool will be
stored inside the 'tmux.sessions.history' list. This command only
lists the lines stored on that field.

NOTE: The session history usually doesn't match with Tmux.`,
	Run: func(cmd *cobra.Command, args []string) {
		for i, s := range config.Tmux.Sessions.History {
			fmt.Printf("%d\t\t%s\n", i, s)
		}
	},
}

var prevCmd = &cobra.Command{
	Use:   "prev",
	Short: "Go to the previous open session.",
	Long: `We keep a list of all the visited sessions in order of usage. You can
use this command plus the 'next' command to move between them.`,
	Run: func(cmd *cobra.Command, args []string) {
		historyLength := len(config.Tmux.Sessions.History)

		if historyLength == 0 {
			errors.HandleError(fmt.Errorf("No sessions found in history"))
			return
		}

		newHistory := make([]string, historyLength)

		for i, s := range config.Tmux.Sessions.History {
			if i == historyLength-1 {
				newHistory[0] = s
			} else {
				newHistory[i+1] = s
			}
		}

		session := newHistory[historyLength-1]

		config.Tmux.Sessions.History = newHistory

		home, err := os.UserHomeDir()
		if err != nil {
			errors.HandleErrorWithReason(err, "can't get the user home directory")
		}

		c, err := yaml.Marshal(config)
		if err != nil {
			errors.HandleErrorWithReason(err, "can't marshal the config file")
		}

		err = os.WriteFile(home+"/.scripts.yaml", c, 0644)
		if err != nil {
			errors.HandleErrorWithReason(err, "can't write back the config file")
		}

		err = tmux.Switch(session)
		if err != nil {
			errors.HandleErrorWithReason(err, fmt.Sprintf("can't switch to session %s", session))
			return
		}
	},
}

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Go to the next open session.",
	Long: `We keep a list of all the visited sessions in order of usage. You can
use this command plus the 'prev' command to move between them.`,
	Run: func(cmd *cobra.Command, args []string) {
		historyLength := len(config.Tmux.Sessions.History)

		if historyLength == 0 {
			errors.HandleError(fmt.Errorf("No sessions found in history"))
			return
		}

		newHistory := make([]string, historyLength)

		for i, s := range config.Tmux.Sessions.History {
			if i == 0 {
				newHistory[historyLength-1] = s
			} else {
				newHistory[i-1] = s
			}
		}

		session := newHistory[historyLength-1]

		config.Tmux.Sessions.History = newHistory

		home, err := os.UserHomeDir()
		if err != nil {
			errors.HandleErrorWithReason(err, "can't get the user home directory")
		}

		c, err := yaml.Marshal(config)
		if err != nil {
			errors.HandleErrorWithReason(err, "can't marshal the config file")
		}

		err = os.WriteFile(home+"/.scripts.yaml", c, 0644)
		if err != nil {
			errors.HandleErrorWithReason(err, "can't write back the config file")
		}

		err = tmux.Switch(session)
		if err != nil {
			errors.HandleErrorWithReason(err, fmt.Sprintf("can't switch to session %s", session))
			return
		}
	},
}

var goCmd = &cobra.Command{
	Use:   "go [SESSION]",
	Short: "Go to the provided session or pick one from those available.",
	Long: `You can either provide a full path to open a new session or leave the
SESSION argument empty to display the list of running sessions to pick one.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var session string
		var err error

		if len(args) > 0 {
			session = args[0]
			err = tmux.Switch(session)
			if err != nil {
				errors.HandleErrorWithReason(err, fmt.Sprintf("can't switch to session %s", session))
				return
			}
		} else {
			session, err = tmux.DisplaySessions()
			if err != nil {
				errors.HandleErrorWithReason(err, "can't display tmux sessions")
				return
			}
		}

		logger.Debugf("Updating config file with session: %s", session)
		var sessions []string

		for _, s := range config.Tmux.Sessions.History {
			if s == session || s == "" {
				continue
			}
			sessions = append(sessions, s)
		}

		sessions = append(sessions, session)

		config.Tmux.Sessions.History = sessions

		home, err := os.UserHomeDir()
		if err != nil {
			errors.HandleErrorWithReason(err, "can't get the user home directory")
		}

		c, err := yaml.Marshal(config)
		if err != nil {
			errors.HandleErrorWithReason(err, "can't marshal the config file")
		}

		err = os.WriteFile(home+"/.scripts.yaml", c, 0644)
		if err != nil {
			errors.HandleErrorWithReason(err, "can't write back the config file")
		}

		err = tmux.Switch(session)
		if err != nil {
			errors.HandleErrorWithReason(err, fmt.Sprintf("can't switch to session %s", session))
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(tmuxCmd)

	tmuxCmd.AddCommand(displayCmd)
	tmuxCmd.AddCommand(newCmd)
	tmuxCmd.AddCommand(goCmd)
	tmuxCmd.AddCommand(nextCmd)
	tmuxCmd.AddCommand(prevCmd)
	tmuxCmd.AddCommand(historyCmd)

	displayCmd.Flags().Bool("no-switch", false, "Don't run the git commit command automatically")
}
