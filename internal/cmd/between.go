// Copyright © 2018 Matthias Diester
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/HeavyWombat/color"
	"github.com/HeavyWombat/dyff/pkg/dyff"
	"github.com/spf13/cobra"
)

var style string
var swap bool

// betweenCmd represents the between command
var betweenCmd = &cobra.Command{
	Use:   "between",
	Short: "Compares differences between documents",
	Long: `
Compares differences between documents and displays the delta. Supported
document types are: YAML (http://yaml.org/) and JSON (http://json.org/).

`,
	Args:    cobra.ExactArgs(2),
	Aliases: []string{"bw"},
	Run: func(cmd *cobra.Command, args []string) {
		var fromLocation, toLocation string
		if swap {
			fromLocation = args[1]
			toLocation = args[0]
		} else {
			fromLocation = args[0]
			toLocation = args[1]
		}

		from, to, err := dyff.LoadFiles(fromLocation, toLocation)
		if err != nil {
			dyff.ExitWithError("Failed to load input files", err)
		}

		diffs := dyff.CompareInputFiles(from, to)

		// TODO Add style Go-Patch
		// TODO Add style Spruce
		// TODO Add style JSON report
		// TODO Add style YAML report
		// TODO Add style one-line report

		switch strings.ToLower(style) {
		case "human", "bosh":
			fmt.Printf(`      _        __  __
    _| |_   _ / _|/ _|  between %s
  / _' | | | | |_| |_       and %s
 | (_| | |_| |  _|  _|
  \__,_|\__, |_| |_|   returned %s
        |___/
`, niceLocation(fromLocation),
				niceLocation(toLocation),
				dyff.Bold(dyff.Plural(len(diffs), "difference")))
			fmt.Print(dyff.DiffsToHumanStyle(diffs))

		default:
			fmt.Printf("Unknown output style %s\n", style)
			cmd.Usage()
		}
	},
}

func niceLocation(location string) string {
	if location == "-" {
		return dyff.Italic("<stdin>")
	}

	if _, err := os.Stat(location); err == nil {
		if abs, err := filepath.Abs(location); err == nil {
			return dyff.Bold(abs)
		}
	}

	if _, err := url.ParseRequestURI(location); err == nil {
		return dyff.Color(location, color.FgHiBlue, color.Underline)
	}

	return location
}

func init() {
	rootCmd.AddCommand(betweenCmd)

	// TODO Add flag for filter on path
	betweenCmd.PersistentFlags().StringVarP(&style, "output", "o", "human", "Specify the output style, e.g. 'human' (more to come ...)")
	betweenCmd.PersistentFlags().BoolVarP(&swap, "swap", "s", false, "Swap `from` and `to` for compare")
	betweenCmd.PersistentFlags().BoolVarP(&dyff.NoTableStyle, "no-table-style", "t", false, "Disable the table output")
	betweenCmd.PersistentFlags().BoolVarP(&dyff.DoNotInspectCerts, "no-cert-inspection", "c", false, "Disable certificate inspection (compare as raw text)")
	betweenCmd.PersistentFlags().BoolVarP(&dyff.UseGoPatchPaths, "use-go-patch-style", "g", false, "Use Go-Patch style paths instead of Spruce Dot-Style")
}
