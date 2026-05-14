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
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cloudbridgeuy/scripts/pkg/errors"
	"github.com/cloudbridgeuy/scripts/pkg/logger"
	"github.com/cloudbridgeuy/scripts/pkg/markdown"
	"github.com/spf13/cobra"
)

var markdownCmd = &cobra.Command{
	Use:     "markdown [flags] <FILE>",
	Aliases: []string{"md"},
	Short:   "Convert a Markdown file into a styled HTML page",
	Long: `Converts a Markdown file into a self-styled HTML page with a terminal
aesthetic, the tokyonight-night palette, syntax-highlighted code fences, and
client-side Mermaid diagram rendering.

The HTML is written beside the source file with a .html extension by default.
Use --output to choose another path, and --open to view the result in the
default browser.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		outputFlag, err := cmd.Flags().GetString("output")
		if err != nil {
			errors.HandleErrorWithReason(err, "Can't get the --output flag")
		}

		open, err := cmd.Flags().GetBool("open")
		if err != nil {
			errors.HandleErrorWithReason(err, "Can't get the --open flag")
		}

		cfg := markdown.NewRenderConfig(args[0], outputFlag, open)
		logger.Debug("resolved render config", "input", cfg.InputPath, "output", cfg.OutputPath)

		src, err := os.ReadFile(cfg.InputPath)
		if err != nil {
			errors.HandleErrorWithReason(err, "Can't read the input file")
		}

		body := markdown.StripFrontmatter(src)

		fallback := strings.TrimSuffix(filepath.Base(cfg.InputPath), filepath.Ext(cfg.InputPath))
		title := markdown.ExtractTitle(body, fallback)

		htmlBody, err := markdown.RenderMarkdown(body)
		if err != nil {
			errors.HandleErrorWithReason(err, "Can't render the Markdown")
		}

		chromaCSS, err := markdown.ChromaCSS()
		if err != nil {
			errors.HandleErrorWithReason(err, "Can't generate the syntax-highlighting CSS")
		}

		page := markdown.BuildPage(htmlBody, title, chromaCSS)

		if err := os.WriteFile(cfg.OutputPath, []byte(page), 0644); err != nil {
			errors.HandleErrorWithReason(err, "Can't write the output file")
		}

		logger.Info("wrote HTML page", "path", cfg.OutputPath)

		if cfg.Open {
			if err := openBrowser(cfg.OutputPath); err != nil {
				errors.HandleErrorWithReason(err, "Can't open the browser")
			}
		}

		fmt.Println(cfg.OutputPath)
	},
}

// openBrowser opens path in the system default browser. It carries no unit
// test because it delegates entirely to the OS launcher.
func openBrowser(path string) error {
	opener := "xdg-open"
	if runtime.GOOS == "darwin" {
		opener = "open"
	}
	return exec.Command(opener, path).Start()
}

func init() {
	rootCmd.AddCommand(markdownCmd)
	markdownCmd.Flags().StringP("output", "o", "", "Write HTML to this path instead of the default sibling path")
	markdownCmd.Flags().Bool("open", false, "Open the result in the default browser")
}
