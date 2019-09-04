package installer

import (
	"flag"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/brinick/atlas-rpm-installer/config"
)

// ParseArgs returns the args struct parsed from the command line
func ParseArgs() (Args, error) {
	// Parse the command line args
	var a Args
	if err := a.parse().validate(); err != nil {
		return a, err
	}
	if err := a.ensureAbsPaths(); err != nil {
		return a, err
	}

	return a, nil
}

// ------------------------------------------------------------------

// Args is the bag of information required to perform an install
type Args struct {
	Release   string
	Branch    string
	Platform  string
	Timestamp string
	Project   string

	// Install base directory below which we install the RPMs
	Installdir string

	// Work base directory below which we
	Workdir string

	// Base directory below which we will find RPMs to install
	RPMSrcBaseDir string

	// Number of seconds after which the install should abort
	Timeout int
}

// ID returns a string representation of this Args struct that can
// be used to identify it
func (a *Args) ID() string {
	now := strconv.FormatInt(time.Now().Unix(), 10)
	time.Now().Unix()
	return strings.Join([]string{
		a.Branch,
		a.Project,
		a.Platform,
		a.Timestamp,
		now,
	}, "__")
}

// ensureAbsPaths converts all paths to absolute
func (a *Args) ensureAbsPaths() error {
	// Make sure we are dealing with absolute paths
	var err error
	a.Installdir, err = filepath.Abs(a.Installdir)
	if err != nil {
		return err
	}
	a.RPMSrcBaseDir, err = filepath.Abs(a.RPMSrcBaseDir)
	if err != nil {
		return err
	}
	a.Workdir, err = filepath.Abs(a.Workdir)
	if err != nil {
		return err
	}

	return nil
}

// Usage gets the command line args help
func (a *Args) Usage() string {
	return "TODO: add usage --"
}

func (a *Args) parse() *Args {
	// --------------------------------------------------
	// - Required args -----
	// --------------------------------------------------
	flag.StringVar(
		&a.Release,
		"r",
		"",
		`[Required] The release to install 
		e.g 21.2/x86_64-slc6-gcc62-opt/2019-12-31T1234 [required]`,
	)

	flag.StringVar(
		&a.Project,
		"p",
		"",
		"[Required] The project to install e.g. Athena",
	)

	// --------------------------------------------------
	// - Optional args -----
	// --------------------------------------------------
	defaultRPMBaseDir := config.EOS.NightlyBaseDir
	flag.StringVar(
		&a.RPMSrcBaseDir,
		"rpmBasedir",
		defaultRPMBaseDir,
		fmt.Sprintf(
			"The base directory below which RPMs to install are found (default: %s)",
			defaultRPMBaseDir,
		),
	)

	defaultInstallBase := config.Paths.InstallBase
	flag.StringVar(
		&a.Installdir,
		"installdir",
		defaultInstallBase,
		fmt.Sprintf("The install base directory to use (default: %s)", defaultInstallBase),
	)

	defaultWorkBase := config.Paths.WorkBase
	flag.StringVar(
		&a.Workdir,
		"workdir",
		defaultWorkBase,
		fmt.Sprintf("The work base directory to use (default: %s)", defaultWorkBase),
	)

	flag.IntVar(
		&a.Timeout,
		"global.timeout",
		0,
		"Number of seconds after which the install attempt should time out (default=0 => never time out)",
	)

	flag.Parse()
	return a
}

func (a *Args) validate() error {
	tokens := strings.Split(a.Release, "/")
	if len(tokens) != 3 {
		msg := "-r argument expected of the form:\n"
		msg += "<branch>/<platform>/<timestamp>"
		return fmt.Errorf(msg)
	}

	a.Branch, a.Platform, a.Timestamp = tokens[0], tokens[1], tokens[2]
	if err := a.validatePlatform(); err != nil {
		return err
	}

	if err := a.validateTimestamp(); err != nil {
		return err
	}

	return nil
}

func (a *Args) validatePlatform() error {
	toks := strings.Split(a.Platform, "-")
	if len(toks) != 4 {
		return fmt.Errorf("Badly formed platform '%s'", a.Platform)
	}

	// helper function
	validate := func(needle string, haystack []string) error {
		if !contains(needle, haystack) {
			msg := "Platform contains illegal component %s. Legal: %s"
			return fmt.Errorf(msg, needle, strings.Join(haystack, ", "))
		}

		return nil
	}

	binary, os, compiler, build := toks[0], toks[1], toks[2], toks[3]

	items := []struct {
		value string
		legal []string
	}{
		{binary, []string{"x86_64"}},
		{os, []string{"slc6", "centos7"}},
		{compiler, []string{"gcc49", "gcc62", "gcc8"}},
		{build, []string{"opt", "dbg"}},
	}

	for _, item := range items {
		if err := validate(item.value, item.legal); err != nil {
			return err
		}
	}

	return nil
}

func (a *Args) validateTimestamp() error {
	stamp, err := time.Parse("2006-01-02T1504", a.Timestamp)
	if err != nil {
		return fmt.Errorf(
			"Badly formed timestamp %s, error: %v",
			a.Timestamp,
			err,
		)
	}

	now := time.Now()
	if stamp.After(now) {
		return fmt.Errorf("Date/timestamp %s is in the future", a.Timestamp)
	}

	_30Days := 30 * 24 * 3600.0
	if now.Sub(stamp).Seconds() > _30Days {
		// TODO: is this a valid concern? can we legitimately use older timestamps?
		return fmt.Errorf("Date/timestamp %s is > 30 days in the past", a.Timestamp)
	}

	return nil
}

// ------------------------------------------------------------------

func contains(value string, allowed []string) bool {
	for _, a := range allowed {
		if a == value {
			return true
		}
	}

	return false
}
