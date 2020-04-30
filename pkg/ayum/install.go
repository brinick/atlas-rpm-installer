package ayum

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/brinick/logging"
	"github.com/brinick/shell"
)

type installer interface {
	Install(context.Context, ...string) error
}

type cmdInstall struct {
	lister
	rpmInstaller   *ayumCommand
	rpmReinstaller *ayumCommand
	log            logging.Logger
}

// Install will install the provided RPMs. Firstly, establish if
// any are already installed, and if so run an ayum reinstall.
// Otherwise just plain install. If any install or reinstall exits
// with non-zero exitcode, stop and return an error.
func (c *cmdInstall) Install(ctx context.Context, rpmsToInstall ...string) error {
	if len(rpmsToInstall) == 0 {
		return nil
	}

	metrics.Count("ayum_nrpms_to_install", len(rpmsToInstall))

	localPackages, err := c.Installed(ctx)
	if err != nil {
		return err
	}

	c.log.Info(
		"Checked for locally installed packages",
		logging.F("nFound", len(*localPackages)),
	)

	metrics.Count("ayum_nlocal_packages", len(*localPackages))

	rpmsNames := removeFileExt(rpmsToInstall...)
	toReinstall, toInstall := localPackages.matching(rpmsNames...)

	// TODO: this could be a metric
	c.log.Info(
		"Categorised RPMs into already/not already installed",
		logging.F("nToReinstall", len(toReinstall)),
		logging.F("nToInstall", len(toInstall)),
	)

	if err := c.reinstallRPMs(ctx, toReinstall...); err != nil {
		return fmt.Errorf("reinstall RPMs failed (%w)", err)
	}

	if err := c.installRPMs(ctx, toInstall...); err != nil {
		return fmt.Errorf("install RPMs failed (%w)", err)
	}

	return nil
}

func (c *cmdInstall) installRPMs(ctx context.Context, rpms ...string) error {
	return c.doInstall(ctx, c.rpmInstaller, rpms)
}

func (c *cmdInstall) reinstallRPMs(ctx context.Context, rpms ...string) error {
	return c.doInstall(ctx, c.rpmReinstaller, rpms)
}

func (c *cmdInstall) doInstall(ctx context.Context, runner *ayumCommand, rpms []string) error {
	if len(rpms) == 0 {
		return nil
	}

	// Format and set the ayum install command with the rpms
	runner.cmd = fmt.Sprintf(runner.cmd, strings.Join(rpms, " "))
	runner.Run(shell.Context(ctx))

	// TODO: error check and log messages
	return runner.Result().Err()
}

// removeFileExt is a helper function to remove the file extension from a list of file names
func removeFileExt(filenames ...string) []string {
	var names []string
	for _, fn := range filenames {
		ext := filepath.Ext(fn)
		name := strings.TrimSuffix(fn, ext)
		names = append(names, name)
	}

	return names
}
