package ayum

import (
	"context"
	"fmt"
	"strings"

	"github.com/brinick/logging"
	"github.com/brinick/shell"
)

type lister interface {
	Installed(context.Context) (*localPackages, error)
}

type cmdList struct {
	log logging.Logger
	cmd *ayumCommand
}

// Installed returns the list of locally installed packages.
// If none are found, an empty slice is returned. If an error occurs,
// the package list is nil.
func (c *cmdList) Installed(ctx context.Context) (*localPackages, error) {
	c.cmd.Run(shell.Context(ctx))
	if c.cmd.Result().Err() == nil {
		packages := c.parseInstalled(c.cmd.Result().Stdout().Text())
		return packages, nil
	}

	stdout := c.cmd.Result().Stdout()
	stderr := c.cmd.Result().Stderr()

	if stdout.Text() == stderr.Text() {
		c.log.Info("No locally installed packages")
		return &localPackages{}, nil
	}

	c.log.Error(
		"Unable to retrieve locally installed package list",
		logging.F("err", c.cmd.Result().Err()),
	)

	for _, line := range stdout.Lines() {
		c.log.Info(line)
	}

	for _, line := range stderr.Lines() {
		c.log.Error(line)
	}

	// TODO: report exactly why it failed
	return nil, fmt.Errorf("ayum list installed - command failed")
}

// parseInstalled parses the text returned by the ayum list installed command
func (c *cmdList) parseInstalled(packagesText string) *localPackages {
	var packages localPackages

	lines := strings.Split(packagesText, "installed")
	for _, line := range lines {
		tokens := strings.Fields(line)
		if len(tokens) != 2 {
			c.log.Info(
				"ayum list installed - skipping unexpected line",
				logging.F("l", line),
			)
			continue
		}

		name, version := tokens[0], tokens[1]
		// RPMs are labelled as <name>.noarch so we remove this last part
		name = strings.Replace(name, ".noarch", "", 1)
		packages = append(packages, &localPackage{name, version})
	}

	return &packages
}

// ----------------------------------------------------------------------

// localPackage is a locally installed RPM package
type localPackage struct {
	Name    string
	Version string
}

type localPackages []*localPackage

func (lp *localPackages) matching(rpmNames ...string) ([]string, []string) {
	// Put the names in a map for quick look up
	var d = map[string]bool{}
	for _, rpm := range rpmNames {
		d[rpm] = true
	}

	var installed, notinstalled []string
	for _, p := range *lp {
		pName := p.Name
		if _, found := d[pName]; found {
			installed = append(installed, pName)
			continue
		}

		notinstalled = append(notinstalled, pName)
	}

	return installed, notinstalled
}
