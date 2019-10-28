package config

import (
	"flag"
	"os"
	"path/filepath"
)

type dirsOpts struct {
	// Base directory below which to install
	InstallBase string
	Logs        string
	WorkBase    string
	RPMSrcBase  string

	// Where is the repo for stable releases
	StableRelsDir string
}

func (d *dirsOpts) flags() {
	flag.StringVar(
		&d.InstallBase,
		"dirs.install",
		"",
		"Base directory below which to install",
	)

	flag.StringVar(
		&d.Logs,
		"dirs.logs",
		filepath.Join(os.Getenv("HOME"), "logs"),
		"Directory in which to create install logs (default: $HOME/logs/)",
	)

	flag.StringVar(
		&d.WorkBase,
		"dirs.work",
		os.Getenv("HOME"),
		"Directory in which to do work (default: $HOME)",
	)

	flag.StringVar(
		&d.RPMSrcBase,
		"dirs.rpmsrc",
		"",
		"Directory in which to find source RPMs",
	)

	flag.StringVar(
		&d.StableRelsDir,
		"dirs.stable-releases",
		"/cvmfs/atlas.cern.ch/repo/sw/software",
		"Directory where the repository for stable releases can be found",
	)
}

func (d *dirsOpts) validate() error {
	return nil
}
