package cli

import "flag"

type globalOpts struct {
	TimeOut int
}

func (g *globalOpts) flags() {
	flag.IntVar(
		&g.TimeOut,
		"global.timeout",
		0,
		"Number of seconds after which the whole install process should abort",
	)
}
