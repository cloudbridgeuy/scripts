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
	"sort"
	"strconv"
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

func findDirectories(config directory) ([]string, error) {
	regex := regexp.MustCompile(config.grep)
	cmd := exec.Command(
		"/usr/bin/find",
		config.path,
		"-mindepth",
		strconv.Itoa(config.mindepth),
		"-maxdepth",
		strconv.Itoa(config.maxdepth),
		"-type",
		"d",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			return nil, fmt.Errorf("%w: %s", err, message)
		}
		return nil, err
	}

	var matches []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if regex.MatchString(trimmed) {
			matches = append(matches, trimmed)
		}
	}

	return matches, nil
}

func reverseSyncPlan(history []string, sessions []string) (toCreate []string, toKill []string) {
	inHistory := make(map[string]bool, len(history))
	inTmux := make(map[string]bool, len(sessions))

	for _, session := range history {
		inHistory[session] = true
	}

	for _, session := range sessions {
		inTmux[session] = true
	}

	for _, session := range history {
		if !inTmux[session] {
			toCreate = append(toCreate, session)
		}
	}

	for _, session := range sessions {
		if !inHistory[session] {
			toKill = append(toKill, session)
		}
	}

	return toCreate, toKill
}

func lastSession(sessions []string) (string, bool) {
	if len(sessions) == 0 {
		return "", false
	}

	return sessions[len(sessions)-1], true
}

func rotateHistoryPrev(history []string) []string {
	sessionsLength := len(history)
	sessions := make([]string, sessionsLength)

	for i, session := range history {
		if i == sessionsLength-1 {
			sessions[0] = session
		} else {
			sessions[i+1] = session
		}
	}

	return sessions
}

func rotateHistoryNext(history []string) []string {
	sessionsLength := len(history)
	sessions := make([]string, sessionsLength)

	for i, session := range history {
		if i == 0 {
			sessions[sessionsLength-1] = session
		} else {
			sessions[i-1] = session
		}
	}

	return sessions
}

