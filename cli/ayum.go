package cli

import (
	"flag"

	"github.com/brinick/atlas-rpm-installer/pkg/ayum"
)

func (a *ayum.Opts) flags() {
	flag.StringVar(
		&a.Dir,
		"ayum.dir",
		"",
		"Directory into which we download the AYUM source repo",
	)

	flag.StringVar(
		&a.Repo,
		"ayum.src-repo",
		"https://gitlab.cern.ch/atlas-sit/ayum.git",
		"The source Git repo for ayum",
	)

	flag.IntVar(
		&a.GitCloneTimeout,
		"ayum.download-timeout",
		60,
		"Maximum number of seconds to allow for git cloning the repo",
	)

}
