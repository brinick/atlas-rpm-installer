package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/brinick/atlas-rpm-installer/pkg/ayum"
)

// AyumOpts are options for ayum
type AyumOpts struct {
	ayum.Opts
}

func (a *AyumOpts) flags() {
	flag.StringVar(
		&a.AyumDir,
		"ayum.dir",
		"",
		"Directory into which we download the AYUM source repo (default is value of the -dirs.work variable)",
	)

	flag.StringVar(
		&a.InstallDir,
		"ayum.install-dir",
		"",
		"Directory below which to install RPMs (default is value of the -dirs.install + <branchName>)",
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

	flag.StringVar(
		&a.MonitoringFormat,
		"ayum.monitoring-format",
		"",
		"Output monitoring format e.g. statsd (default no monitoring)",
	)
}

func (a *AyumOpts) validate() error {
	return nil
}

func (a *AyumOpts) String() string {
	return strings.Join(
		[]string{
			"- Ayum Options:",
			fmt.Sprintf("   - Src Repo: %s", a.SrcRepo),
			fmt.Sprintf("   - Ayum Dir: %s", a.AyumDir),
			fmt.Sprintf("   - Install Dir: %s", a.InstallDir),
			fmt.Sprintf("   - Download TimeOut: %ds", a.DownloadTimeout),
			fmt.Sprintf("   - Command TimeOut: %ds", a.Timeout),
			fmt.Sprintf("   - Install TimeOut: %ds", a.InstallTimeout),
			fmt.Sprintf("   - Monitoring Format: %s", a.MonitoringFormat),
		},
		"\n",
	)
}
