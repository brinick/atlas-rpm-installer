package config

import (
	"flag"
	"fmt"
	"strings"
	"time"

	installer "github.com/brinick/atlas-rpm-installer"
)

// Add to this as required...
var legalPlatforms = platforms{
	binary:   []string{"x86_64"},
	os:       []string{"slc6", "centos7"},
	compiler: []string{"gcc49", "gcc62", "gcc8", "clang10"},
	build:    []string{"opt", "dbg"},
}

// ----------------------------------------------------------------------

type platforms struct {
	binary   []string
	os       []string
	compiler []string
	build    []string
}

func (p *platforms) isValid(platform string) bool {
	toks := strings.Split(platform, "-")
	if len(toks) != 4 {
		return false
	}

	return contains(toks[0], p.binary) &&
		contains(toks[1], p.os) &&
		contains(toks[2], p.compiler) &&
		contains(toks[3], p.build)
}

// ----------------------------------------------------------------------

// InstallOpts are options relating to the software to install
type InstallOpts struct {
	installer.Opts
	Release string `json:""`
}

func (i *InstallOpts) String() string {
	return strings.Join(
		[]string{
			"- Install Options:",
			fmt.Sprintf("   - Release: %s", i.Release),
			fmt.Sprintf("   - Project: %s", i.Project),
			fmt.Sprintf("   - Tags file: %s", i.TagsFile),
		},
		"\n",
	)
}

func (i *InstallOpts) flags() {
	flag.StringVar(&i.Release, "release", "", "The release to install")
	flag.StringVar(&i.Project, "project", "", "The project to install")

	flag.StringVar(
		&i.TagsFile,
		"tagsfile",
		"/cvmfs/atlas-nightlies.cern.ch/repo/sw/tags",
		"Location of the tags file",
	)
}

func (i *InstallOpts) validate() error {
	tokens := strings.Split(i.Release, "/")
	if len(tokens) != 3 {
		msg := "-release argument expected of the form:\n"
		msg += "   <branch>/<platform>/<timestamp>\n"
		return fmt.Errorf(msg)
	}

	i.Branch, i.Platform, i.Timestamp = tokens[0], tokens[1], tokens[2]

	if !legalPlatforms.isValid(i.Platform) {
		return fmt.Errorf("%s: illegal platform", i.Platform)
	}

	if err := i.validateTimestamp(); err != nil {
		return err
	}

	project := strings.TrimSpace(i.Project)
	if len(project) == 0 {
		msg := "Please provide a -project option\n"
		return fmt.Errorf(msg)
	}

	// 5 is a fairly safe bet. Ideally we'd get a list of valid projects.
	if len(project) < 5 {
		msg := "Illegal -project argument given ('%s')\n"
		return fmt.Errorf(fmt.Sprintf(msg, project))
	}

	return nil
}

func (i *InstallOpts) validateTimestamp() error {
	loc, _ := time.LoadLocation("Local")
	stamp, err := time.ParseInLocation("2006-01-02T1504", i.Timestamp, loc)
	if err != nil {
		return fmt.Errorf(
			fmt.Sprintf("Badly formed timestamp %s, error: %v\n",
				i.Timestamp,
				err,
			),
		)
	}

	now := time.Now()
	if stamp.After(now) {
		return fmt.Errorf("Date/timestamp %s is in the future", i.Timestamp)
	}

	_30Days := 30 * 24 * 3600.0
	if now.Sub(stamp).Seconds() > _30Days {
		// TODO: is this a valid concern? can we legitimately use older timestamps?
		return fmt.Errorf("Date/timestamp %s is > 30 days in the past", i.Timestamp)
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
