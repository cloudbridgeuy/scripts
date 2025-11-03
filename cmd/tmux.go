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
	"sync"

	"github.com/bitfield/script"
	"github.com/cloudbridgeuy/scripts/pkg/errors"
	"github.com/cloudbridgeuy/scripts/pkg/logger"
	"github.com/cloudbridgeuy/scripts/pkg/tmux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
				grep:     ".*",
			},
			{
				path:     home + "/Projects",
				mindepth: 3,
				maxdepth: 4,
				grep:     ".*/Projects/[^/]*/[^/]*/branches/[^/]*",
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
				grep:     ".*",
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

				logger.Debugf("/usr/bin/find %s -mindepth %d -maxdepth %d -type d", dir.path, dir.mindepth, dir.maxdepth)
				regex := regexp.MustCompile(dir.grep)
				output, err := script.
					Exec(fmt.Sprintf("/usr/bin/find %s -mindepth %d -maxdepth %d -type d", dir.path, dir.mindepth, dir.maxdepth)).
					MatchRegexp(regex).
					String()
				if err != nil {
					logger.Errorf("find %s -mindepth %d -maxdepth %d -type d", dir.path, dir.mindepth, dir.maxdepth)
					logger.Errorf(err.Error())
					errChan <- err
					return
				}

				mu.Lock()
				allOutput = allOutput + "\n" + strings.TrimSpace(output)
				// allOutput = allOutput + "\n" + strings.Join(dirs, "\n")
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

		session, err := script.
			Echo(strings.TrimSpace(allOutput)).
			Exec("sort -ur").
			Exec(`fzf \
        --header 'Select the directory where you want your session to be created.' \
        --preview "eza -lha --icons --group-directories-first --git --no-user --color=always {}" \
        --preview-window="right:40%" \
        --height="100%"`).
			WithStderr(os.Stdout).
			String()
		if err != nil {
			errors.HandleErrorWithReason(err, "can't get the selected session")
			return
		}

		session = strings.TrimSpace(session)

		addToTmuxHistory(session)

		if err := saveConfig(); err != nil {
			errors.HandleErrorWithReason(err, "can't save the config file")
			return
		}

		if err = tmux.Switch(session); err != nil {
			errors.HandleErrorWithReason(err, fmt.Sprintf("can't switch to session %s", session))
			return
		}
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

		if err = tmux.Switch(session); err != nil {
			errors.HandleErrorWithReason(err, fmt.Sprintf("can't switch to session %s", session))
			return
		}
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync session to and from tmux.",
	Long: `Use this command whenever the sessions running in 'tmux' differs from
the one tracked by the tool's history. You can choose to sync in any
direction through the '--reverse' option. The default is to sync from
tmux to this tool, but if you include the '--reverse' option, then
sessions will be opened and closed from 'tmux' until both lists match.`,
	Run: func(cmd *cobra.Command, args []string) {
		reverse, err := cmd.Flags().GetBool("reverse")
		if err != nil {
			errors.HandleErrorWithReason(err, "can't get the --reverse flag")
		}

		sessions, err := tmux.Ls()
		if err != nil {
			errors.HandleErrorWithReason(err, "can't list tmux sessions")
		}

		inHistory := make(map[string]bool)
		inTmux := make(map[string]bool)

		history := getTmuxHistory()
		for _, session := range history {
			inHistory[session] = true
		}

		for _, session := range sessions {
			inTmux[session] = true
		}

		if reverse {
			for _, session := range append(sessions, history...) {
				if inHistory[session] {
					if inTmux[session] {
						continue
					} else {
						if err := tmux.NewSession(session); err != nil {
							errors.HandleErrorWithReason(err, fmt.Sprintf("can't create session %s", session))
							return
						}
					}
				} else {
					if err := tmux.KillSession(session); err != nil {
						errors.HandleErrorWithReason(err, fmt.Sprintf("can't kill session %s", session))
						return
					}
				}
			}
		} else {
			setTmuxHistory(sessions)

			if err := saveConfig(); err != nil {
				errors.HandleErrorWithReason(err, "can't save the config file")
				return
			}
		}

		for _, session := range sessions {
			inTmux[session] = true
		}

		if reverse {
			for _, session := range append(sessions, history...) {
				if inHistory[session] {
					if inTmux[session] {
						continue
					} else {
						if err := tmux.NewSession(session); err != nil {
							errors.HandleErrorWithReason(err, fmt.Sprintf("can't create session %s", session))
						}
					}
				} else {
					if err := tmux.KillSession(session); err != nil {
						errors.HandleErrorWithReason(err, fmt.Sprintf("can't kill session %s", session))
					}
				}
			}
		} else {
			setTmuxHistory(sessions)

			if err := saveConfig(); err != nil {
				errors.HandleErrorWithReason(err, "can't save the config file")
			}
		}

		session := sessions[len(sessions)-1]
		if err = tmux.Switch(session); err != nil {
			errors.HandleErrorWithReason(err, fmt.Sprintf("can't switch to session %s", session))
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "See list of active sessions.",
	Long: `Every movement through tmux session done through this tool will be
stored inside the 'tmux.sessions.history' list. This command only
lists the lines stored on that field.

NOTE: The session history usually doesn't match with Tmux.`,
	Run: func(cmd *cobra.Command, args []string) {
		history := getTmuxHistory()
		sessionsLength := len(history)

		var maxWidth int

		for _, s := range history {
			if len(s) > maxWidth {
				maxWidth = len(s)
			}
		}

		format := fmt.Sprintf("%%-3d %%-%ds\t", maxWidth)

		for i, s := range history {
			fmt.Printf(format, i+1, s)

			if i == sessionsLength-1 {
				fmt.Printf("*\n")
			} else {
				fmt.Printf("\n")
			}
		}
	},
}

var prevCmd = &cobra.Command{
	Use:   "prev",
	Short: "Go to the previous open session.",
	Long: `We keep a list of all the visited sessions in order of usage. You can
use this command plus the 'next' command to move between them.`,
	Run: func(cmd *cobra.Command, args []string) {
		history := getTmuxHistory()
		sessionsLength := len(history)

		if sessionsLength == 0 {
			errors.HandleError(fmt.Errorf("No sessions found in history"))
			return
		}

		sessions := make([]string, sessionsLength)

		for i, s := range history {
			if i == sessionsLength-1 {
				sessions[0] = s
			} else {
				sessions[i+1] = s
			}
		}

		session := sessions[sessionsLength-1]

		setTmuxHistory(sessions)

		if err := saveConfig(); err != nil {
			errors.HandleErrorWithReason(err, "can't save the config file")
			return
		}

		if err := tmux.Switch(session); err != nil {
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
		history := getTmuxHistory()
		sessionsLength := len(history)

		if sessionsLength == 0 {
			errors.HandleError(fmt.Errorf("No sessions found in history"))
			return
		}

		sessions := make([]string, sessionsLength)

		for i, s := range history {
			if i == 0 {
				sessions[sessionsLength-1] = s
			} else {
				sessions[i-1] = s
			}
		}

		session := sessions[sessionsLength-1]

		setTmuxHistory(sessions)

		if err := saveConfig(); err != nil {
			errors.HandleErrorWithReason(err, "can't save the config file")
			return
		}

		if err := tmux.Switch(session); err != nil {
			errors.HandleErrorWithReason(err, fmt.Sprintf("can't switch to session %s", session))
			return
		}
	},
}

var addCmd = &cobra.Command{
	Use:   "add SESSION",
	Short: "Add a new session.",
	Long:  "Creates a new tmux sessions and transitions to it.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		session := args[0]

		addToTmuxHistory(session)

		if err := saveConfig(); err != nil {
			errors.HandleErrorWithReason(err, "can't save the config file")
			return
		}

		err := tmux.Switch(session)
		if err != nil {
			errors.HandleErrorWithReason(err, fmt.Sprintf("can't switch to session %s", session))
			return
		}
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove SESSION",
	Short: "Removes an existing session.",
	Long:  "Creates a new tmux sessions and transitions to it.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		session := args[0]

		if err := tmux.KillSession(session); err != nil {
			errors.HandleErrorWithReason(err, fmt.Sprintf("can't kill session %s", session))
			return
		}

		removeFromTmuxHistory(session)

		if err := saveConfig(); err != nil {
			errors.HandleErrorWithReason(err, "can't save the config file")
			return
		}
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Output the current tmux configuration to stdout.",
	Long:  "Displays the current tmux configuration including session history for debugging purposes.",
	Run: func(cmd *cobra.Command, args []string) {
		settings := viper.AllSettings()
		c, err := yaml.Marshal(settings)
		if err != nil {
			errors.HandleErrorWithReason(err, "can't marshal the config")
			return
		}
		fmt.Print(string(c))
	}}

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
		} else {
			session, err = tmux.DisplaySessions()
			if err != nil {
				errors.HandleErrorWithReason(err, "can't display tmux sessions")
				return
			}
		}

		logger.Debugf("Updating config file with session: %s", session)

		addToTmuxHistory(session)

		if err := saveConfig(); err != nil {
			errors.HandleErrorWithReason(err, "can't save the config file")
			return
		}

		if err := tmux.Switch(session); err != nil {
			errors.HandleErrorWithReason(err, fmt.Sprintf("can't switch to session %s", session))
			return
		}
	},
}

