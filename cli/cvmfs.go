package cli

import (
	"flag"

	"github.com/brinick/atlas-rpm-installer/pkg/fs/cvmfs"
)

func (c *cvmfs.Opts) flags() {
	flag.SrtingVar(
		&c.Binary,
		"cvmfs.exe",
		"/usr/bin/cvmfs_server",
		"Path to the CVMFS server executable"
	)

	flag.StringVar(
		&c.NightlyRepo,
		"cvmfs.nightly-repo",
		"atlas-nightlies.cern.ch",
		"The CVMFS nightly repo name",
	)
	flag.StringVar(
		&c.StableRepoDir,
		"cvmfs.stable-releases-dir",
		"atlas-nightlies.cern.ch",
		"The CVMFS directory where stable releases can be found",
	)

	flag.StringVar(
		&c.GatewayNode,
		"cvmfs.gateway",
		"lxcvmfs78.cern.ch",
		"Gateway node for CVMFS operations",
	)

	flag.IntVar(
		&c.MaxTransitionAttempts,
		"cvmfs.max-transition-attempts",
		10,
		"Maximum number of attempts to be made to open a transaction, before aborting",
	)
}

func (c *cvmfs.Opts) validate() {
	//
}
