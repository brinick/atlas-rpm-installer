package config

import "flag"

type globalOpts struct {
	TimeOut     int
	EmailOnFail []string
}

func (g *globalOpts) flags() {
	flag.IntVar(
		&g.TimeOut,
		"global.timeout",
		0,
		"Number of seconds after which the whole install process should abort",
	)
}

func (g *globalOpts) validate() error {
	return nil
}
