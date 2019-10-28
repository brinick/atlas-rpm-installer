package ayum

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/brinick/logging"
	"github.com/brinick/shell"
)

type lister interface {
	Installed(context.Context) (*localPackages, error)
}

func listCommand(ayumExe string) string {
	return fmt.Sprintf("%s -q list installed", ayumExe)
}

type cmdList struct {
	cmd     []string
	timeout int
	log     logging.Logger
}

// Installed returns the list of locally installed packages.
// If none are found, an empty slice is returned. If an error occurs,
// the package list is nil.
func (c *cmdList) Installed(ctx context.Context) (*localPackages, error) {
	ac := ayumCommand{
		cmd: c.cmd,
		opts: []shell.Option{
			shell.Context(ctx),
			shell.Timeout(time.Duration(c.timeout) * time.Second),
		},
	}

	ac.run()

	if !ac.result.IsError() {
		packages := c.parseInstalled(ac.result.Stdout().Text())
		return packages, nil
	}

	if ac.result.Stdout().Text() == ac.result.Stderr().Text() {
		c.log.Info("No locally installed packages")
		return &localPackages{}, nil
	}

	c.log.Error(
		"Unable to retrieve locally installed package list",
		logging.F("err", ac.result.Error()),
	)

	for _, line := range ac.result.Stderr().Lines() {
		c.log.Error(line)
	}

	for _, line := range ac.result.Stdout().Lines() {
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
	var d map[string]bool
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
