package config

import (
	"flag"
	"fmt"
	"strings"
	"time"

	installer "github.com/brinick/atlas-rpm-installer"
)

type installOpts struct {
	installer.Opts
	Release string
}

func (i *installOpts) flags() {
	flag.StringVar(&i.Release, "release", "", "The release to install")
	flag.StringVar(&i.Project, "project", "", "The project to install")

	flag.StringVar(
		&i.TagsFile,
		"tagsfile",
		"/cvmfs/atlas-nightlies.cern.ch/repo/sw/tags",
		"Location of the tags file",
	)
}

func (i *installOpts) validate() error {
	tokens := strings.Split(i.Release, "/")
	if len(tokens) != 3 {
		msg := "-r argument expected of the form:\n"
		msg += "<branch>/<platform>/<timestamp>"
		return fmt.Errorf(msg)
	}

	i.Branch, i.Platform, i.Timestamp = tokens[0], tokens[1], tokens[2]
	if err := i.validatePlatform(); err != nil {
		return err
	}

	if err := i.validateTimestamp(); err != nil {
		return err
	}

	return nil
}

func (i *installOpts) validatePlatform() error {
	toks := strings.Split(i.Platform, "-")
	if len(toks) != 4 {
		return fmt.Errorf("Badly formed platform '%s'", i.Platform)
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

func (i *installOpts) validateTimestamp() error {
	stamp, err := time.Parse("2006-01-02T1504", i.Timestamp)
	if err != nil {
		return fmt.Errorf(
			"Badly formed timestamp %s, error: %v",
			i.Timestamp,
			err,
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