func switchWithRotation(history []string, rotate func([]string) []string) ([]string, string, error) {
	if len(history) == 0 {
		return nil, "", fmt.Errorf("No sessions found in history")
	}

	available, err := existingHistorySessions(history)
	if err != nil {
		return nil, "", err
	}

	if len(available) == 0 {
		return []string{}, "", fmt.Errorf("no available sessions found in history")
	}

	rotated := available
	var lastErr error

	for attempts := 0; attempts < len(available); attempts++ {
		rotated = rotate(rotated)
		session := rotated[len(rotated)-1]

		if err := tmux.SwitchExisting(session); err == nil {
			return rotated, session, nil
		} else {
			lastErr = err
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no reachable session found")
	}

	return nil, "", lastErr
}

func existingHistorySessions(history []string) ([]string, error) {
	seen := make(map[string]bool, len(history))
	available := make([]string, 0, len(history))

	for _, session := range history {
		if session == "" || seen[session] {
			continue
		}

		exists, err := tmux.SessionExists(session)
		if err != nil {
			return nil, err
		}

		if !exists {
			continue
		}

		seen[session] = true
		available = append(available, session)
	}

	return available, nil
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
				path:     home + "/Projects/Bare",
				mindepth: 2,
				maxdepth: 2,
				grep:     ".*/Projects/Bare/[^/]*/[^/]*",
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
			wg             sync.WaitGroup
			mu             sync.Mutex
			allDirectories []string
			errChan        = make(chan error, len(directories))
		)

		for _, dir := range directories {
			wg.Add(1)
			go func(dir directory) {
				defer wg.Done()

				logger.Debugf("/usr/bin/find %s -mindepth %d -maxdepth %d -type d", dir.path, dir.mindepth, dir.maxdepth)
				matches, err := findDirectories(dir)
				if err != nil {
					logger.Errorf("find %s -mindepth %d -maxdepth %d -type d", dir.path, dir.mindepth, dir.maxdepth)
					logger.Errorf(err.Error())
					errChan <- err
					return
				}

				mu.Lock()
				allDirectories = append(allDirectories, matches...)
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

		if len(allDirectories) == 0 {
			errors.HandleErrorWithReason(fmt.Errorf("no directories found"), "can't evaluate the list of possible directories")
			return
		}

		deduped := make(map[string]bool, len(allDirectories))
		var sortedDirectories []string
		for _, directory := range allDirectories {
			if deduped[directory] {
				continue
			}
			deduped[directory] = true
			sortedDirectories = append(sortedDirectories, directory)
		}
		sort.Strings(sortedDirectories)

		session, err := script.
			Echo(strings.Join(sortedDirectories, "\n")).
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
		if session == "" {
			return
		}

		if err = tmux.Switch(session); err != nil {
			errors.HandleErrorWithReason(err, fmt.Sprintf("can't switch to session %s", session))
			return
		}

		addToTmuxHistory(session)

		if err := saveConfig(); err != nil {
			errors.HandleErrorWithReason(err, "can't save the config file")
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

		addToTmuxHistory(session)

		if err := saveConfig(); err != nil {
			errors.HandleErrorWithReason(err, "can't save the config file")
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
			return
		}

		sessions, err := tmux.Ls()
		if err != nil {
			errors.HandleErrorWithReason(err, "can't list tmux sessions")
			return
		}
		history := getTmuxHistory()

		if reverse {
			toCreate, toKill := reverseSyncPlan(history, sessions)

			for _, session := range toCreate {
				if err := tmux.NewSession(session); err != nil {
					errors.HandleErrorWithReason(err, fmt.Sprintf("can't create session %s", session))
					return
				}
			}

			for _, session := range toKill {
				if err := tmux.KillSession(session); err != nil {
					errors.HandleErrorWithReason(err, fmt.Sprintf("can't kill session %s", session))
					return
				}
			}

			sessions, err = tmux.Ls()
			if err != nil {
				errors.HandleErrorWithReason(err, "can't list tmux sessions")
				return
			}
		}

		setTmuxHistory(sessions)

		if err := saveConfig(); err != nil {
			errors.HandleErrorWithReason(err, "can't save the config file")
			return
		}

		session, ok := lastSession(sessions)
		if !ok {
			return
		}

		if err = tmux.Switch(session); err != nil {
			errors.HandleErrorWithReason(err, fmt.Sprintf("can't switch to session %s", session))
			return
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

		sessions, _, err := switchWithRotation(history, rotateHistoryPrev)
		if err != nil {
			errors.HandleErrorWithReason(err, "can't switch to a previous session")
			return
		}

		setTmuxHistory(sessions)

		if err := saveConfig(); err != nil {
			errors.HandleErrorWithReason(err, "can't save the config file")
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

		sessions, _, err := switchWithRotation(history, rotateHistoryNext)
		if err != nil {
			errors.HandleErrorWithReason(err, "can't switch to a next session")
			return
		}

		setTmuxHistory(sessions)

		if err := saveConfig(); err != nil {
			errors.HandleErrorWithReason(err, "can't save the config file")
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

		err := tmux.Switch(session)
		if err != nil {
			errors.HandleErrorWithReason(err, fmt.Sprintf("can't switch to session %s", session))
			return
		}

		addToTmuxHistory(session)

		if err := saveConfig(); err != nil {
			errors.HandleErrorWithReason(err, "can't save the config file")
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

		if err := tmux.Switch(session); err != nil {
			errors.HandleErrorWithReason(err, fmt.Sprintf("can't switch to session %s", session))
			return
		}

		addToTmuxHistory(session)

		if err := saveConfig(); err != nil {
			errors.HandleErrorWithReason(err, "can't save the config file")
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

	displayCmd.Flags().Bool("no-switch", false, "Display sessions without switching")

	syncCmd.Flags().Bool("reverse", false, "Sync from history to 'tmux'")
}
