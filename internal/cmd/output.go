package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	iconSuccess = "󰄬"
	iconFailure = "󰅖"
	iconCurrent = "●"
)

type output struct {
	w      io.Writer
	pretty bool
	green  *color.Color
	red    *color.Color
	cyan   *color.Color
	yellow *color.Color
	header *color.Color
}

func commandOutput(cmd *cobra.Command) output {
	w := cmd.OutOrStdout()
	file, isFile := w.(*os.File)
	return newOutput(w, isFile && file == os.Stdout && !color.NoColor)
}

func newOutput(w io.Writer, pretty bool) output {
	o := output{
		w:      w,
		pretty: pretty,
		green:  color.New(color.FgGreen, color.Bold),
		red:    color.New(color.FgRed),
		cyan:   color.New(color.FgCyan),
		yellow: color.New(color.FgYellow),
		header: color.New(color.Bold),
	}
	for _, style := range []*color.Color{o.green, o.red, o.cyan, o.yellow, o.header} {
		if pretty {
			style.EnableColor()
		} else {
			style.DisableColor()
		}
	}
	return o
}

func (o output) table(header []string, rows [][]string, styles []*color.Color) error {
	var buffer bytes.Buffer
	w := tabwriter.NewWriter(&buffer, 0, 4, 1, ' ', 0)
	_, _ = fmt.Fprintln(w, strings.Join(header, "\t"))
	for _, row := range rows {
		_, _ = fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	if err := w.Flush(); err != nil {
		return err
	}

	lines := strings.Split(strings.TrimSuffix(buffer.String(), "\n"), "\n")
	if len(lines) != len(styles)+1 {
		return fmt.Errorf("table rows contain newlines")
	}
	if _, err := o.header.Fprintln(o.w, lines[0]); err != nil {
		return err
	}
	for i, line := range lines[1:] {
		if _, err := styles[i].Fprintln(o.w, line); err != nil {
			return err
		}
	}
	return nil
}

func (o output) successf(format string, args ...any) error {
	if !o.pretty {
		_, err := fmt.Fprintf(o.w, format+"\n", args...)
		return err
	}
	_, err := o.green.Fprintf(o.w, iconSuccess+" "+format+"\n", args...)
	return err
}

func (o output) marker(current bool) string {
	if !current {
		return " "
	}
	if !o.pretty {
		return "*"
	}
	return o.green.Sprint(iconCurrent)
}

func (o output) status(ok bool) string {
	if !o.pretty {
		return strconv.FormatBool(ok)
	}
	if ok {
		return o.green.Sprint(iconSuccess)
	}
	return o.red.Sprint(iconFailure)
}

func (o output) name(value string) string {
	return o.cyan.Sprint(value)
}

func (o output) detail(value any) string {
	return o.yellow.Sprint(value)
}
