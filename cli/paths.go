package cli

import "flag"

type pathsOpts struct {
	Install    string
	Logs       string
	Work       string
	RPMSrcBase string
}

func (p *pathsOpts) flags() {
	flag.StringVar(
		&p.Install,
		"dirs.install",
		"",
		"Directory in which to install",
	)

	flag.StringVar(
		&p.Logs,
		"dirs.logs",
		"",
		"Directory in which to create install logs",
	)

	flag.StringVar(
		&p.Work,
		"dirs.work",
		"",
		"Directory in which to do work",
	)

	flag.StringVar(
		&p.RPMSrcBase,
		"dirs.rpmsrc",
		"",
		"Directory in which to find source RPMs",
	)
}
