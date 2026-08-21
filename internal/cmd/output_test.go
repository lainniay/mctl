package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestOutputPlainKeepsScriptFriendlyText(t *testing.T) {
	var buffer bytes.Buffer
	out := newOutput(&buffer, false)

	if err := out.successf("current node: %s", "Japan-01"); err != nil {
		t.Fatal(err)
	}
	if got := buffer.String(); got != "current node: Japan-01\n" {
		t.Fatalf("success output = %q", got)
	}
	if out.marker(true) != "*" || out.status(true) != "true" || out.name("Japan-01") != "Japan-01" {
		t.Fatal("plain list formatting changed")
	}
}

func TestOutputPrettyUsesColorAndIcons(t *testing.T) {
	var buffer bytes.Buffer
	out := newOutput(&buffer, true)

	if err := out.table(
		[]string{"", "GROUP", "TYPE"},
		[][]string{{iconCurrent, "NodeSelection", "Selector"}},
		[]*color.Color{out.green},
	); err != nil {
		t.Fatal(err)
	}
	table := buffer.String()
	for _, want := range []string{"\x1b[", "GROUP", iconCurrent + " NodeSelection"} {
		if !strings.Contains(table, want) {
			t.Fatalf("pretty table %q does not contain %q", table, want)
		}
	}
	if err := out.successf("configuration is valid"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buffer.String(), iconSuccess) {
		t.Fatalf("pretty success output = %q", buffer.String())
	}
	if !strings.Contains(out.status(false), iconFailure) {
		t.Fatalf("pretty failure status = %q", out.status(false))
	}
}

func TestOutputTableRejectsNewlines(t *testing.T) {
	out := newOutput(&bytes.Buffer{}, true)
	if err := out.table([]string{"NAME"}, [][]string{{"bad\nname"}}, []*color.Color{out.red}); err == nil {
		t.Fatal("table accepted a multiline cell")
	}
}
