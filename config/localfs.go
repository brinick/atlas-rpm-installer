package config

import (
	"flag"
	"fmt"

	"github.com/brinick/atlas-rpm-installer/pkg/filesystem/localfs"
)

// LocalfsOpts are options for configuring local filesystem installation
type LocalfsOpts struct {
	localfs.Opts
}

func (l *LocalfsOpts) flags() {
	flag.StringVar(
		&l.SudoUser,
		"localfs.sudo-user",
		"",
		"The sudo user, if any, required to install nightlies locally (default none)",
	)

	flag.IntVar(
		&l.MaxTransactionAttempts,
		"localfs.max-transaction-attempts",
		10,
		"Max number of attempts to be made to open a transaction, before aborting",
	)
}

func (l *LocalfsOpts) validate() error {
	var min, max = 1, 10
	if l.MaxTransactionAttempts < min || l.MaxTransactionAttempts > max {
		return fmt.Errorf(
			"Max attempts to open a local FS transaction must be in range %d-%d",
			min,
			max,
		)
	}
	return nil
}
