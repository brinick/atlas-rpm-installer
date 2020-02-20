package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DirsOpts are options for various directories
type DirsOpts struct {
	// Base directory below which to install
	InstallBase string
	Logs        string
	WorkBase    string
	RPMSrcBase  string

	// Where is the repo for stable releases
	StableRelsDir string
}

func (d *DirsOpts) flags() {
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

func (d *DirsOpts) validate() error {
	return nil
}

func (d *DirsOpts) String() string {
	return strings.Join(
		[]string{
			"- Directories Options:",
			fmt.Sprintf("   - Install base: %s", d.InstallBase),
			fmt.Sprintf("   - Work base: %s", d.WorkBase),
			fmt.Sprintf("   - Logs dir: %s", d.Logs),
			fmt.Sprintf("   - RPM src base: %s", d.RPMSrcBase),
			fmt.Sprintf("   - Stable releases dir: %s", d.StableRelsDir),
		},
		"\n",
	)
}
