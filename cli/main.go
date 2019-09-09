package cli

import "flag"

// Parse parses and validates the command line args
func Parse() (*Args, error) {
	var args Args
	args.parse()
	err := args.validate()
	return &args, err
}

// Args is the full set of available command line args
type Args struct {
	Global  *globalOpts
	Install *installOpts
	Dirs    *pathsOpts
	Ayum    *ayumOpts
	CVMFS   *cvmfsOpts
}

func (a *Args) parse() {
	flag.Parse()
}

// flags defines all the flags for this struct
func (a *Args) flags() {
	a.Global.flags()
	a.Install.flags()
	a.Dirs.flags()
	a.CVMFS.flags()
	a.Ayum.flags()
}

// Validate validates the parsed args
func (a *Args) validate() error {
	return nil
}
