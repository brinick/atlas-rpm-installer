package config

import "flag"

type eosOpts struct {
	BaseDir        string
	NightlyBaseDir string
}

func (e *eosOpts) flags() {
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

func (e *eosOpts) validate() error {
	return nil
}