var claudeCmd = &cobra.Command{
	Use:   "claude",
	Short: "Configure current session with claude, nvim, and zsh windows.",
	Long: `Creates three windows named 'claude', 'nvim', and 'zsh' running their
respective tools, then removes all other windows. Windows are created in order
and the zsh window is selected at the end.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			errors.HandleErrorWithReason(err, "can't get current working directory")
			return
		}

		// Get list of existing windows before creating new ones
		existingWindows, err := tmux.ListWindows()
		if err != nil {
			errors.HandleErrorWithReason(err, "can't list existing windows")
			return
		}

		// Create the three new windows in order
		windows := []struct {
			name    string
			command string
		}{
			{"claude", "zsh -i -c claude"},
			{"nvim", "zsh -i -c nvim"},
			{"zsh", "zsh"},
		}

		for _, w := range windows {
			if err := tmux.NewWindow(w.name, w.command, cwd); err != nil {
				errors.HandleErrorWithReason(err, fmt.Sprintf("can't create window %s", w.name))
				return
			}
		}

		// Remove all existing windows
		for _, windowID := range existingWindows {
			if err := tmux.KillWindow(windowID); err != nil {
				errors.HandleErrorWithReason(err, fmt.Sprintf("can't kill window %s", windowID))
				// Continue even if we fail to kill a window
			}
		}

		// Select the zsh window
		if err := tmux.SelectWindow("zsh"); err != nil {
			errors.HandleErrorWithReason(err, "can't select zsh window")
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
	tmuxCmd.AddCommand(listCmd)
	tmuxCmd.AddCommand(addCmd)
	tmuxCmd.AddCommand(removeCmd)
	tmuxCmd.AddCommand(syncCmd)
	tmuxCmd.AddCommand(configCmd)
	tmuxCmd.AddCommand(claudeCmd)

	displayCmd.Flags().Bool("no-switch", false, "Don't run the git commit command automatically")

	syncCmd.Flags().Bool("reverse", false, "Sync from 'tmux' to the history")
}
