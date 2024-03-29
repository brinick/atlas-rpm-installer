package config

import (
	"flag"
	"fmt"
	"strings"
)

// GlobalOpts are options for the global installation context
type GlobalOpts struct {
	TimeOut int
}

func (g *GlobalOpts) flags() {
	flag.IntVar(
		&g.TimeOut,
		"global.timeout",
		0,
		"Integer number of seconds after which the whole install process should abort (default 0 i.e. no timeout)",
	)
}

func (g *GlobalOpts) validate() error {
	return nil
}

func (g *GlobalOpts) String() string {
	return strings.Join(
		[]string{
			"- Global Options:",
			fmt.Sprintf("   - Time out: %ds", g.TimeOut),
		},
		"\n",
	)
}
