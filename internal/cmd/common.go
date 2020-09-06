// Copyright © 2020 The Homeport Team
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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"strings"

	"github.com/gonvenience/bunt"
	"github.com/gonvenience/neat"
	"github.com/gonvenience/wrap"
	"github.com/gonvenience/ytbx"
	"github.com/homeport/dyff/pkg/dyff"
	"github.com/spf13/cobra"
	yamlv3 "gopkg.in/yaml.v3"
)

const defaultOutputStyle = "human"

type reportConfig struct {
	style              string
	ignoreOrderChanges bool
	noTableStyle       bool
	doNotInspectCerts  bool
	exitWithCount      bool
	omitHeader         bool
	useGoPatchPaths    bool
}

var reportOptions reportConfig

func applyReportOptionsFlags(cmd *cobra.Command) {
	// Compare options
	cmd.PersistentFlags().BoolVarP(&reportOptions.ignoreOrderChanges, "ignore-order-changes", "i", false, "ignore order changes in lists")

	// Main output preferences
	cmd.PersistentFlags().StringVarP(&reportOptions.style, "output", "o", defaultOutputStyle, "specify the output style, supported styles: human, or brief")
	cmd.PersistentFlags().BoolVarP(&reportOptions.omitHeader, "omit-header", "b", false, "omit the dyff summary header")
	cmd.PersistentFlags().BoolVarP(&reportOptions.exitWithCount, "set-exit-status", "s", false, "set exit status to number of diff (capped at 255)")

	// Human/BOSH output related flags
	cmd.PersistentFlags().BoolVarP(&reportOptions.noTableStyle, "no-table-style", "l", false, "do not place blocks next to each other, always use one row per text block")
	cmd.PersistentFlags().BoolVarP(&reportOptions.doNotInspectCerts, "no-cert-inspection", "x", false, "disable x509 certificate inspection, compare as raw text")
	cmd.PersistentFlags().BoolVarP(&reportOptions.useGoPatchPaths, "use-go-patch-style", "g", false, "use Go-Patch style paths in outputs")
}

// OutputWriter encapsulates the required fields to define the look and feel of
// the output
type OutputWriter struct {
	PlainMode        bool
	Restructure      bool
	OmitIndentHelper bool
	OutputStyle      string
}

// WriteToStdout is a convenience function to write the content of the documents
// stored in the provided input file to the standard output
func (w *OutputWriter) WriteToStdout(filename string) error {
	if err := w.write(os.Stdout, filename); err != nil {
		return wrap.Errorf(err, "failed to write output _%s_", filename)
	}

	return nil
}

// WriteInplace writes the content of the documents stored in the provided input
// file to the file itself overwriting the conent in place.
func (w *OutputWriter) WriteInplace(filename string) error {
	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)

	// Force plain mode to make sure there are no ANSI sequences
	w.PlainMode = true
	if err := w.write(bufWriter, filename); err != nil {
		return wrap.Errorf(err, "failed to write output _%s_", filename)
	}

	// Write the buffered output to the provided input file (override in place)
	bufWriter.Flush()
	if err := ioutil.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return wrap.Errorf(err, "failed to overwrite file _%s_ in place", filename)
	}

	return nil
}

func (w *OutputWriter) write(writer io.Writer, filename string) error {
	inputFile, err := ytbx.LoadFile(filename)
	if err != nil {
		return wrap.Errorf(err, "failed to load input file _%s_", filename)
	}

	for _, document := range inputFile.Documents {
		if w.Restructure {
			ytbx.RestructureObject(document)
		}

		switch {
		case w.PlainMode && w.OutputStyle == "json":
			output, err := neat.NewOutputProcessor(false, false, &neat.DefaultColorSchema).ToCompactJSON(document)
			if err != nil {
				return err
			}
			fmt.Fprintf(writer, "%s\n", output)

		case w.PlainMode && w.OutputStyle == "yaml":
			encoder := yamlv3.NewEncoder(writer)
			encoder.SetIndent(2)
			encoder.Encode(document)
			encoder.Close()

		case w.OutputStyle == "json":
			output, err := neat.NewOutputProcessor(!w.OmitIndentHelper, true, &neat.DefaultColorSchema).ToJSON(document)
			if err != nil {
				return err
			}
			fmt.Fprintf(writer, "%s\n", output)

		case w.OutputStyle == "yaml":
			output, err := neat.NewOutputProcessor(!w.OmitIndentHelper, true, &neat.DefaultColorSchema).ToYAML(document)
			if err != nil {
				return err
			}
			fmt.Fprintf(writer, "%s\n", output)
		}
	}

	return nil
}

func parseSetting(setting string) (bunt.SwitchState, error) {
	switch strings.ToLower(setting) {
	case "auto":
		return bunt.AUTO, nil

	case "off", "no", "false":
		return bunt.OFF, nil

	case "on", "yes", "true":
		return bunt.ON, nil

	default:
		return bunt.OFF, fmt.Errorf("invalid state '%s' used, supported modes are: auto, on, or off", setting)
	}
}

func initSettings() {
	if debugMode {
		dyff.SetLoggingLevel(dyff.DEBUG)
	}
}

func writeReport(cmd *cobra.Command, report dyff.Report) error {
	var reportWriter dyff.ReportWriter
	switch strings.ToLower(reportOptions.style) {
	case "human", "bosh":
		reportWriter = &dyff.HumanReport{
			Report:               report,
			DoNotInspectCerts:    reportOptions.doNotInspectCerts,
			NoTableStyle:         reportOptions.noTableStyle,
			OmitHeader:           reportOptions.omitHeader,
			UseGoPatchPaths:      reportOptions.useGoPatchPaths,
			MinorChangeThreshold: 0.1,
		}

	case "brief", "short", "summary":
		reportWriter = &dyff.BriefReport{
			Report: report,
		}

	default:
		return wrap.Errorf(
			fmt.Errorf(cmd.UsageString()),
			"unknown output style %s", reportOptions.style,
		)
	}

	if err := reportWriter.WriteReport(os.Stdout); err != nil {
		return wrap.Errorf(err, "failed to print report")
	}

	// If configured, make sure `dyff` exists with an exit status
	if reportOptions.exitWithCount {
		return ExitCode{
			Value: int(math.Min(float64(len(report.Diffs)), 255.0)),
		}
	}

	return nil
}
