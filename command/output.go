package command

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/command/views"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/tfdiags"
)

// OutputCommand is a Command implementation that reads an output
// from a Terraform state and prints it.
type OutputCommand struct {
	Meta
}

type outputArguments struct {
	name      string
	viewType  views.ViewType
	statePath string
}

func (c *OutputCommand) Run(cliArgs []string) int {
	// Parse and validate flags
	args, err := c.ParseArguments(cliArgs)
	if err != nil {
		c.Ui.Error(err.Error())
		c.Ui.Error(c.Help())
		return 1
	}

	view := views.NewOutput(args.viewType, c.View())

	// Fetch data from state
	outputs, diags := c.Outputs(args.statePath)
	if diags.HasErrors() {
		view.Diagnostics(diags)
		return 1
	}

	// Render the view
	viewDiags := view.Output(args.name, outputs)
	diags = diags.Append(viewDiags)

	view.Diagnostics(diags)

	if diags.HasErrors() {
		return 1
	}

	return 0
}

func (c *OutputCommand) ParseArguments(cliArgs []string) (*outputArguments, error) {
	// Extract -no-color
	cliArgs = c.Meta.process(cliArgs)

	args := &outputArguments{}

	var jsonOutput, rawOutput bool
	cmdFlags := c.Meta.defaultFlagSet("output")
	cmdFlags.BoolVar(&jsonOutput, "json", false, "json")
	cmdFlags.BoolVar(&rawOutput, "raw", false, "raw")
	cmdFlags.StringVar(&args.statePath, "state", "", "path")
	cmdFlags.Usage = func() { c.Ui.Error(c.Help()) }
	if err := cmdFlags.Parse(cliArgs); err != nil {
		return nil, fmt.Errorf("Error parsing command-line flags: %s\n", err.Error())
	}

	cliArgs = cmdFlags.Args()
	if len(cliArgs) > 1 {
		return nil, fmt.Errorf("The output command expects exactly one argument with the name\n" +
			"of an output variable or no arguments to show all outputs.\n")
	}

	if jsonOutput && rawOutput {
		return nil, fmt.Errorf("The -raw and -json options are mutually-exclusive.\n")
	}

	if rawOutput && len(cliArgs) == 0 {
		return nil, fmt.Errorf("You must give the name of a single output value when using the -raw option.\n")
	}

	switch {
	case jsonOutput:
		args.viewType = views.ViewJSON
	case rawOutput:
		args.viewType = views.ViewRaw
	default:
		args.viewType = views.ViewHuman
	}

	if len(cliArgs) > 0 {
		args.name = cliArgs[0]
	}

	return args, nil
}

func (c *OutputCommand) Outputs(statePath string) (map[string]*states.OutputValue, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics

	// Allow state path override
	if statePath != "" {
		c.Meta.statePath = statePath
	}

	// Load the backend
	b, backendDiags := c.Backend(nil)
	diags = diags.Append(backendDiags)
	if diags.HasErrors() {
		return nil, diags
	}

	// This is a read-only command
	c.ignoreRemoteBackendVersionConflict(b)

	env, err := c.Workspace()
	if err != nil {
		diags = diags.Append(fmt.Errorf("Error selecting workspace: %s", err))
		return nil, diags
	}

	// Get the state
	stateStore, err := b.StateMgr(env)
	if err != nil {
		diags = diags.Append(fmt.Errorf("Failed to load state: %s", err))
		return nil, diags
	}

	if err := stateStore.RefreshState(); err != nil {
		diags = diags.Append(fmt.Errorf("Failed to load state: %s", err))
		return nil, diags
	}

	state := stateStore.State()
	if state == nil {
		state = states.NewState()
	}

	return state.RootModule().OutputValues, nil
}

func (c *OutputCommand) Help() string {
	helpText := `
Usage: terraform output [options] [NAME]

  Reads an output variable from a Terraform state file and prints
  the value. With no additional arguments, output will display all
  the outputs for the root module.  If NAME is not specified, all
  outputs are printed.

Options:

  -state=path      Path to the state file to read. Defaults to
                   "terraform.tfstate".

  -no-color        If specified, output won't contain any color.

  -json            If specified, machine readable output will be
                   printed in JSON format.

  -raw             For value types that can be automatically
                   converted to a string, will print the raw
                   string directly, rather than a human-oriented
                   representation of the value.
`
	return strings.TrimSpace(helpText)
}

func (c *OutputCommand) Synopsis() string {
	return "Show output values from your root module"
}
