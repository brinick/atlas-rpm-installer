package config

import (
	"flag"
	"fmt"
	"strings"
)

// EosOpts are options for EOS
type EosOpts struct {
	BaseDir        string
	NightlyBaseDir string
}

func (e *EosOpts) flags() {
	flag.StringVar(
		&e.BaseDir,
		"eos.basedir",
		"/eos/project/a/atlas-software-dist/www/RPMs/",
		"Base directory in EOS for storing RPMs",
	)
	flag.StringVar(
		&e.NightlyBaseDir,
		"eos.nightly-basedir",
		"/eos/project/a/atlas-software-dist/www/RPMs/nightlies/",
		"Base directory in EOS for storing nightly RPMs",
	)
}

func (e *EosOpts) validate() error {
	return nil
}

func (e *EosOpts) String() string {
	return strings.Join(
		[]string{
			"- EOS Options:",
			fmt.Sprintf("   - Base Dir: %s", e.BaseDir),
			fmt.Sprintf("   - Nightly Base Dir: %s", e.NightlyBaseDir),
		},
		"\n",
	)
}
