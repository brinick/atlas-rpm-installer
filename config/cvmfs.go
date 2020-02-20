package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/brinick/atlas-rpm-installer/pkg/filesystem/cvmfs"
)

// CvmfsOpts are options for CVMFS
type CvmfsOpts struct {
	cvmfs.Opts
}

func (c *CvmfsOpts) flags() {
	flag.StringVar(
		&c.Binary,
		"cvmfs.exe",
		"/usr/bin/cvmfs_server",
		"Path to the CVMFS server executable",
	)

	flag.StringVar(
		&c.NightlyRepo,
		"cvmfs.nightly-repo",
		"atlas-nightlies.cern.ch",
		"The CVMFS nightly repo name",
	)

	flag.StringVar(
		&c.GatewayNode,
		"cvmfs.gateway",
		"lxcvmfs78.cern.ch",
		"Gateway node for CVMFS operations",
	)

	flag.IntVar(
		&c.MaxTransactionAttempts,
		"cvmfs.max-transaction-attempts",
		10,
		"Maximum number of attempts to be made to open a transaction, before aborting",
	)
}

func (c *CvmfsOpts) validate() error {
	var min, max = 1, 30
	if c.MaxTransactionAttempts < min || c.MaxTransactionAttempts > max {
		return fmt.Errorf(
			"Max attempts to open a CVMFS transaction must be in range %d-%d",
			min,
			max,
		)
	}
	return nil
}

func (c *CvmfsOpts) String() string {
	return strings.Join(
		[]string{
			"- CVMFS Options:",
			fmt.Sprintf("   - Binary: %s", c.Binary),
			fmt.Sprintf("   - Gateway Node: %s", c.GatewayNode),
			fmt.Sprintf("   - Nightly Repo: %s", c.NightlyRepo),
			fmt.Sprintf("   - Max Open Transaction Attempts: %d", c.MaxTransactionAttempts),
		},
		"\n",
	)
}
