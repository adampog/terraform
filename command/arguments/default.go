package arguments

import (
	"flag"
	"io/ioutil"
)

func defaultFlagSet(name string) *flag.FlagSet {
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.SetOutput(ioutil.Discard)

	// Disable default usage rendering
	f.Usage = func() {}

	return f
}
