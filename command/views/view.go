package views

import (
	"fmt"

	"github.com/hashicorp/terraform/command/format"
	"github.com/hashicorp/terraform/internal/terminal"
	"github.com/hashicorp/terraform/tfdiags"
	"github.com/mitchellh/colorstring"
)

type View struct {
	streams         *terminal.Streams
	colorize        *colorstring.Colorize
	compactWarnings bool
	configSources   func() map[string][]byte
}

func NewView(streams *terminal.Streams) *View {
	return &View{
		streams: streams,
		colorize: &colorstring.Colorize{
			Colors:  colorstring.DefaultColors,
			Disable: true,
			Reset:   true,
		},
		configSources: func() map[string][]byte { return nil },
	}
}

func (v *View) output(s string) {
	fmt.Fprint(v.streams.Stdout.File, s)
}

func (v *View) EnableColor(color bool) {
	v.colorize.Disable = !color
}

func (v *View) SetConfigSources(cb func() map[string][]byte) {
	v.configSources = cb
}

func (v *View) Diagnostics(diags tfdiags.Diagnostics) {
	diags.Sort()

	if len(diags) == 0 {
		return
	}

	diags = diags.ConsolidateWarnings(1)

	// Since warning messages are generally competing
	if v.compactWarnings {
		// If the user selected compact warnings and all of the diagnostics are
		// warnings then we'll use a more compact representation of the warnings
		// that only includes their summaries.
		// We show full warnings if there are also errors, because a warning
		// can sometimes serve as good context for a subsequent error.
		useCompact := true
		for _, diag := range diags {
			if diag.Severity() != tfdiags.Warning {
				useCompact = false
				break
			}
		}
		if useCompact {
			msg := format.DiagnosticWarningsCompact(diags, v.colorize)
			msg = "\n" + msg + "\nTo see the full warning notes, run Terraform without -compact-warnings.\n"
			v.output(msg)
			return
		}
	}

	for _, diag := range diags {
		var msg string
		if v.colorize.Disable {
			msg = format.DiagnosticPlain(diag, v.configSources(), v.streams.Stderr.Columns())
		} else {
			msg = format.Diagnostic(diag, v.configSources(), v.colorize, v.streams.Stderr.Columns())
		}

		if diag.Severity() == tfdiags.Error {
			fmt.Fprint(v.streams.Stderr.File, msg)
		} else {
			fmt.Fprint(v.streams.Stdout.File, msg)
		}
	}
}

func (v *View) HelpPrompt(command string) {
	fmt.Fprintf(v.streams.Stderr.File, helpPrompt, command)
}

const helpPrompt = `
For more help on using this command, run:
  terraform %s -help
`
