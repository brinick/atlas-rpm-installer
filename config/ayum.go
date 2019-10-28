package config

import (
	"flag"

	"github.com/brinick/atlas-rpm-installer/pkg/ayum"
)

type ayumOpts struct {
	ayum.Opts
}

func (a *ayumOpts) flags() {
	flag.StringVar(
		&a.AyumDir,
		"ayum.dir",
		"",
		"Directory into which we download the AYUM source repo",
	)

	flag.StringVar(
		&a.InstallDir,
		"ayum.install-dir",
		"",
		"Directory below which to install RPMs",
	)

	flag.StringVar(
		&a.SrcRepo,
		"ayum.src-repo",
		"https://gitlab.cern.ch/atlas-sit/ayum.git",
		"The source Git repo for ayum",
	)

	flag.IntVar(
		&a.DownloadTimeout,
		"ayum.download-timeout",
		60,
		"Maximum number of seconds to allow for git cloning the repo",
	)

	flag.IntVar(
		&a.Timeout,
		"ayum.cmd-timeout",
		60,
		"Maximum number of seconds to allow for running ayum commands like configure",
	)

	flag.IntVar(
		&a.InstallTimeout,
		"ayum.install-timeout",
		3600,
		"Maximum number of seconds to allow for running an ayum install command",
	)
}

func (a *ayumOpts) validate() error {
	return nil
}
