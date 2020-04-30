package config

import (
	"flag"
	"fmt"

	"github.com/brinick/atlas-rpm-installer/pkg/filesystem/afs"
)

// AfsOpts are options for configuring AFS installation
type AfsOpts struct {
	afs.Opts
}

func (a *AfsOpts) flags() {
	flag.StringVar(
		&a.SudoUser,
		"afs.sudo-user",
		"",
		"The sudo user, if any, required to install nightlies on AFS (default none)",
	)

	flag.IntVar(
		&a.MaxTransactionAttempts,
		"afs.max-transaction-attempts",
		10,
		"Max number of attempts to be made to open a transaction, before aborting",
	)
}

func (a *AfsOpts) validate() error {
	var min, max = 1, 10
	if a.MaxTransactionAttempts < min || a.MaxTransactionAttempts > max {
		return fmt.Errorf(
			"Max attempts to open an AFS transaction must be in range %d-%d",
			min,
			max,
		)
	}
	return nil
}
